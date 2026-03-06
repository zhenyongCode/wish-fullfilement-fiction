// =================================================================================
// Example Service Functions
// Demonstrates how to register service functions and methods
// =================================================================================

package logic

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"wish-fullfilement-fiction/internal/service"
	"wish-fullfilement-fiction/internal/servicefunc"
)

var chatService *service.ChatService

func init() {
	// Register function-style service
	servicefunc.RegisterFunc("echo", echoFunc)

	// Register method-style service
	receiver := &ExampleService{}
	servicefunc.RegisterMethod("hello", receiver, "Hello")
	servicefunc.RegisterMethod("add", receiver, "Add")

	// Register chat service (lazy initialization)
	// Note: ChatService will be initialized on first call to ensure config is loaded
	servicefunc.RegisterFunc("chat", chatExec)
}

// chatExec wraps ChatService.Exec for lazy initialization
func chatExec(ctx context.Context, params g.Map) (g.Map, error) {
	// Initialize chat service on first call
	if chatService == nil {
		var err error
		chatService, err = service.NewChatService(ctx)
		if err != nil {
			return nil, err
		}
	}
	return chatService.Exec(ctx, params)
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
