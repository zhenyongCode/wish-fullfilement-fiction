package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"wish-fullfilement-fiction/internal/consts"
)

type Result struct {
	Output   string
	ExitCode int
	Meta     map[string]any
}

type Tool interface {
	Name() string
	Schema() string
	Description() string
	Run(ctx context.Context, input map[string]any) (Result, error)
}

type Registry interface {
	Register(tool Tool) error
	Get(name string) (Tool, bool)
	List() []Tool
}

type Executor interface {
	Execute(ctx context.Context, name string, input map[string]any, timeout time.Duration) (Result, error)
}

func resolvePath(workspace, path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", fmt.Errorf("%w: path is required", ErrInvalidInput)
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	base := strings.TrimSpace(workspace)
	if base == "" {
		base = consts.DefaultWorkspace
	}
	return filepath.Clean(filepath.Join(base, p)), nil
}
