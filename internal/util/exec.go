package util

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Commander interface {
	Run(ctx context.Context, name string, args []string) (string, string, int, error)
}

type RealCommander struct{}

func (r RealCommander) Run(ctx context.Context, name string, args []string) (string, string, int, error) {
	bin, err := exec.LookPath(name)
	if err != nil {
		return "", "", -1, fmt.Errorf("could not find %q: %w", name, err)
	}
	bufOut, bufErr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

	err = cmd.Run()
	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		err = nil
	}

	return bufOut.String(), bufErr.String(), cmd.ProcessState.ExitCode(), err
}

type MockCommander struct {
	Commands map[string]MockCommand
}

type MockCommand struct {
	Output   string
	ExitCode int
}

func (c MockCommander) Run(ctx context.Context, name string, args []string) (string, string, int, error) {
	joinedArgs := strings.Join(args, " ")
	key := fmt.Sprintf("%s %s", name, joinedArgs)

	cmd, ok := c.Commands[key]
	if !ok {
		return "", "", -1, fmt.Errorf("unknown command: %s", key)
	}
	return cmd.Output, "", cmd.ExitCode, nil
}
