package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ATechnoHazard/hestia-chat/api/views"
	entities2 "github.com/ATechnoHazard/hestia-chat/api/views/entities"
	"github.com/ATechnoHazard/hestia-chat/api/websocket"
	"github.com/ATechnoHazard/hestia-chat/internal/utils"
	"github.com/ATechnoHazard/hestia-chat/pkg/chat"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/wI2L/jettison"
	"io/ioutil"
	"net/http"
)

var AuthUrl = "hestia-auth.herokuapp.com"

func sendMessage(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		// Unmarshal request body
		msg := &entities.Message{}
		if err := json.Unmarshal(ctx.PostBody(), msg); err != nil {
			views.Wrap(ctx, err)
			return
		}

		// Pull token off headers
		token := string(ctx.Request.Header.Peek("Authorization"))

		// Marshal auth request body
		reqBody, _ := jettison.Marshal(map[string]string{"token": token})

		// Send auth request
		resp, err := http.Post(fmt.Sprintf("http://%s/api/user/verify", AuthUrl), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		// Unmarshal auth response
		defer resp.Body.Close()
		authResp := &entities2.AuthResponse{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}
		err = json.Unmarshal(body, authResp)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg.From = authResp.UserID

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

		msgs, err := msgSvc.GetMessages(msg.ReceiverRefer)
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
	r.POST(base+"/sendMessage", sendMessage(msgSvc))
	r.POST(base+"/createChat", createChat(msgSvc))
	r.POST(base+"/getMessages", getChatMessages(msgSvc))
}
