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
	ApplyMove(ctx context.Context, cmd *Move) (*domain.MoveResult, error)
	ApplyBuff(ctx context.Context, cmd *Buff) (*domain.BuffResult, error)
	ApplyAttack(ctx context.Context, cmd *Attack) (*domain.AttackResult, error)
	ApplyTrade(ctx context.Context, cmd *Trade) (*domain.TradeResult, error)
}

type MetricsRecorder interface {
	RecordLatency(ctx context.Context, endpoint string, duration time.Duration)
	RecordContention(ctx context.Context, endpoint string, wait time.Duration)
	IncrementCounter(ctx context.Context, name string, delta int)
}
