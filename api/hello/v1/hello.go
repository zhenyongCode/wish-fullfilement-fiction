package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type HelloReq struct {
	g.Meta `path:"/hello" tags:"Hello" method:"get,post" summary:"You first hello api"`
	Data   g.MapAnyAny `json:"data" example:"{\"func:\": \"nana\", param\": \"hello\"}"`
}
type HelloRes struct {
	g.Meta `mime:"text/html" example:"string"`
	Data   g.Map `json:"data"`
}

// GetData HelloRes 实现它
func (h *HelloRes) GetData() interface{} { return h.Data }
