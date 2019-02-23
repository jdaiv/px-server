package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type ActionStr string

const (
	ACTION_CLOSE        = "close"
	ACTION_PING         = "ping"
	ACTION_LOGIN        = "login"
	ACTION_CREATE_USER  = "create_user"
	ACTION_CHAT_MESSAGE = "chat_message"
	ACTION_LIST_USERS   = "list_users"
	ACTION_GAME_STATE   = "game_state"
	ACTION_GAME_ACTION  = "game_action"
)

type WSMessage struct {
	Action ActionStr `json:"action"`
	Data   string    `json:"data"`
}

type WSResponse struct {
	Error   int         `json:"error"`
	Message string      `json:"message"`
	Action  ActionStr   `json:"action"`
	Data    interface{} `json:"data"`
}

type incomingAction struct {
	Msg    WSMessage
	Source *Client
}

var incoming = make(chan incomingAction)

type WSHandler func(source *Client, data []byte) (interface{}, error)

var wsRouter = map[ActionStr]WSHandler{
	ACTION_PING:  pingHandler,
	ACTION_LOGIN: loginHandler,
	// "logout":       logoutHandler,

	ACTION_CHAT_MESSAGE: handleChatMessage,
	ACTION_LIST_USERS:   handleListUsers,

	ACTION_GAME_STATE:  handleGameState,
	ACTION_GAME_ACTION: handleGameAction,
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

		handler, exists := wsRouter[in.Msg.Action]
		if !exists {
			log.Printf("[ws/%s] handler not found", in.Msg.Action)
			err = ErrorMissingAction
		} else {
			response, err = handler(in.Source, []byte(in.Msg.Data))
		}

		if err != nil {
			cErr, ok := err.(ClientError)
			if !ok {
				log.Printf("[ws/send] uncaught error: %v", err)
				cErr = ErrorInternal
			}
			log.Printf("[ws/send] error: %v", cErr)
			in.Source.Write(WSResponse{
				Error:   cErr.Code(),
				Message: cErr.ExternalMessage(),
				Action:  in.Msg.Action,
				Data:    nil,
			})
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
			in.Source.Write(wsRespsonse)
		}
	}
}

func pingHandler(source *Client, data []byte) (interface{}, error) {
	// log.Printf("[ws/ping] hello %s", source.Conn.RemoteAddr())
	now := time.Now().UnixNano()
	source.LastPing = now
	return WSResponse{
		Error:  0,
		Action: ACTION_PING,
		Data:   now,
	}, nil
}
