package main

import (
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"net/http"
	"os"
)

var indexBytes = []byte("Hestia-chat v0.1\ngithub.com/GDGVIT/hestia-chat")
var ok = []byte{'O', 'K'}

var log *zap.SugaredLogger

func index(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.Write(indexBytes)
}

func alive(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.Write(ok)
}

func init() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log = logger.Sugar()
}

func main() {
	mux := router.New()
	mux.PanicHandler = func(ctx *fasthttp.RequestCtx, i interface{}) {
		ctx.SetStatusCode(http.StatusInternalServerError)
		log.Fatalw("PANIC:",
			"err", i)
	}

	base := os.Getenv("BASE_PATH")

	// general
	mux.GET(base+"/", index)
	mux.GET(base+"/alive", alive)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Infow("Bringing up chat microservice", "port", port)
	log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf(":%s", port), mux.Handler))
}
