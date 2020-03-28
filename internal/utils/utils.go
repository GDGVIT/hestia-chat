package utils

import (
	"github.com/valyala/fasthttp"
	"github.com/wI2L/jettison"
)

func Message(status int, message string) map[string]interface{} {
	return map[string]interface{}{"code": status, "message": message}
}

func Respond(ctx *fasthttp.RequestCtx, data map[string]interface{}) {
	ctx.SetContentType("application/json; charset=utf-8")
	d, _ := jettison.Marshal(data)
	_, _ = ctx.Write(d)
}
