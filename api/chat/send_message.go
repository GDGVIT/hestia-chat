package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ATechnoHazard/hestia-chat/api/middleware"
	"github.com/ATechnoHazard/hestia-chat/api/views"
	entities2 "github.com/ATechnoHazard/hestia-chat/api/views/entities"
	"github.com/ATechnoHazard/hestia-chat/api/websocket"
	"github.com/ATechnoHazard/hestia-chat/internal/utils"
	"github.com/ATechnoHazard/hestia-chat/pkg"
	"github.com/ATechnoHazard/hestia-chat/pkg/chat"
	"github.com/ATechnoHazard/hestia-chat/pkg/entities"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/wI2L/jettison"
	"log"
	"net/http"
	"os"
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
		//userID, _ := strconv.Atoi(string(ctx.Request.Header.Peek("user_id")))

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

		ch := &entities.Chat{RequestSender: msg.Sender, RequestReceiver: msg.ReceiverRefer}
		ch, err := msgSvc.GetChat(ch)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		if ch.IsReported {
			retMsg := make(map[string]interface{})
			retMsg["status"] = 400
			retMsg["messages"] = "this chat is blocked"
			ctx.SetContentType("application/json; charset=utf-8")
			ctx.SetStatusCode(400)
			d, _ := jettison.Marshal(retMsg)
			_, _ = ctx.Write(d)
			return
		}

		msgs, items, err := msgSvc.GetMessages(msg.ReceiverRefer, msg.Sender)
		if err != nil {
			views.Wrap(ctx, err)
			return
		}

		retMsg := utils.Message(http.StatusOK, "Successfully retrieved chat messages")
		retMsg["messages"] = msgs
		retMsg["items"] = items
		utils.Respond(ctx, retMsg)
		return
	}
}

func createChat(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		chatRoom := &entities.Chat{}
		if err := json.Unmarshal(ctx.PostBody(), chatRoom); err != nil {
			log.Println(string(ctx.PostBody()))
			views.Wrap(ctx, err)
			return
		}

		postBody1, _ := jettison.Marshal(&entities2.UserDetails{ID: chatRoom.RequestSender})
		resp, err := http.Post(fmt.Sprintf("%s/getDetailsById", os.Getenv("AUTH_SERVICE")), "application/json", bytes.NewBuffer(postBody1))
		if err != nil {
			views.Wrap(ctx, err)
			return
		}
		ud1 := &entities2.UserDetails{}
		if err := json.NewDecoder(resp.Body).Decode(ud1); err != nil {
			views.Wrap(ctx, err)
			return
		}

		postBody2, _ := jettison.Marshal(&entities2.UserDetails{ID: chatRoom.RequestReceiver})
		resp2, err := http.Post(fmt.Sprintf("%s/getDetailsById", os.Getenv("AUTH_SERVICE")), "application/json", bytes.NewBuffer(postBody2))
		if err != nil {
			views.Wrap(ctx, err)
			return
		}
		ud2 := &entities2.UserDetails{}
		if err := json.NewDecoder(resp2.Body).Decode(ud2); err != nil {
			views.Wrap(ctx, err)
			return
		}

		chatRoom.SenderName = ud1.Name
		chatRoom.ReceiverName = ud2.Name

		if err := msgSvc.CreateChat(chatRoom); err != nil {
			if err == pkg.ErrAlreadyExists {
				item := &entities.Item{
					RequestSender:   chatRoom.RequestSender,
					RequestReceiver: chatRoom.RequestReceiver,
					Item:            chatRoom.Title,
					ReqDesc:         chatRoom.ReqDesc,
				}

				if err := msgSvc.AddItem(item); err != nil {
					views.Wrap(ctx, err)
					return
				}

				ctx.SetStatusCode(http.StatusInternalServerError)
				msg := utils.Message(http.StatusInternalServerError, "Chat already exists")
				msg["chat_room"] = chatRoom
				utils.Respond(ctx, msg)
				return
			}
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

		msg := utils.Message(http.StatusOK, "Successfully fetched other chats")
		msg["chats"] = chats
		utils.Respond(ctx, msg)
		return
	}
}

func delChat(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		delReq := &entities2.DelReq{}
		if err := json.Unmarshal(ctx.PostBody(), delReq); err != nil {
			views.Wrap(ctx, err)
			return
		}

		if err := msgSvc.DeleteChat(delReq.Receiver, delReq.Sender, delReq.WhoDeleted); err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Chat deleted successfully")
		utils.Respond(ctx, msg)
		return
	}
}

func updateChat(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		updateChat := &entities.Chat{}
		if err := json.Unmarshal(ctx.PostBody(), updateChat); err != nil {
			views.Wrap(ctx, err)
			return
		}

		if err := msgSvc.UpdateChat(updateChat); err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Chat updated successfully")
		utils.Respond(ctx, msg)
		return
	}
}

func addItem(msgSvc chat.Service) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		updateChat := &entities.Chat{}
		if err := json.Unmarshal(ctx.PostBody(), updateChat); err != nil {
			views.Wrap(ctx, err)
			return
		}

		item := &entities.Item{
			RequestSender:   updateChat.RequestSender,
			RequestReceiver: updateChat.RequestReceiver,
			Item:            updateChat.Title,
			ReqDesc:         updateChat.ReqDesc,
		}

		if err := msgSvc.AddItem(item); err != nil {
			views.Wrap(ctx, err)
			return
		}

		msg := utils.Message(http.StatusOK, "Items updated successfully")
		utils.Respond(ctx, msg)
		return
	}
}

func MakeMessageHandler(r *router.Router, msgSvc chat.Service, base string) {
	r.POST(base+"/sendMessage", middleware.JwtAuth(sendMessage(msgSvc)))
	r.POST(base+"/createChat", middleware.JwtAuth(createChat(msgSvc)))
	r.POST(base+"/getMessages", middleware.JwtAuth(getChatMessages(msgSvc)))
	r.POST(base+"/getChats", middleware.JwtAuth(getChatsForUser(msgSvc)))
	r.POST(base+"/getOtherChats", middleware.JwtAuth(getMyChats(msgSvc)))
	r.POST(base+"/getMyChats", middleware.JwtAuth(getOtherChats(msgSvc)))
	r.DELETE(base+"/delChat", middleware.JwtAuth(delChat(msgSvc)))
	r.POST(base+"/updateChat", middleware.JwtAuth(updateChat(msgSvc)))
	r.POST(base+"/addItem", middleware.JwtAuth(addItem(msgSvc)))
}
