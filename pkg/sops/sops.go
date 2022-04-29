package sops

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"filippo.io/age"
	"github.com/giantswarm/microerror"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/key"
	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
)

func New(config Config) (*Client, error) {
	return &Client{
		lastpassClient: config.LastpassClient,
		clusterName:    config.ClusterName,
	}, nil
}

func (c *Client) EnsureEncryptionKey(ctx context.Context) (*EncryptionKey, error) {
	err := c.LoadPrivateKey(ctx)
	if lastpass.IsNotFound(err) {
		// fall through
	} else if err != nil {
		return nil, microerror.Mask(err)
	} else {
		return c.encryptionKey, nil
	}

	c.encryptionKey, err = generateEncryptionKey()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	share := key.EncryptionKeySecretShare
	group := key.EncryptionKeySecretGroup
	name := key.EncryptionKeySecretName(c.clusterName)

	_, err = c.lastpassClient.Create(ctx, share, group, name, c.encryptionKey.PrivateKey)
	return c.encryptionKey, microerror.Mask(err)
}

func (c *Client) DeleteEncryptionKey(ctx context.Context) error {
	share := key.EncryptionKeySecretShare
	group := key.EncryptionKeySecretGroup
	name := key.EncryptionKeySecretName(c.clusterName)

	account, err := c.lastpassClient.Get(ctx, share, group, name)
	if lastpass.IsNotFound(err) {
		return nil // already deleted, nothing to do
	} else if err != nil {
		return microerror.Mask(err)
	}

	err = c.lastpassClient.Delete(ctx, account.ID)
	return microerror.Mask(err)
}

