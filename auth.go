package main

import (
	"log"
)

func loginHandler(source *Client, target string, data []byte) (interface{}, error) {
	var tokenStr string

	err := parseIncoming(data, &tokenStr)
	if err != nil {
		return nil, ErrorInvalidData
	}

	if err := source.Authenticate(tokenStr); err != nil {
		return nil, err
	}

	log.Printf("[ws/auth] %s logged in as %s",
		source.Conn.RemoteAddr(), source.User.NameNormal)

	return WSResponse{
		Error:  0,
		Action: WSAction{"auth", "login", "all"},
		Data:   map[string]string{"name": source.User.Name},
	}, nil
}

func logoutHandler(source *Client, target string, data []byte) (interface{}, error) {
	return true, source.Logout()
}
