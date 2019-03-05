package rpg

const (
	// _internal_ incoming actions
	ACTION_TICK = "tick"
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

var PlayerIncomingActions = map[string]bool{
	ACTION_JOIN:         true,
	ACTION_LEAVE:        true,
	ACTION_MOVE:         true,
	ACTION_USE:          true,
	ACTION_TAKE_ITEM:    true,
	ACTION_EQUIP_ITEM:   true,
	ACTION_UNEQUIP_ITEM: true,
	ACTION_USE_ITEM:     true,
	ACTION_DROP_ITEM:    true,
}
