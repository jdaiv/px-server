package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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
		if wsMsg.Scope == "conn" && wsMsg.Action == "close" {
			log.Printf("[ws/recv] %s closed connection", ws.RemoteAddr())
			return
		}
		incoming <- incomingAction{Msg: wsMsg, Source: client}
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")

	if len(username) < 1 {
		jsonErr(w, http.StatusBadRequest,
			"auth", "login", "username missing")
		return
	} else if len(username) > 255 {
		jsonErr(w, http.StatusBadRequest,
			"auth", "login", "username too long")
		return
	}

	password := r.FormValue("password")

	if len(password) < 8 {
		jsonErr(w, http.StatusBadRequest,
			"auth", "login", "password too short")
		return
	} else if len(password) > 512 {
		jsonErr(w, http.StatusBadRequest,
			"auth", "login", "password too long")
		return
	}

	user, err := AuthenticateUser(username, password)
	if err != nil {
		if err == ErrorUserMissing {
			user, err = CreateUser(username, password)
			if err != nil {
				log.Printf("[api/auth] error authenticating user: %v", err)
				jsonErr(w, http.StatusInternalServerError,
					"auth", "login", "internal error")
				return
			}
		} else if err == ErrorInvalidPassword {
			jsonErr(w, http.StatusBadRequest,
				"auth", "login", "invalid password")
			return
		} else {
			log.Printf("[api/auth] error authenticating user: %v", err)
			jsonErr(w, http.StatusInternalServerError,
				"auth", "login", "internal error")
			return
		}
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.Name
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString(JWTSecret)
	if err != nil {
		log.Printf("[api/auth] error authenticating user: %v", err)
		jsonErr(w, http.StatusInternalServerError,
			"auth", "login", "internal error")
		return
	}

	log.Printf("[api/auth] %s logged in", user.Name)

	jsonWrite(w, http.StatusOK, WSResponse{
		Error:   0,
		Message: "success",
		Scope:   "auth",
		Action:  "login",
		Data:    t,
	})
}

func jsonWrite(w http.ResponseWriter, code int, v interface{}) {
	// w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(v)
	if err != nil {
		log.Printf("[server] error encoding json: %v", err)
	}
}

func jsonErr(w http.ResponseWriter, code int, scope, action, msg string) {
	jsonWrite(w, code, WSResponse{
		Error:   -1,
		Message: msg,
		Scope:   scope,
		Action:  action,
	})
}
