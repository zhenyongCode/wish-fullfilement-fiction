package hello

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"

	"wish-fullfilement-fiction/api/hello/v1"
	"wish-fullfilement-fiction/internal/servicefunc"
)

func (c *ControllerV1) Hello(ctx context.Context, req *v1.HelloReq) (res *v1.HelloRes, err error) {
	g.Log().Info(ctx, "HelloReq", req)

	// Default response if no func specified
	if req.Data == nil || len(req.Data) == 0 || req.Data["func"] == nil {
		g.RequestFromCtx(ctx).Response.WriteExit("Hello, World!")
		return
	}

	// Extract function name
	funcName, ok := req.Data["func"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid 'func' parameter: expected string")
	}

	// Extract params (optional)
	var params g.Map
	if req.Data["param"] != nil {
		params, ok = req.Data["param"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid 'param' parameter: expected map[string]interface{}")
		}
	}

	// Execute service function
	exeRes, err := servicefunc.ServiceFuncExe(ctx, funcName, params)
	if err != nil {
		return nil, err
	}

	res = &v1.HelloRes{
		g.Meta{},
		exeRes,
	}
	return res, nil
}
