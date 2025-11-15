package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/touka-aoi/paralle-vs-single/domain"
	"github.com/touka-aoi/paralle-vs-single/handler"
	"github.com/touka-aoi/paralle-vs-single/repository/state"
	"github.com/touka-aoi/paralle-vs-single/utils"
)

var (
	ErrInvalidPayload = errors.New("service: invalid payload")
)

type InteractionService struct {
	state   state.InteractionState
	metrics state.MetricsRecorder
}

func NewInteractionService(state state.InteractionState, metics state.MetricsRecorder) (*InteractionService, error) {
	return &InteractionService{
		state:   state,
		metrics: metics,
	}, nil
}

func (s *InteractionService) Move(ctx context.Context, payload *handler.MovePayload) (*domain.MoveResult, error) {
	start := time.Now()
	defer s.record("move", start)
	if err := s.validate(payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	return s.state.ApplyMove(ctx, &state.Move{
		RoomID: payload.RoomID,
		MoveCommand: domain.MoveCommand{
			UserID:       payload.Command.UserID,
			NextPosition: payload.Command.NextPosition,
			Facing:       payload.Command.Facing,
		},
	})
}

func (s *InteractionService) Buff(ctx context.Context, payload *handler.BuffPayload) (*domain.BuffResult, error) {
	start := time.Now()
	defer s.record("buff", start)
	if err := s.validate(payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	return s.state.ApplyBuff(ctx, &state.Buff{
		RoomID: payload.RoomID,
		BuffCommand: domain.BuffCommand{
			UserID:    payload.Command.UserID,
			TargetIDs: payload.Command.TargetIDs,
			Buff:      payload.Command.Buff,
		},
	})
}

func (s *InteractionService) Attack(ctx context.Context, payload *handler.AttackPayload) (*domain.AttackResult, error) {
	start := time.Now()
	defer s.record("attack", start)
	if err := s.validate(payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	return s.state.ApplyAttack(ctx, &state.Attack{
		RoomID: payload.RoomID,
		AttackCommand: domain.AttackCommand{
			UserID:   payload.Command.UserID,
			TargetID: payload.Command.TargetID,
			Damage:   payload.Command.Damage,
		},
	})
}

func (s *InteractionService) Trade(ctx context.Context, payload *handler.TradePayload) (*domain.TradeResult, error) {
	start := time.Now()
	defer s.record("trade", start)
	if err := s.validate(payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	return s.state.ApplyTrade(ctx, &state.Trade{
		RoomID: payload.RoomID,
		TradeCommand: domain.TradeCommand{
			UserID:               payload.Command.UserID,
			PartnerID:            payload.Command.PartnerID,
			Offer:                payload.Command.Offer,
			Request:              payload.Command.Request,
			RequiresConfirmation: payload.Command.RequiresConfirmation,
		},
	})
}

func (s *InteractionService) record(endpoint string, started time.Time) {
	duration := time.Since(started)
	ctx := context.Background()
	s.metrics.RecordLatency(ctx, endpoint, duration)
	s.metrics.IncrementCounter(ctx, "requests."+endpoint, 1)
}

func (s *InteractionService) validate(payload utils.Validator) error {
	return payload.Validate()
}
