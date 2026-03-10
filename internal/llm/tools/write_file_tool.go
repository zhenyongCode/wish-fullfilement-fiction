package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type writeFileTool struct {
	workspace string
}

func (t *writeFileTool) Description() string {
	//TODO implement me
	return "Write content to a file at the given path. Creates parent directories if needed."
}

func newWriteFileTool(workspace string) Tool {
	return &writeFileTool{workspace: workspace}
}

func (t *writeFileTool) Name() string {
	return "write_file"
}

func (t *writeFileTool) Schema() string {
	return `{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}`
}

func (t *writeFileTool) Run(_ context.Context, input map[string]any) (Result, error) {
	path, _ := input["path"].(string)
	content, _ := input["content"].(string)
	resolved, err := resolvePath(t.workspace, path)
	if err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return Result{}, fmt.Errorf("create parent dir failed: %w", err)
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return Result{}, fmt.Errorf("write file failed: %w", err)
	}
	return Result{Output: "ok", ExitCode: 0, Meta: map[string]any{"path": resolved}}, nil
}
