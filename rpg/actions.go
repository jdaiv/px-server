package rpg

const (
	// incoming actions
	ACTION_JOIN  = "join"
	ACTION_LEAVE = "leave"
	ACTION_MOVE  = "move"
	ACTION_USE   = "use"

	// outgoing actions
	ACTION_UPDATE        = "state_update"
	ACTION_UPDATE_PLAYER = "player_update"
	ACTION_CHAT          = "chat_message"
)
