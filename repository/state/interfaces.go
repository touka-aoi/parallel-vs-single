package state

import (
	"context"
	"time"

	"github.com/touka-aoi/paralle-vs-single/domain"
)

type Move struct {
	RoomID string
	domain.MoveCommand
}

type Buff struct {
	RoomID string
	domain.BuffCommand
}

type Attack struct {
	RoomID string
	domain.AttackCommand
}

type Trade struct {
	RoomID string
	domain.TradeCommand
}

type InteractionState interface {
	RegisterPlayer(ctx context.Context, playerID string, roomID string) error
	ApplyMove(ctx context.Context, cmd *Move) (*domain.MoveResult, error)
	ApplyAttack(ctx context.Context, cmd *Attack) (*domain.AttackResult, error)
}

type MetricsRecorder interface {
	RecordLatency(ctx context.Context, endpoint string, duration time.Duration)
	RecordContention(ctx context.Context, endpoint string, wait time.Duration)
	IncrementCounter(ctx context.Context, name string, delta int)
}
