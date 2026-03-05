// =================================================================================
// Example Service Functions
// Demonstrates how to register service functions and methods
// =================================================================================

package logic

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"wish-fullfilement-fiction/internal/servicefunc"
)

func init() {
	// Register function-style service
	servicefunc.RegisterFunc("echo", echoFunc)

	// Register method-style service
	receiver := &ExampleService{}
	servicefunc.RegisterMethod("hello", receiver, "Hello")
	servicefunc.RegisterMethod("add", receiver, "Add")
}

// echoFunc is a simple echo function
func echoFunc(ctx context.Context, params g.Map) (g.Map, error) {
	return g.Map{
		"echo":  params,
		"count": len(params),
	}, nil
}

// ExampleService demonstrates method-style service registration
type ExampleService struct{}

// Hello returns a greeting
func (s *ExampleService) Hello(ctx context.Context, params g.Map) (g.Map, error) {
	name := "World"
	if n, ok := params["name"].(string); ok && n != "" {
		name = n
	}
	return g.Map{
		"message": "Hello, " + name + "!",
	}, nil
}

// Add performs addition
func (s *ExampleService) Add(ctx context.Context, params g.Map) (g.Map, error) {
	a, _ := params["a"].(float64)
	b, _ := params["b"].(float64)
	return g.Map{
		"result": a + b,
		"a":      a,
		"b":      b,
	}, nil
}
