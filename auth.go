package main

import (
	"fmt"
	"log"

	jwt "github.com/dgrijalva/jwt-go"
)

var authenticatedUsers = make(map[string]*Client)

func handleAuthAction(source *Client, action string, data []byte) (interface{}, error) {
	switch action {
	case "login":
		var tokenStr string

		err := parseIncoming(data, &tokenStr)
		if err != nil {
			return nil, ErrorInvalidData
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return JWTSecret, nil
		})

		if err != nil {
			return nil, ErrorInvalidToken
		}

		var username string
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			username = claims["name"].(string)
		} else {
			return nil, ErrorInvalidToken
		}

		// check if user is already connected

		if client, exists := authenticatedUsers[username]; exists {
			client.Conn.Close()
		}

		user, err := LoadUser(username)
		if err != nil {
			return nil, ErrorUserMissing
		}

		source.User = user
		source.Authenticated = true

		authenticatedUsers[username] = source

		log.Printf("[ws/auth] %s logged in as %s", source.Conn.RemoteAddr(), user.Name)

		return WSResponse{
			Error:   0,
			Message: "success",
			Scope:   "auth",
			Action:  "login",
			Data:    map[string]string{"name": user.Name},
		}, nil
	case "logout":
		source.Authenticated = false
		source.User = User{}
		return WSResponse{
			Error:   0,
			Message: "success",
			Scope:   "auth",
			Action:  "logout",
			Data:    nil,
		}, nil
	}
	return nil, ErrorMissingAction
}
