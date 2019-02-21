package main

func handleGameState(source *Client, data []byte) (interface{}, error) {
	return WSResponse{
		Error:  0,
		Action: ACTION_GAME_STATE,
		Data:   game.Zones["start"],
	}, nil
}
