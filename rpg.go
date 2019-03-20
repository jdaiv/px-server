package main

import (
	"bitbucket.org/panicexpress/backend/rpg"
)

func handleGameAction(source *Client, data []byte) (interface{}, error) {
	if !source.Authenticated {
		return nil, ErrorUnauthenticated
	}

	var msg rpg.IncomingMessageData

	if err := parseIncoming(data, &msg); err != nil {
		return nil, err
	}

	if legal, ok := rpg.PlayerIncomingActions[msg.Type]; !legal || !ok {
		return nil, nil
	}

	game.Incoming <- rpg.IncomingMessage{PlayerId: source.User.Id, Data: msg}

	return nil, nil
}

func handleGameEditAction(source *Client, data []byte) (interface{}, error) {
	if !source.Authenticated || !source.User.SuperUser {
		return nil, ErrorUnauthenticated
	}

	var msg rpg.IncomingMessageData

	if err := parseIncoming(data, &msg); err != nil {
		return nil, err
	}

	if msg.Type != rpg.ACTION_EDIT {
		return nil, nil
	}

	game.Incoming <- rpg.IncomingMessage{PlayerId: source.User.Id, Data: msg}

	return nil, nil
}
