package lastpass

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/giantswarm/capi-bootstrap/pkg/shell"
)

func New() (Client, error) {
	return Client{
		authenticated: false,
	}, nil
}

func (c *Client) Login(credentials Credentials) error {
	if c.authenticated {
		return errors.New("already authenticated")
	}

	totpToken, err := generateTOTP(credentials.TOTPSecret)
	if err != nil {
		return err
	}

	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		return err
	}

	// echo <PASSWORD>\n<OTP> | LPASS_DISABLE_PINENTRY=1 lpass login --trust <USERNAME>
	echoCmd := exec.Command("echo", fmt.Sprintf("%s\n%s", credentials.Password, totpToken))

	loginCmd := exec.Command("lpass", "login", "--trust", "--force", credentials.Username)
	loginCmd.Env = []string{
		"LPASS_DISABLE_PINENTRY=1",
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	}

	echoCmd.Stdout = pipeWrite
	loginCmd.Stdin = pipeRead

	var out strings.Builder
	loginCmd.Stdout = &out

	var errOut strings.Builder
	loginCmd.Stderr = &errOut

	err = echoCmd.Start()
	if err != nil {
		return err
	}

	err = loginCmd.Start()
	if err != nil {
		return err
	}

	err = echoCmd.Wait()
	if err != nil {
		return err
	}

	err = pipeWrite.Close()
	if err != nil {
		return err
	}

	err = loginCmd.Wait()
	if err != nil {
		return err
	}

	if out.String() == "" {
		return errors.New("no output")
	}

	c.authenticated = true

	return nil
}

// Logout using lastpass-cli
func (c *Client) Logout() error {
	if !c.authenticated {
		return errors.New("not authenticated")
	}

	// lpass logout --force
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "lpass",
		Args: []string{"logout", "--force"},
	})
	if err != nil {
		return fmt.Errorf("%q: %s", err, stdErr)
	} else if stdOut == "" {
		return errors.New("no output")
	}

	c.authenticated = false

	return nil
}

func (c *Client) GetSecret(group string, name string) (Secret, error) {
	secrets, err := c.GetSecrets(group, name)
	if err != nil {
		return Secret{}, err
	}

	if len(secrets) != 1 {
		return Secret{}, errors.New("found multiple secrets")
	}

	return secrets[0], nil
}

func (c *Client) GetSecrets(group string, name string) ([]Secret, error) {
	if !c.authenticated {
		return nil, errors.New("not authenticated")
	}

	// lpass show <GROUP>/<NAME> --json --expand-multi
	stdOut, stdErr, err := shell.Execute(shell.Command{
		Name: "lpass",
		Args: []string{
			"show",
			buildFullName(group, name),
			"--json",
			"--expand-multi",
		},
		Env: map[string]string{
			"HOME": os.Getenv("HOME"),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%q: %s", err, stdErr)
	} else if stdOut == "" {
		return nil, errors.New("no output")
	}

	var secrets []Secret
	err = json.Unmarshal([]byte(stdOut), &secrets)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// returns <GROUP>/<NAME> or <NAME>
func buildFullName(group string, name string) string {
	var b bytes.Buffer
	if group != "" {
		b.WriteString(group)
		b.WriteString("/")
		b.WriteString(name)
	} else {
		b.WriteString(name)
	}
	return b.String()
}
