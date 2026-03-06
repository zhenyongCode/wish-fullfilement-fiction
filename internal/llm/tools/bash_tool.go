package tools

import (
	"context"
	"fmt"
	"strings"
	"wish-fullfilement-fiction/internal/consts"
)

type commandRunner func(ctx context.Context, command string) (string, int, error)
type bashTool struct {
	runner commandRunner
}

func (t *bashTool) Description() string {
	return "Run shell command"
}

func newBashTool(runner commandRunner) Tool {
	return &bashTool{runner: runner}
}

func (t *bashTool) Name() string {
	return consts.ToolNameBash
}

func (t *bashTool) Schema() string {
	return `{"type":"object","properties":{"command":{"type":"string"}},"required":["command"]}`
}

func (t *bashTool) Run(ctx context.Context, input map[string]any) (Result, error) {
	command, _ := input["command"].(string)
	command = strings.TrimSpace(command)
	if command == "" {
		return Result{}, fmt.Errorf("%w: bash.command is required", ErrInvalidInput)
	}
	output, exitCode, err := t.runner(ctx, command)
	if err != nil {
		return Result{Output: output, ExitCode: exitCode, Meta: map[string]any{"command": command}}, nil
	}
	return Result{Output: output, ExitCode: exitCode, Meta: map[string]any{"command": command}}, nil
}
