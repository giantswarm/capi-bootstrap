package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ansd/lastpass-go"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/capi-bootstrap/pkg/lastpass/internal"
)

func New() (*Client, error) {
	_, err := exec.LookPath("lpass")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return &Client{}, nil
}

func (c *Client) authenticate(ctx context.Context) error {
	command := exec.CommandContext(ctx, "lpass", "ls")
	command.Env = append([]string{"LPASS_DISABLE_PINENTRY=1"}, buildEnv()...)

	var stdErr bytes.Buffer
	command.Stderr = &stdErr

	err := command.Run()
	if err != nil {
		return microerror.Maskf(commandFailedError, stdErr.String())
	}

	return nil
}

func (c *Client) Create(ctx context.Context, share, group, name, notes string) (*lastpass.Account, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, microerror.Mask(err)
	}

	args := []string{
		"add",
		"--notes", // type of data being provided via stdin
		"--non-interactive",
		"--sync=now",
		filepath.Join(share, group, name),
	}
	command := exec.CommandContext(ctx, "lpass", args...)
	command.Env = buildEnv()

	var stdErr bytes.Buffer
	command.Stderr = &stdErr

	stdIn, err := command.StdinPipe()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = stdIn.Write([]byte(notes))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = command.Start()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = stdIn.Close()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = command.Wait()
	if err != nil {
		return nil, microerror.Maskf(commandFailedError, stdErr.String())
	}

	err = c.sync(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	created, err := c.Get(ctx, share, group, name)
	return created, microerror.Mask(err)
}

func (c *Client) Get(ctx context.Context, share, group, name string) (*lastpass.Account, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, microerror.Mask(err)
	}

	command := exec.CommandContext(ctx, "lpass",
		"show",
		"--sync=now",
		"--expand-multi", // return multiple results if found instead of showing a warning message
		filepath.Join(share, group, name),
		"--json")
	command.Env = buildEnv()

	var stdOut bytes.Buffer
	command.Stdout = &stdOut
	var stdErr bytes.Buffer
	command.Stderr = &stdErr

	err := command.Run()
	if err != nil {
		if strings.Contains(stdErr.String(), "Could not find specified account(s).") {
			return nil, microerror.Maskf(internal.NotFoundError, "account %s not found", filepath.Join(share, group, name))
		}
		return nil, microerror.Mask(err)
	}

	var secrets []jsonSecret
	err = json.Unmarshal(stdOut.Bytes(), &secrets)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(secrets) > 1 {
		return nil, microerror.Maskf(notUniqueError, "found %d secrets matching name %s", len(secrets), filepath.Join(share, group, name))
	}

	secret := secrets[0]

	return &lastpass.Account{
		ID:              secret.ID,
		Name:            secret.Name,
		Username:        secret.Username,
		Password:        secret.Password,
		URL:             secret.URL,
		Group:           secret.Group,
		Share:           secret.Share,
		Notes:           secret.Notes,
		LastModifiedGMT: secret.LastModifiedGMT,
		LastTouch:       secret.LastTouch,
	}, nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	if err := c.authenticate(ctx); err != nil {
		return microerror.Mask(err)
	}

	command := exec.CommandContext(ctx, "lpass", "rm", "--sync=now", id)
	command.Env = buildEnv()

	err := command.Run()
	return microerror.Mask(err)
}

// buildEnv returns environment variables needed by lastpass-cli to find exising lastpass agent
func buildEnv() []string {
	return []string{
		fmt.Sprintf("XDG_RUNTIME_DIR=%s", os.Getenv("XDG_RUNTIME_DIR")),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	}
}

func (c *Client) sync(ctx context.Context) error {
	if err := c.authenticate(ctx); err != nil {
		return microerror.Mask(err)
	}

	command := exec.CommandContext(ctx, "lpass", "sync")
	command.Env = buildEnv()

	err := command.Run()
	return microerror.Mask(err)
}
