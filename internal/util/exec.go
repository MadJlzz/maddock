package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Commander interface {
	Run(ctx context.Context, name string, args []string) (io.Reader, io.Reader, int, error)
}

type RealCommander struct{}

func (r RealCommander) Run(ctx context.Context, name string, args []string) (io.Reader, io.Reader, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	return stdout, stderr, cmd.ProcessState.ExitCode(), err
}

type MockCommander struct {
	Commands map[string]MockCommand
}

type MockCommand struct {
	Output   string
	ExitCode int
}

func (c MockCommander) Run(ctx context.Context, name string, args []string) (io.Reader, io.Reader, int, error) {
	joinedArgs := strings.Join(args, "-")
	key := fmt.Sprintf("%s-%s", name, joinedArgs)

	cmd, ok := c.Commands[key]
	if !ok {
		return nil, nil, -1, fmt.Errorf("unknown command: %s", key)
	}
	return strings.NewReader(cmd.Output), strings.NewReader(""), cmd.ExitCode, nil
}
