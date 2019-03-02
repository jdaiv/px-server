package rpg

const (
	// incoming actions
	ACTION_JOIN         = "join"
	ACTION_LEAVE        = "leave"
	ACTION_MOVE         = "move"
	ACTION_USE          = "use"
	ACTION_TAKE_ITEM    = "take_item"
	ACTION_EQUIP_ITEM   = "equip_item"
	ACTION_UNEQUIP_ITEM = "unequip_item"
	ACTION_USE_ITEM     = "use_item"
	ACTION_DROP_ITEM    = "drop_item"

	// outgoing actions
	ACTION_UPDATE        = "state_update"
	ACTION_UPDATE_PLAYER = "player_update"
	ACTION_CHAT          = "chat_message"
	ACTION_EFFECT        = "play_effect"
)
