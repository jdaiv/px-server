package main

import (
	"bitbucket.org/panicexpress/backend/rpg"
)

func handleGameState(source *Client, data []byte) (interface{}, error) {
	return WSResponse{
		Error:  0,
		Action: ACTION_GAME_STATE,
		Data:   game.Zones["start"],
	}, nil
}

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
