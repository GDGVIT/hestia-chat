package chat

import (
	"encoding/json"
	"github.com/ATechnoHazard/hestia-chat/api/middleware"
	"github.com/ATechnoHazard/hestia-chat/api/views"
	entities2 "github.com/ATechnoHazard/hestia-chat/api/views/entities"
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
		msg.Sender = uint(userID)

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

		msgs, err := msgSvc.GetMessages(msg.ReceiverRefer, msg.Sender)
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

func getChatsForUser(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		user := &entities2.User{}
		if err := json.Unmarshal(ctx.PostBody(), user); err != nil {
			views.Wrap(ctx, err)
			return
		}
		chats, err := msgSvc.GetChatsByID(user.ID)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Successfully fetched chats for user")
		msg["chats"] = chats
		utils.Respond(ctx, msg)
		return
	}
}

func getMyChats(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		user := &entities2.User{}
		if err := json.Unmarshal(ctx.PostBody(), user); err != nil {
			views.Wrap(ctx, err)
			return
		}

		chats, err := msgSvc.GetMyChats(user.ID)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Successfully fetched my chats")
		msg["chats"] = chats
		utils.Respond(ctx, msg)
		return
	}
}

func getOtherChats(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		user := &entities2.User{}
		if err := json.Unmarshal(ctx.PostBody(), user); err != nil {
			views.Wrap(ctx, err)
			return
		}

		chats, err := msgSvc.GetOtherChats(user.ID)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Successfully fetched my chats")
		msg["chats"] = chats
		utils.Respond(ctx, msg)
		return
	}
}

func MakeMessageHandler(r *router.Router, msgSvc chat.Service, base string) {
	r.POST(base+"/sendMessage", middleware.JwtAuth(sendMessage(msgSvc)))
	r.POST(base+"/createChat", middleware.JwtAuth(createChat(msgSvc)))
	r.POST(base+"/getMessages", middleware.JwtAuth(getChatMessages(msgSvc)))
	r.POST(base+"/getChats", middleware.JwtAuth(getChatsForUser(msgSvc)))
	r.POST(base+"/getMyChats", middleware.JwtAuth(getMyChats(msgSvc)))
	r.POST(base+"/getOtherChats", middleware.JwtAuth(getOtherChats(msgSvc)))
}
