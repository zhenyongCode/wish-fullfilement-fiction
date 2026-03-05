package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
	"net/http"
)

type BaseResponse struct {
	Status    int         `json:"status"`
	Message   string      `json:"message"`
	RequestId string      `json:"requestId"`
	Data      interface{} `json:"data"`
}

// HandlerResponse is the default middleware handling handler response object and its error.
func HandlerResponse(r *ghttp.Request) {
	glog.Infof(r.Context(), "Start Req: %s %s , body: %s query: %v", r.Method, r.URL.Path, r.GetBodyString(), r.GetQueryMap())
	startTime := gtime.Now()
	r.Middleware.Next()
	// There's custom buffer content, it then exits current handler.
	if r.Response.BufferLength() > 0 {
		return
	}

	var (
		msg    string
		err    = r.GetError()
		res    = r.GetHandlerResponse()
		status = gerror.Code(err)
	)
	if err != nil {
		if status == gcode.CodeNil {
			status = gcode.CodeInternalError
		}
		msg = err.Error()
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			msg = http.StatusText(r.Response.Status)
			switch r.Response.Status {
			case http.StatusNotFound:
				status = gcode.CodeNotFound
			case http.StatusForbidden:
				status = gcode.CodeNotAuthorized
			default:
				status = gcode.CodeUnknown
			}
			// It creates error as it can be retrieved by other middlewares.
			err = gerror.NewCode(status, msg)
			r.SetError(err)
		} else {
			status = gcode.CodeOK
		}
	}
	processTime := gtime.Now().Sub(startTime)
	// 记录请求路径和处理时间
	glog.Infof(r.Context(), "End Req: %s, Cost: %s", r.URL.Path, processTime.String())
	// res 如果是 g.Map 类型，提取其中的字段到 BaseResponse 结构体中
	// 处理响应内容
	response := BaseResponse{
		Status:    status.Code(),
		Message:   msg,
		RequestId: gctx.CtxId(r.Context()),
	}
	// 如果想将 g.Map 中的字段直接提取到顶层 JSON 中而不是放在 Data 字段下，可以这样修改：
	if res != nil {
		// 检查 res 类型
		// 使用
		if data, ok := extractData(res); ok {
			response.Data = data
		} else {
			response.Data = res
		}
	}
	r.Response.WriteJson(response)
}

type dataHolder interface{ GetData() interface{} }

// 通用取值函数
func extractData(v interface{}) (interface{}, bool) {
	if dh, ok := v.(dataHolder); ok {
		return dh.GetData(), true
	}
	return nil, false
}
