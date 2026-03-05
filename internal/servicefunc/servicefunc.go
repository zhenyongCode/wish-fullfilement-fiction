// =================================================================================
// Service Function Registry and Executor
// Supports dynamic registration and execution of service functions and methods
// =================================================================================

package servicefunc

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// ServiceFunc defines the standard service function signature
// ctx: request context
// params: input parameters as map
// Returns: result map and error
type ServiceFunc func(ctx context.Context, params g.Map) (g.Map, error)

// registry holds all registered service functions
var (
	registry = make(map[string]ServiceFunc)
	mu       sync.RWMutex
	logger   *glog.Logger
)

func init() {
	logger = g.Log()
}

// RegisterFunc registers a service function by name
// Panics if function with same name already exists
func RegisterFunc(name string, fn ServiceFunc) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("servicefunc: function '%s' already registered", name))
	}

	registry[name] = fn
	logger.Debugf(context.Background(), "servicefunc: registered function '%s'", name)
}

// RegisterMethod registers a method from a receiver struct
// The method must have signature: func(ctx context.Context, params g.Map) (g.Map, error)
// Panics if method signature is invalid or function with same name already exists
func RegisterMethod(name string, receiver interface{}, methodName string) {
	mu.Lock()
	defer mu.Unlock()

	// Validate receiver is not nil
	if receiver == nil {
		panic(fmt.Sprintf("servicefunc: receiver for method '%s' is nil", name))
	}

	// Get method by name
	val := reflect.ValueOf(receiver)
	method := val.MethodByName(methodName)

	if !method.IsValid() {
		panic(fmt.Sprintf("servicefunc: method '%s' not found on receiver", methodName))
	}

	// Validate method signature
	methodType := method.Type()
	if methodType.NumIn() != 2 || methodType.NumOut() != 2 {
		panic(fmt.Sprintf("servicefunc: method '%s' must have signature func(context.Context, g.Map) (g.Map, error)", methodName))
	}

	// Check input types
	if methodType.In(0).String() != "context.Context" {
		panic(fmt.Sprintf("servicefunc: method '%s' first parameter must be context.Context", methodName))
	}
	if methodType.In(1).String() != "g.Map" && methodType.In(1).String() != "map[string]interface {}" {
		panic(fmt.Sprintf("servicefunc: method '%s' second parameter must be g.Map", methodName))
	}

	// Check output types
	if methodType.Out(0).String() != "g.Map" && methodType.Out(0).String() != "map[string]interface {}" {
		panic(fmt.Sprintf("servicefunc: method '%s' first return must be g.Map", methodName))
	}
	if methodType.Out(1).String() != "error" {
		panic(fmt.Sprintf("servicefunc: method '%s' second return must be error", methodName))
	}

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("servicefunc: function '%s' already registered", name))
	}

	// Wrap the method as ServiceFunc
	registry[name] = func(ctx context.Context, params g.Map) (g.Map, error) {
		results := method.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(params),
		})

		var result g.Map
		var err error

		if !results[0].IsNil() {
			result = results[0].Interface().(g.Map)
		}

		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}

		return result, err
	}

	logger.Debugf(context.Background(), "servicefunc: registered method '%s' -> %s.%s", name, reflect.TypeOf(receiver).Name(), methodName)
}

// ServiceFuncExe executes a registered service function by name
// Returns error if function not found or execution fails
func ServiceFuncExe(ctx context.Context, name string, params g.Map) (g.Map, error) {
	mu.RLock()
	fn, exists := registry[name]
	mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("servicefunc: function '%s' not found", name)
	}

	logger.Infof(ctx, "servicefunc: executing '%s' with params: %v", name, params)

	result, err := fn(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "servicefunc: execution of '%s' failed: %v", name, err)
		return nil, err
	}

	logger.Infof(ctx, "servicefunc: execution of '%s' completed, result: %v", name, result)
	return result, nil
}

// ListRegistered returns all registered function names
func ListRegistered() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// IsRegistered checks if a function is registered
func IsRegistered(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, exists := registry[name]
	return exists
}