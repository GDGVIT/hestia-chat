package websocket

import (
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
)

var Clients = make(map[uint]*websocket.Conn)
var Broadcast = make(chan entities.Message, 5)
var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func HandleConns(ctx *fasthttp.RequestCtx) {
	chatID, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("chat")))
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		Clients[uint(chatID)] = conn
		log.Printf("Connecting to websocket with chatID %v ", chatID)
		select {
		case msg := <-Broadcast:
			log.Println(msg, Clients)
			err := Clients[msg.ReceiverRefer].WriteJSON(msg)
			if err != nil {
				log.Println("Error writing to client", "err", err, "client", conn)
				_ = conn.Close()
				break
			}
		}
	})

	if err != nil {
		log.Fatal(err)
	}
}
