package domain

import "time"

// Vec2 は 2 次元座標を表す値オブジェクト。
type Vec2 struct {
	X float64
	Y float64
}

// BuffEffect はバフ効果の識別子と量、持続時間を表す。
type BuffEffect struct {
	EffectID  string
	Magnitude float64
	Duration  time.Duration
	Tags      []string
}

// ItemChange はインベントリの増減を表すドメイン値オブジェクト。
type ItemChange struct {
	ItemID        string
	QuantityDelta int
	Metadata      map[string]string
}

// MoveCommand は移動処理に必要なドメインコマンド。
type MoveCommand struct {
	ActorID      string
	RoomID       string
	NextPosition Vec2
	Facing       float64
}

// BuffCommand はバフ適用処理のドメインコマンド。
type BuffCommand struct {
	CasterID  string
	RoomID    string
	TargetIDs []string
	Effect    BuffEffect
}

// AttackCommand は攻撃処理のドメインコマンド。
type AttackCommand struct {
	AttackerID        string
	TargetID          string
	RoomID            string
	SkillID           string
	Damage            int
	AdditionalEffects []string
}

// TradeCommand はトレード処理のドメインコマンド。
type TradeCommand struct {
	InitiatorID          string
	PartnerID            string
	RoomID               string
	Offer                []ItemChange
	Request              []ItemChange
	RequiresConfirmation bool
}
