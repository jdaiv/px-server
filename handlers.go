package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type WSAction struct {
	Scope  string `json:"scope"`
	Type   string `json:"type"`
	Target string `json:"target"`
}

type WSMessage struct {
	Action WSAction `json:"action"`
	Data   string   `json:"data"`
}

type WSResponse struct {
	Error   int         `json:"error"`
	Message string      `json:"message"`
	Action  WSAction    `json:"action"`
	Data    interface{} `json:"data"`
}

type incomingAction struct {
	Msg    WSMessage
	Source *Client
}

var incoming = make(chan incomingAction)
var wsRouter *WSRouter

func configureWSRoutes() {
	wsRouter = NewWSRouter()

	wsRouter.AddHandler("global", "ping", pingHandler)

	wsRouter.AddHandler("auth", "login", loginHandler)
	wsRouter.AddHandler("auth", "logout", logoutHandler)

	wsRouter.AddHandler("chat", "message", handleChatMessage)
	wsRouter.AddHandler("chat", "list_rooms", handleListRooms)
	wsRouter.AddHandler("chat", "list_users", handleListUsers)
	wsRouter.AddHandler("chat", "join_room", handleJoinRoom)
	wsRouter.AddHandler("chat", "create_room", handleCreateRoom)
	wsRouter.AddHandler("chat", "update_room", handleModifyRoom)

	wsRouter.AddHandler("activity", "list", handleActivityList)
	wsRouter.AddDefaultHandler("activity", handleActivityAction)
}

func parseIncoming(data []byte, v interface{}) error {
	err := json.Unmarshal([]byte(data), v)
	if err != nil {
		// log.Printf("[server] error: %v", err)
		return fmt.Errorf("json payload: invalid data")
	}
	return nil
}

func incomingMessages() {
	for {
		in := <-incoming
		var err error
		var response interface{}
		handler, err := wsRouter.GetHandler(in.Msg.Action.Scope, in.Msg.Action.Type)
		if err != nil {
			var defaultHandler WSDefaultHandler
			defaultHandler, err = wsRouter.GetDefaultHandler(in.Msg.Action.Scope)
			if err != nil {
				log.Printf("[ws/%s/%s] handler not found",
					in.Msg.Action.Scope, in.Msg.Action.Type)
				err = ErrorMissingAction
			} else {
				response, err = defaultHandler(in.Source, in.Msg.Action.Target, in.Msg.Action.Type, []byte(in.Msg.Data))
			}
		} else {
			response, err = handler(in.Source, in.Msg.Action.Target, []byte(in.Msg.Data))
		}
		if err != nil {
			cErr, ok := err.(ClientError)
			if !ok {
				log.Printf("[ws/send] uncaught error: %v", err)
				cErr = ErrorInternal
			}
			log.Printf("[ws/send] error: %v", cErr)
			jsonErr := in.Source.Conn.WriteJSON(WSResponse{
				Error:   cErr.Code(),
				Message: cErr.ExternalMessage(),
				Action:  in.Msg.Action,
				Data:    nil,
			})
			if jsonErr != nil {
				log.Printf("[server] error sending json payload: %v", err)
			}
		}
		if response != nil {
			wsRespsonse, ok := response.(WSResponse)
			if !ok {
				wsRespsonse = WSResponse{
					Error:   0,
					Message: "success",
					Action:  in.Msg.Action,
					Data:    response,
				}
			}
			if err := in.Source.Conn.WriteJSON(wsRespsonse); err != nil {
				log.Printf("[server] error sending json payload: %v", err)
			}
		}
	}
}

func pingHandler(source *Client, target string, data []byte) (interface{}, error) {
	log.Printf("[ws/ping] hello %s", source.Conn.RemoteAddr())
	return WSResponse{
		Error:  0,
		Action: WSAction{"global", "pong", "all"},
	}, nil
}
