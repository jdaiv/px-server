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
		if wsMsg.Action.Scope == "conn" &&
			wsMsg.Action.Type == "close" {
			log.Printf("[ws/recv] %s closed connection", ws.RemoteAddr())
			return
		}
		incoming <- incomingAction{Msg: wsMsg, Source: client}
	}
}

func login(w http.ResponseWriter, r *http.Request) {
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

	password := r.FormValue("password")

	if err := ValidatePassword(password); err != nil {
		if cErr, ok := err.(ClientError); ok {
			jsonErr(w, "auth", "login", cErr)
		} else {
			log.Printf("[api/auth] error validating password: %v", err)
			jsonErr(w, "auth", "login", ErrorInternal)
		}
		return
	}

	nName := normalizeUsername(username)
	user, err := AuthenticateUser(nName, password)
	if err != nil {
		if err == ErrorUserMissing {
			user, err = CreateUser(username, password)
			if err != nil {
				log.Printf("[api/auth] error authenticating user: %v", err)
				jsonErr(w, "auth", "login", ErrorInternal)
				return
			}
		} else if err == ErrorInvalidLogin {
			jsonErr(w, "auth", "login", ErrorInvalidLogin)
			return
		} else {
			log.Printf("[api/auth] error authenticating user: %v", err)
			jsonErr(w, "auth", "login", ErrorInternal)
			return
		}
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.NameNormal
	claims["full_name"] = user.Name
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString(JWTSecret)
	if err != nil {
		log.Printf("[api/auth] error authenticating user: %v", err)
		jsonErr(w, "auth", "login", ErrorInternal)
		return
	}

	log.Printf("[api/auth] %s logged in", user.NameNormal)

	jsonWrite(w, WSResponse{
		Error:   0,
		Message: "success",
		Action:  WSAction{"auth", "login", "all"},
		Data:    t,
	})
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
