package main

import (
	"log"
)

func loginHandler(source *Client, data []byte) (interface{}, error) {
	var tokenStr string

	err := parseIncoming(data, &tokenStr)
	if err != nil {
		return nil, ErrorInvalidData
	}

	if err := source.Authenticate(tokenStr); err != nil {
		return nil, err
	}

	log.Printf("[ws/auth] %s logged in as %s (%v)",
		source.Conn.RemoteAddr(), source.User.NameNormal, source.User.SuperUser)

	return WSResponse{
		Error:  0,
		Action: ACTION_LOGIN,
		Data:   map[string]string{"name": source.User.Name},
	}, nil
}
