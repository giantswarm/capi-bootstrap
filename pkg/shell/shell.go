package shell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// VerifyBinaryExists verifies that <name> binary exists in the current $PATH
func VerifyBinaryExists(name string) error {
	stdOut, _, err := Execute(Command{
		Name: "which",
		Args: []string{name},
	})
	if err != nil || stdOut == "" {
		return errors.New("not found")
	}
	return nil
}

type Command struct {
	Name string
	Args []string
	Env  map[string]string
	Tee  bool // set this to true to see the output of the command on stdout, useful for debugging
}

func Execute(command Command) (string, string, error) {
	cmd := exec.Command(command.Name, command.Args...)
	for key, value := range command.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	var stdOut strings.Builder
	cmd.Stdout = &stdOut
	var stdErr strings.Builder
	cmd.Stderr = &stdErr
	if command.Tee {
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdOut)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stdErr)
	}
	err := cmd.Run()
	return stdOut.String(), stdErr.String(), err
}