func (c *Client) EncryptSecret(ctx context.Context, secret *core.Secret) ([]byte, error) {
	secretYAML, err := yaml.Marshal(secret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	encrypted, err := c.encryptYAML(ctx, secretYAML, "^(data|stringData)$")
	return encrypted, microerror.Mask(err)
}

func (c *Client) encryptYAML(ctx context.Context, plaintext []byte, encryptedRegex string) ([]byte, error) {
	err := ensureSopsMetadata(plaintext, false)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = c.LoadPublicKey(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	args := []string{
		"--encrypt",
		"--age",
		c.encryptionKey.PublicKey,
		"--input-type",
		"yaml",
		"--output-type",
		"yaml",
		"/dev/stdin",
	}
	if encryptedRegex != "" {
		args = append([]string{"--encrypted-regex", encryptedRegex}, args...)
	}

	sopsCommand := exec.CommandContext(ctx, "sops", args...)

	stdIn, err := sopsCommand.StdinPipe()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var sopsOut bytes.Buffer
	sopsCommand.Stdout = &sopsOut
	var sopsErr bytes.Buffer
	sopsCommand.Stderr = &sopsErr

	err = sopsCommand.Start()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = stdIn.Write(plaintext)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = stdIn.Close()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = sopsCommand.Wait()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return sopsOut.Bytes(), nil
}

func (c *Client) EncryptYAML(ctx context.Context, data []byte) ([]byte, error) {
	encrypted, err := c.encryptYAML(ctx, data, "")
	return encrypted, microerror.Mask(err)
}

func (c *Client) DecryptSecret(ctx context.Context, encryptedYAML []byte) (*core.Secret, error) {
	decryptedYAML, err := c.decryptYAML(ctx, encryptedYAML)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secret core.Secret
	err = yaml.Unmarshal(decryptedYAML, &secret)
	return &secret, microerror.Mask(err)
}

func (c *Client) decryptYAML(ctx context.Context, encrypted []byte) ([]byte, error) {
	err := ensureSopsMetadata(encrypted, true)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = c.LoadPrivateKey(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	args := []string{
		"--decrypt",
		"--input-type",
		"yaml",
		"--output-type",
		"yaml",
		"/dev/stdin",
	}

	sopsCommand := exec.CommandContext(ctx, "sops", args...)
	sopsCommand.Env = []string{
		fmt.Sprintf("SOPS_AGE_KEY=%s", c.encryptionKey.PrivateKey),
	}
	sopsCommand.Stdin = pipeRead
	var sopsOut bytes.Buffer
	sopsCommand.Stdout = &sopsOut
	var sopsErr bytes.Buffer
	sopsCommand.Stderr = &sopsErr

	_, err = pipeWrite.Write(encrypted)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = sopsCommand.Start()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = pipeWrite.Close()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = sopsCommand.Wait()
	if err != nil {
		return nil, microerror.Maskf(commandFailedError, sopsErr.String())
	}

	return sopsOut.Bytes(), nil
}

func (c *Client) RenderConfig(ctx context.Context) ([]byte, error) {
	err := c.LoadPublicKey(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	creationRule := CreationRule{
		PathRegex:      ".*.yaml",
		EncryptedRegex: "^(data|stringData)$",
	}

	switch c.encryptionKey.Type {
	case "age":
		creationRule.Age = c.encryptionKey.PublicKey
	default:
		return nil, microerror.Maskf(invalidConfigError, "unsupported encryption type %s", c.encryptionKey.Type)
	}

	content, err := yaml.Marshal(SopsConfig{
		CreationRules: []CreationRule{
			creationRule,
		},
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return content, nil
}

func (c *Client) LoadPublicKey(ctx context.Context) error {
	if c.encryptionKey == nil {
		var publicKey string
		if value, ok := os.LookupEnv("SOPS_AGE_RECIPIENTS"); ok {
			publicKey = value
		} else if publicKey == "" {
			if c.lastpassClient == nil {
				return microerror.Maskf(invalidConfigError, "SOPS_AGE_RECIPIENTS environment variable not defined and lastpass client not initialized")
			}

			share := key.EncryptionKeySecretShare
			group := key.EncryptionKeySecretGroup
			name := key.EncryptionKeySecretName(c.clusterName)

			var err error
			encryptionKeyAccount, err := c.lastpassClient.Get(ctx, share, group, name)
			if err != nil {
				return microerror.Mask(err)
			}

			identity, err := age.ParseX25519Identity(encryptionKeyAccount.Notes)
			if err != nil {
				return microerror.Mask(err)
			}

			publicKey = identity.Recipient().String()
		}

		recipient, err := age.ParseX25519Recipient(publicKey)
		if err != nil {
			return microerror.Mask(err)
		}

		c.encryptionKey = &EncryptionKey{
			PublicKey: recipient.String(),
			Type:      "age",
		}
	}

	return nil
}

func (c *Client) LoadPrivateKey(ctx context.Context) error {
	if c.encryptionKey == nil || c.encryptionKey.PrivateKey == "" {
		var privateKey string
		if value, ok := os.LookupEnv("SOPS_AGE_KEY_FILE"); ok {
			contents, err := os.ReadFile(value)
			if err != nil {
				return microerror.Mask(err)
			}
			privateKey = string(contents)
		} else if value, ok := os.LookupEnv("SOPS_AGE_KEY"); ok {
			privateKey = value
		} else {
			if c.lastpassClient == nil {
				return microerror.Maskf(invalidConfigError, "SOPS_AGE_KEY_FILE/SOPS_AGE_KEY environment variable not defined and lastpass client not initialized")
			}

			share := key.EncryptionKeySecretShare
			group := key.EncryptionKeySecretGroup
			name := key.EncryptionKeySecretName(c.clusterName)

			var err error
			encryptionKeyAccount, err := c.lastpassClient.Get(ctx, share, group, name)
			if err != nil {
				return microerror.Mask(err)
			}

			privateKey = encryptionKeyAccount.Notes
		}

		identity, err := age.ParseX25519Identity(privateKey)
		if err != nil {
			return microerror.Mask(err)
		}

		c.encryptionKey = &EncryptionKey{
			PrivateKey: identity.String(),
			PublicKey:  identity.Recipient().String(),
			Type:       "age",
		}
	}

	return nil
}

func ensureSopsMetadata(content []byte, exists bool) error {
	var mapData map[string]interface{}
	err := yaml.Unmarshal(content, &mapData)
	if err != nil {
		return microerror.Mask(err)
	}

	if _, ok := mapData["sops"]; exists && !ok {
		return microerror.Maskf(invalidConfigError, "input file already encrypted")
	} else if ok && !exists {
		return microerror.Maskf(invalidConfigError, "input file not encrypted")
	}

	return nil
}

func generateEncryptionKey() (*EncryptionKey, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &EncryptionKey{
		PrivateKey: identity.String(),
		PublicKey:  identity.Recipient().String(),
		Type:       "age",
	}, nil
}
