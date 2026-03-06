package tools

import (
	"context"
	"fmt"
	"os"
)

type readFileTool struct {
	workspace string
}

func (t *readFileTool) Description() string {
	//TODO implement me
	return "Read file from workspace"
}

func newReadFileTool(workspace string) Tool {
	return &readFileTool{workspace: workspace}
}

func (t *readFileTool) Name() string {
	return "read_file"
}

func (t *readFileTool) Schema() string {
	return `{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`
}

func (t *readFileTool) Run(_ context.Context, input map[string]any) (Result, error) {
	path, _ := input["path"].(string)
	resolved, err := resolvePath(t.workspace, path)
	if err != nil {
		return Result{}, err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return Result{}, fmt.Errorf("read file failed: %w", err)
	}
	return Result{Output: string(data), ExitCode: 0, Meta: map[string]any{"path": resolved}}, nil
}
