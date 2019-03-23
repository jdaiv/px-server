package rpg

const (
	// _internal_ incoming actions
	ACTION_TICK = "tick"
	// incoming actions
	ACTION_JOIN         = "join"
	ACTION_LEAVE        = "leave"
	ACTION_MOVE         = "move"
	ACTION_FACE         = "face"
	ACTION_USE          = "use"
	ACTION_TAKE_ITEM    = "take_item"
	ACTION_EQUIP_ITEM   = "equip_item"
	ACTION_UNEQUIP_ITEM = "unequip_item"
	ACTION_USE_ITEM     = "use_item"
	ACTION_DROP_ITEM    = "drop_item"
	ACTION_ATTACK       = "attack"
	// outgoing actions
	ACTION_UPDATE        = "state_update"
	ACTION_UPDATE_PLAYER = "player_update"
	ACTION_CHAT          = "chat_message"
	ACTION_EFFECT        = "play_effect"
	// special actions
	ACTION_EDIT = "edit"
)

var PlayerIncomingActions = map[string]bool{
	ACTION_JOIN:         true,
	ACTION_LEAVE:        true,
	ACTION_MOVE:         true,
	ACTION_FACE:         true,
	ACTION_USE:          true,
	ACTION_TAKE_ITEM:    true,
	ACTION_EQUIP_ITEM:   true,
	ACTION_UNEQUIP_ITEM: true,
	ACTION_USE_ITEM:     true,
	ACTION_DROP_ITEM:    true,
	ACTION_ATTACK:       true,
}

type ActionParams map[string]interface{}

func (p ActionParams) getInt(name string) (int, bool) {
	param, ok := p[name]
	if !ok {
		return 0, false
	}

	v, ok := param.(float64)
	if !ok {
		return 0, false
	}

	return int(v), true
}

func (p ActionParams) getFloat(name string) (float64, bool) {
	param, ok := p[name]
	if !ok {
		return 0, false
	}

	v, ok := param.(float64)
	return v, ok
}

func (p ActionParams) getString(name string) (string, bool) {
	param, ok := p[name]
	if !ok {
		return "", false
	}

	v, ok := param.(string)
	return v, ok
}
