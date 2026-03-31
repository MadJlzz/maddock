package util

import (
	"bytes"
	"context"
	"errors"
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
	bin, err := exec.LookPath(name)
	if err != nil {
		return nil, nil, -1, fmt.Errorf("could not find %q: %w", name, err)
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		err = nil
	}
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
