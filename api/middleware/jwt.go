package middleware

import (
	"github.com/ATechnoHazard/hestia-chat/internal/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	"net/http"
	"os"
	"strconv"
)

type Token struct {
	jwt.Claims
	UserID uint `json:"_id"`
}

func JwtAuth(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var response map[string]interface{}
		tokenHeader := string(ctx.Request.Header.Peek("Authorization")) // Grab the token from the header

		if tokenHeader == "" { // Token is missing, returns with error code 403 Unauthorized
			response = utils.Message(http.StatusForbidden, "Missing auth token")
			ctx.SetStatusCode(http.StatusForbidden)
			utils.Respond(ctx, response)
			return
		}

		tk := &Token{}

		token, err := jwt.ParseWithClaims(tokenHeader, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("TOKEN_PASSWORD")), nil
		})

		if err != nil { // Malformed token, returns with http code 403 as usual
			response = utils.Message(http.StatusForbidden, "Malformed authentication token")
			ctx.SetStatusCode(http.StatusForbidden)
			utils.Respond(ctx, response)
			return
		}

		if !token.Valid { // Token is invalid, maybe not signed on this server
			response = utils.Message(http.StatusForbidden, "Token is not valid")
			ctx.SetStatusCode(http.StatusForbidden)
			utils.Respond(ctx, response)
			return
		}

		ctx.Request.Header.Set("user_id", strconv.Itoa(int(tk.UserID)))
		next(ctx) // proceed in the middleware chain
	}
}
