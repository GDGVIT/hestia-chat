package chat

import (
	"encoding/json"
	"github.com/ATechnoHazard/hestia-chat/api/middleware"
	"github.com/ATechnoHazard/hestia-chat/api/views"
	"github.com/ATechnoHazard/hestia-chat/api/websocket"
	"github.com/ATechnoHazard/hestia-chat/internal/utils"
	"github.com/ATechnoHazard/hestia-chat/pkg/chat"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)

func sendMessage(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		// Unmarshal request body
		msg := &entities.Message{}
		if err := json.Unmarshal(ctx.PostBody(), msg); err != nil {
			views.Wrap(ctx, err)
			return
		}

		// Pull token off headers
		userID, _ := strconv.Atoi(string(ctx.Request.Header.Peek("user_id")))
		msg.From = uint(userID)

		// Save message to db
		if err := msgSvc.SaveMessage(msg); err != nil {
			views.Wrap(ctx, err)
			return
		}

		// Add message to broadcast channel
		websocket.Broadcast <- *msg
		utils.Respond(ctx, utils.Message(http.StatusOK, "Successfully sent message"))
		return
	}
}

func getChatMessages(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		msg := &entities.Message{}
		if err := json.Unmarshal(ctx.PostBody(), msg); err != nil {
			views.Wrap(ctx, err)
			return
		}

		msgs, err := msgSvc.GetMessages(msg.ReceiverRefer, msg.From)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		retMsg := utils.Message(http.StatusOK, "Successfully retrieved chat messages")
		retMsg["messages"] = msgs
		utils.Respond(ctx, retMsg)
		return
	}
}

func createChat(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		chatRoom := &entities.Chat{}
		if err := json.Unmarshal(ctx.PostBody(), chatRoom); err != nil {
			views.Wrap(ctx, err)
			return
		}

		if err := msgSvc.CreateChat(chatRoom); err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Successfully created chat room")
		msg["chat_room"] = chatRoom
		utils.Respond(ctx, msg)
		return
	}
}

func MakeMessageHandler(r *router.Router, msgSvc chat.Service, base string) {
	r.POST(base+"/sendMessage", middleware.JwtAuth(sendMessage(msgSvc)))
	r.POST(base+"/createChat", middleware.JwtAuth(createChat(msgSvc)))
	r.POST(base+"/getMessages", middleware.JwtAuth(getChatMessages(msgSvc)))
}
