package rpg

const (
	SEQ_ACTION_ANIM = iota
	SEQ_ACTION_DAMAGE
	SEQ_ACTION_EFFECT
)

const (
	SEQ_TARGET_TYPE_PLAYER = iota
	SEQ_TARGET_TYPE_NPC
	SEQ_TARGET_TYPE_XY
)

type Sequence struct {
	TotalDuration int
	Actions       []SeqAction
	ActionIdx     int
	ActionTimer   int
	Done          bool
}

type SeqAction struct {
	// Generic Params
	Type     int
	Duration int

	// Specific Params
	Damage      DamageInfo
	TargetType  int
	TargetX     int
	TargetY     int
	SourceType  int
	SourceX     int
	SourceY     int
	EffectName  string
	AnimationId string
}

func NewSequence() *Sequence {
	return &Sequence{Actions: make([]SeqAction, 0)}
}

func (s *Sequence) AddDamage(info DamageInfo, targetId int, targetIsNpc bool) {
	var targetType int
	if targetIsNpc {
		targetType = SEQ_TARGET_TYPE_NPC
	} else {
		targetType = SEQ_TARGET_TYPE_PLAYER
	}
	action := SeqAction{
		Type:       SEQ_ACTION_DAMAGE,
		Damage:     info,
		TargetType: targetType,
		TargetX:    targetId,
	}
	s.Actions = append(s.Actions, action)
}

func (s *Sequence) AddAnim(animId string, targetId int, targetIsNpc bool, duration int) {
	var targetType int
	if targetIsNpc {
		targetType = SEQ_TARGET_TYPE_NPC
	} else {
		targetType = SEQ_TARGET_TYPE_PLAYER
	}
	action := SeqAction{
		Type:        SEQ_ACTION_ANIM,
		Duration:    duration,
		AnimationId: animId,
		TargetType:  targetType,
		TargetX:     targetId,
	}
	s.Actions = append(s.Actions, action)
	s.TotalDuration += duration
}

func (s *Sequence) AddEffect(effectName string, targetX, targetY int, duration int) {
	action := SeqAction{
		Type:       SEQ_ACTION_EFFECT,
		Duration:   duration,
		EffectName: effectName,
		TargetType: SEQ_TARGET_TYPE_XY,
		TargetX:    targetX,
		TargetY:    targetY,
	}
	s.Actions = append(s.Actions, action)
	s.TotalDuration += duration
}

func (s *Sequence) AddSpellEffect(effectName string, sourceId int, sourceIsNpc bool, targetX, targetY int, duration int) {
	var sourceType int
	if sourceIsNpc {
		sourceType = SEQ_TARGET_TYPE_NPC
	} else {
		sourceType = SEQ_TARGET_TYPE_PLAYER
	}
	action := SeqAction{
		Type:       SEQ_ACTION_EFFECT,
		Duration:   duration,
		EffectName: effectName,
		SourceType: sourceType,
		SourceX:    sourceId,
		TargetType: SEQ_TARGET_TYPE_XY,
		TargetX:    targetX,
		TargetY:    targetY,
	}
	s.Actions = append(s.Actions, action)
	s.TotalDuration += duration
}
