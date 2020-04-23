package views

import (
	"errors"
	"github.com/ATechnoHazard/hestia-chat/pkg"
	"github.com/valyala/fasthttp"
	"github.com/wI2L/jettison"
	"net/http"
)

type ErrView struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var (
	ErrMethodNotAllowed = errors.New("error: Method is not allowed")
	ErrInvalidToken     = errors.New("error: Invalid Authorization token")
	ErrUserExists       = errors.New("error: User already exists")
)

var ErrHTTPStatusMap = map[string]int{
	pkg.ErrNotFound.Error():      http.StatusNotFound,
	pkg.ErrInvalidSlug.Error():   http.StatusBadRequest,
	pkg.ErrAlreadyExists.Error(): http.StatusConflict,
	pkg.ErrNotFound.Error():      http.StatusNotFound,
	pkg.ErrDatabase.Error():      http.StatusInternalServerError,
	pkg.ErrUnauthorized.Error():  http.StatusUnauthorized,
	pkg.ErrForbidden.Error():     http.StatusForbidden,
	ErrMethodNotAllowed.Error():  http.StatusMethodNotAllowed,
	ErrInvalidToken.Error():      http.StatusBadRequest,
	ErrUserExists.Error():        http.StatusConflict,
}

func Wrap(ctx *fasthttp.RequestCtx, err error) {
	ctx.SetContentType("application/json; charset=utf-8")
	code := ErrHTTPStatusMap[err.Error()]
	if code == 0 {
		code = http.StatusInternalServerError
	}

	errView := ErrView{
		Message: err.Error(),
		Status:  code,
	}

	json, _ := jettison.Marshal(errView)
	ctx.SetStatusCode(code)
	_, _ = ctx.Write(json)
}
