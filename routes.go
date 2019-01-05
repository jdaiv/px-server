package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		// temporary workaround for local dev
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func join(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[api/ws] error upgrading %s: %v", r.Host, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("bad things"))
		return
	}
	defer ws.Close()

	client := MakeClient(ws)
	defer RemoveClient(client)

	for {
		var wsMsg WSMessage
		err := ws.ReadJSON(&wsMsg)
		if _, ok := err.(*websocket.CloseError); ok {
			log.Printf("[ws/recv] goodbye %s", ws.RemoteAddr())
			return
		} else if err != nil {
			log.Printf("[ws/recv] error: %v", err)
			return
		}
		// log.Printf("[ws_recv] %s/%s", wsMsg.Scope, wsMsg.Action)
		if wsMsg.Action.Scope == "conn" &&
			wsMsg.Action.Type == "close" {
			log.Printf("[ws/recv] %s closed connection", ws.RemoteAddr())
			return
		}
		incoming <- incomingAction{Msg: wsMsg, Source: client}
	}
}

func login(w http.ResponseWriter, r *http.Request) {

	password := r.FormValue("password")

	if len(password) > 0 {
		if err := ValidatePassword(password); err != nil {
			if cErr, ok := err.(ClientError); ok {
				jsonErr(w, "auth", "login", cErr)
			} else {
				log.Printf("[api/auth] error validating password: %v", err)
				jsonErr(w, "auth", "login", ErrorInternal)
			}
			return
		}

		user, err := AuthenticateUser(password)
		if err != nil {
			jsonErr(w, "auth", "login", ErrorInvalidLogin)
			return
		}

		jsonWrite(w, WSResponse{
			Error:   0,
			Message: "success",
			Action:  WSAction{"auth", "valid", ""},
			Data:    true,
		})

		log.Printf("[api/auth] user logged in %s", user.NameNormal)
	}

	username := r.FormValue("username")

	if err := ValidateUsername(username); err != nil {
		if cErr, ok := err.(ClientError); ok {
			jsonErr(w, "auth", "login", cErr)
		} else {
			log.Printf("[api/auth] error validating username: %v", err)
			jsonErr(w, "auth", "login", ErrorInternal)
		}
		return
	}

	user, password, err := CreateUser(username)
	if err != nil {
		jsonErr(w, "auth", "login", ErrorInternal)
		return
	}

	jsonWrite(w, WSResponse{
		Error:   0,
		Message: "success",
		Action:  WSAction{"auth", "create", ""},
		Data:    password,
	})

	log.Printf("[api/auth] created user %s", user.NameNormal)
}

func jsonWrite(w http.ResponseWriter, v interface{}) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(v)
	if err != nil {
		log.Printf("[server] error encoding json: %v", err)
	}
}

func jsonErr(w http.ResponseWriter, scope, action string, err ClientError) {
	jsonWrite(w, WSResponse{
		Error:   err.Code(),
		Message: err.ExternalMessage(),
		Action:  WSAction{scope, action, "all"},
	})
}
