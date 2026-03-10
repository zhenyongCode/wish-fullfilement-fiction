package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type editFileTool struct {
	workspace string
}

func (t *editFileTool) Description() string {
	//TODO implement me
	return "Edit a file by replacing old_text with new_text. The old_text must exist exactly in the file."
}

func newEditFileTool(workspace string) Tool {
	return &editFileTool{workspace: workspace}
}

func (t *editFileTool) Name() string {
	return "edit_file"
}

func (t *editFileTool) Schema() string {
	return `{"type":"object","properties":{"path":{"type":"string"},"old":{"type":"string"},"new":{"type":"string"}},"required":["path","new"]}`
}

func (t *editFileTool) Run(_ context.Context, input map[string]any) (Result, error) {
	path, _ := input["path"].(string)
	oldValue, _ := input["old"].(string)
	newValue, _ := input["new"].(string)
	resolved, err := resolvePath(t.workspace, path)
	if err != nil {
		return Result{}, err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return Result{}, fmt.Errorf("read file failed: %w", err)
	}
	original := string(data)
	updated := original
	if strings.TrimSpace(oldValue) == "" {
		updated = original + newValue
	} else {
		updated = strings.Replace(original, oldValue, newValue, 1)
		if updated == original {
			return Result{}, errors.New("edit file failed: old content not found")
		}
	}

	if err := os.WriteFile(resolved, []byte(updated), 0o644); err != nil {
		return Result{}, fmt.Errorf("write file failed: %w", err)
	}
	return Result{Output: "ok", ExitCode: 0, Meta: map[string]any{"path": resolved}}, nil
}
