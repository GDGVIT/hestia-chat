package main

import (
	"fmt"
	chat2 "github.com/ATechnoHazard/hestia-chat/api/chat"
	"github.com/ATechnoHazard/hestia-chat/api/middleware"
	"github.com/ATechnoHazard/hestia-chat/api/websocket"
	"github.com/ATechnoHazard/hestia-chat/pkg/chat"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/fasthttp/router"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
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

	if os.Getenv("ENV") != "PROD" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func makeDB() *gorm.DB {
	conn, err := pq.ParseURL(os.Getenv("DB_URI"))
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open("postgres", conn)
	if err != nil {
		log.Fatal(err)
	}

	if os.Getenv("DEBUG") == "true" {
		db = db.Debug()
	}
	log.Infow("Automigrating db")
	db.AutoMigrate(&entities.Message{}, &entities.Chat{})
	return db
}

func main() {
	mux := router.New()
	mux.PanicHandler = func(ctx *fasthttp.RequestCtx, i interface{}) {
		ctx.SetStatusCode(http.StatusInternalServerError)
		log.Fatalw("PANIC:",
			"err", i)
	}

	base := os.Getenv("BASE_PATH")

	db := makeDB()
	msgSvc := chat.NewChatService(db)

	// general
	mux.GET(base+"/", index)
	mux.GET(base+"/alive", alive)
	mux.GET(base+"/ws", websocket.HandleConns)
	chat2.MakeMessageHandler(mux, msgSvc, base)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	log.Infow("Bringing up websocket dispatcher")
	//go websocket.HandleMessages()

	log.Infow("Bringing up chat microservice", "port", port, "base path", base)
	log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf(":%s", port), middleware.CORS(mux.Handler)))
}
