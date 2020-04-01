package websocket

import (
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"log"
	"strconv"
)

var Clients = make(map[sendKey]*websocket.Conn)
var Broadcast = make(chan entities.Message, 5)
var upgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type sendKey struct {
	Sender   uint
	Receiver uint
}

func HandleConns(ctx *fasthttp.RequestCtx) {
	sender, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("sender")))
	receiver, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("receiver")))

	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		Clients[sendKey{Sender: uint(sender), Receiver: uint(receiver)}] = conn
		defer func() {
			if r := recover(); r != nil {
				log.Println("recovered from panic")
			}
		}()
		log.Printf("Connecting to websocket with sender %v ", sender)
		select {
		case msg := <-Broadcast:
			log.Println(msg, Clients)
			err := Clients[sendKey{Sender: msg.Sender, Receiver: msg.ReceiverRefer}].WriteJSON(msg)
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
