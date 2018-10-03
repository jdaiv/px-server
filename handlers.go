package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type WSMessage struct {
	Action string `json:"action"`
	Data   string `json:"data"`
	Scope  string `json:"scope"`
}

type WSResponse struct {
	Error   int         `json:"error"`
	Message string      `json:"message"`
	Scope   string      `json:"scope"`
	Action  string      `json:"action"`
	Data    interface{} `json:"data"`
}

type incomingAction struct {
	Msg    WSMessage
	Source *Client
}

type ScopeHandler func(source *Client, action string, data []byte) (interface{}, error)

var incoming = make(chan incomingAction)
var handlers = map[string]ScopeHandler{
	"chat":     handleChatAction,
	"auth":     handleAuthAction,
	"global":   handleGlobalAction,
	"activity": handleActivityAction,
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
		scopeHandler, exists := handlers[in.Msg.Scope]
		if !exists {
			log.Printf("scope not found: %s", in.Msg.Scope)
			break
		}
		response, err := scopeHandler(in.Source, in.Msg.Action, []byte(in.Msg.Data))
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
				Scope:   in.Msg.Scope,
				Action:  in.Msg.Action,
				Data:    nil,
			})
			if jsonErr != nil {
				log.Printf("[server] error sending json payload: %v", err)
			}
		}
		if response != nil {
			if err := in.Source.Conn.WriteJSON(response); err != nil {
				log.Printf("[server] error sending json payload: %v", err)
			}
		}
	}
}

func handleGlobalAction(source *Client, action string, data []byte) (interface{}, error) {
	switch action {
	case "ping":
		log.Printf("[ws/ping] hello %s", source.Conn.RemoteAddr())
		return WSResponse{
			Error:   0,
			Message: "success",
			Scope:   "global",
			Action:  "pong",
			Data:    nil,
		}, nil
	}
	return nil, ErrorMissingAction
}
