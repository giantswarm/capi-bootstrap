package sops

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"filippo.io/age"
	"github.com/giantswarm/microerror"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass"
)

func New(config Config) (*Client, error) {
	if config.ClusterName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterName must not be empty", config)
	}
	if config.LastpassClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.LastpassClient must not be empty", config)
	}
	return &Client{
		lastpassClient: config.LastpassClient,
		clusterName:    config.ClusterName,
	}, nil
}

func (c *Client) EnsureEncryptionKey(ctx context.Context) (*EncryptionKey, error) {
	_, err := c.loadEncryptionKey(ctx)
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

	_, err = c.lastpassClient.CreateAccount(ctx, "Shared-Team Rocket", "Encryption Keys", c.clusterName, c.encryptionKey.PrivateKey)
	return c.encryptionKey, microerror.Mask(err)
}

func (c *Client) DeleteEncryptionKey(ctx context.Context) error {
	account, err := c.lastpassClient.GetAccount(ctx, "Shared-Team Rocket", "Encryption Keys", c.clusterName)
	if lastpass.IsNotFound(err) {
		return nil // already deleted, nothing to do
	} else if err != nil {
		return microerror.Mask(err)
	}

	err = c.lastpassClient.DeleteAccount(ctx, account.ID)
	return microerror.Mask(err)
}

func (c *Client) EncryptSecret(secret *core.Secret) ([]byte, error) {
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	secretContent, err := yaml.Marshal(secret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = pipeWrite.Write(secretContent)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	sopsCommand := exec.Command("sops", "--encrypted-regex", "^(data|stringData)$", "--encrypt", "--age", c.encryptionKey.PublicKey, "--input-type", "yaml", "--output-type", "yaml", "/dev/stdin")
	sopsCommand.Stdin = pipeRead
	var sopsOut bytes.Buffer
	sopsCommand.Stdout = &sopsOut
	var sopsErr bytes.Buffer
	sopsCommand.Stderr = &sopsErr

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
		return nil, microerror.Mask(err)
	}

	return sopsOut.Bytes(), nil
}

func (c *Client) DecryptSecret(encrypted []byte) (*core.Secret, error) {
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	err = os.WriteFile("/tmp/age.key", []byte(c.encryptionKey.PrivateKey), 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove("/tmp/age.key")

	sopsCommand := exec.Command("sops", "--decrypt", "--input-type", "yaml", "--output-type", "yaml", "/dev/stdin")
	sopsCommand.Env = []string{
		"SOPS_AGE_KEY_FILE=/tmp/age.key",
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
		return nil, microerror.Mask(err)
	}

	var decryptedSecret core.Secret
	err = yaml.Unmarshal(sopsOut.Bytes(), &decryptedSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return &decryptedSecret, nil
}

func (c *Client) RenderConfig() (string, error) {
	creationRule := CreationRule{
		PathRegex:      ".*.yaml",
		EncryptedRegex: "^(data|stringData)$",
	}

	switch c.encryptionKey.Type {
	case "age":
		creationRule.Age = c.encryptionKey.PublicKey
	default:
		return "", microerror.Maskf(invalidConfigError, "unsupported encryption type %s", c.encryptionKey.Type)
	}

	content, err := yaml.Marshal(SopsConfig{
		CreationRules: []CreationRule{
			creationRule,
		},
	})
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(content), nil
}

func (c *Client) loadEncryptionKey(ctx context.Context) (*EncryptionKey, error) {
	if c.encryptionKey == nil {
		encryptionKeyAccount, err := c.lastpassClient.GetAccount(ctx, "Shared-Team Rocket", "Encryption Keys", c.clusterName)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		identity, err := age.ParseX25519Identity(encryptionKeyAccount.Notes)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c.encryptionKey = &EncryptionKey{
			PrivateKey: identity.String(),
			PublicKey:  identity.Recipient().String(),
			Type:       "age",
		}
	}

	return c.encryptionKey, nil
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
