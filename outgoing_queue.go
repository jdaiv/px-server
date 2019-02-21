package main

import "log"

type outgoingMessage struct {
	Data interface{}
	Dest *Client
}

var outgoing = make(chan outgoingMessage)

func outgoingMessages() {
	for {
		out := <-outgoing
		err := out.Dest.Conn.WriteJSON(out.Data)
		if err != nil {
			log.Fatal(err)
		}
	}
}
