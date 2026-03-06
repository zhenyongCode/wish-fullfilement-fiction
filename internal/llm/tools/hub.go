package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
	"wish-fullfilement-fiction/internal/consts"
	"wish-fullfilement-fiction/internal/llm"
)

var (
	ErrToolExists      = errors.New("tool already registered")
	ErrToolNotFound    = errors.New("tool not found")
	ErrInvalidToolName = errors.New("invalid tool name")
	ErrInvalidInput    = errors.New("invalid tool input")
)

type Hub struct {
	mu             sync.RWMutex
	registry       map[string]Tool
	workspace      string
	defaultTimeout time.Duration
	Tools          []llm.ToolSpec
	// 默认runner，主要用于bash工具，如果用户没有设置runner，则bash工具不可用
	runner commandRunner
}

func NewHub(defaultTimeout time.Duration) *Hub {
	if defaultTimeout <= 0 {
		defaultTimeout = consts.DefaultToolTimeout
	}
	workspace := consts.DefaultWorkspace
	return &Hub{
		registry:       map[string]Tool{},
		workspace:      strings.TrimSpace(workspace),
		defaultTimeout: defaultTimeout,
		runner:         defaultCommandRunner,
	}
}
func (h *Hub) GetTools() []llm.ToolSpec {
	return h.Tools
}
func defaultCommandRunner(ctx context.Context, command string) (string, int, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if output == "" {
		output = "(no output)"
	}
	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		return output, exitCode, err
	}
	return output, exitCode, nil
}

func (h *Hub) SetCommandRunner(runner commandRunner) {
	if runner == nil {
		return
	}
	h.runner = runner
	h.mu.Lock()
	h.registry[consts.ToolNameBash] = newBashTool(runner)
	h.mu.Unlock()
}

func (h *Hub) Register(tool Tool) error {
	name := strings.TrimSpace(tool.Name())
	if name == "" {
		return ErrInvalidToolName
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.registry[name]; exists {
		if name != consts.ToolNameBash {
			return fmt.Errorf("%w: %s", ErrToolExists, name)
		}
	}
	h.registry[name] = tool
	h.Tools = append(h.Tools, llm.ToolSpec{
		Name:        tool.Name(),
		Description: tool.Description(),
		JSONSchema:  tool.Schema(),
	})
	return nil
}

func (h *Hub) Get(name string) (Tool, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	tool, ok := h.registry[strings.TrimSpace(name)]
	return tool, ok
}

func (h *Hub) List() []Tool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	tools := make([]Tool, 0, len(h.registry))
	for _, tool := range h.registry {
		tools = append(tools, tool)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name() < tools[j].Name()
	})
	return tools
}

func (h *Hub) Execute(ctx context.Context, name string, input map[string]any, timeout time.Duration) (Result, error) {
	tool, ok := h.Get(name)
	if !ok {
		return Result{}, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	effectiveTimeout := timeout
	if effectiveTimeout <= 0 {
		effectiveTimeout = h.defaultTimeout
	}
	execCtx, cancel := context.WithTimeout(ctx, effectiveTimeout)
	defer cancel()

	start := time.Now()

	res, err := tool.Run(execCtx, input)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		g.Log().Debugf(ctx, "tool execution failed %v %d", err, duration) //logging.Field{Key: "module", Value: "tools"},
		//logging.Field{Key: "operation", Value: "tools.exec"},
		//logging.Field{Key: "tool_name", Value: tool.Name()},
		//logging.Field{Key: "duration_ms", Value: duration},

		return Result{}, err
	}

	g.Log().Debugf(ctx, "tool execution end") //logging.Field{Key: "module", Value: "tools"},
	//logging.Field{Key: "operation", Value: "tools.exec"},
	//logging.Field{Key: "tool_name", Value: tool.Name()},
	//logging.Field{Key: "duration_ms", Value: duration},
	//logging.Field{Key: "output_size", Value: len(res.Output)},
	//logging.Field{Key: "exit_code", Value: res.ExitCode},

	return res, nil
}
