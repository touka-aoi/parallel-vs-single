package handler

import (
	"errors"
	"fmt"

	"github.com/touka-aoi/paralle-vs-single/application/domain"
	"github.com/touka-aoi/paralle-vs-single/utils"
)

type MovePayload struct {
	Meta    domain.Meta
	RoomID  string
	Command domain.MoveCommand
}

type BuffPayload struct {
	Meta    domain.Meta
	RoomID  string
	Command domain.BuffCommand
}

type AttackPayload struct {
	Meta    domain.Meta
	RoomID  string
	Command domain.AttackCommand
}

type TradePayload struct {
	Meta    domain.Meta
	RoomID  string
	Command domain.TradeCommand
}

func (p *MovePayload) Validate() error {
	cmd := p.Command
	if cmd.UserID == "" {
		return errors.New("actor id is required")
	}
	if p.RoomID == "" {
		return errors.New("room id is required")
	}
	if !utils.FiniteVec(cmd.NextPosition) {
		return fmt.Errorf("invalid position: %+v", cmd.NextPosition)
	}
	return nil
}

func (p *BuffPayload) Validate() error {
	cmd := p.Command
	if cmd.UserID == "" {
		return errors.New("caster id is required")
	}
	if p.RoomID == "" {
		return errors.New("room id is required")
	}
	if cmd.Buff.BuffID == "" {
		return errors.New("effect id is required")
	}
	if cmd.Buff.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	return nil
}

func (p *AttackPayload) Validate() error {
	cmd := p.Command
	if cmd.UserID == "" || cmd.TargetID == "" {
		return errors.New("attacker and target ids are required")
	}
	if p.RoomID == "" {
		return errors.New("room id is required")
	}
	if cmd.Damage <= 0 {
		return errors.New("damage must be positive")
	}
	return nil
}

func (p *TradePayload) Validate() error {
	cmd := p.Command
	if cmd.UserID == "" || cmd.PartnerID == "" {
		return errors.New("initiator and partner ids are required")
	}
	if p.RoomID == "" {
		return errors.New("room id is required")
	}
	if len(cmd.Offer) == 0 && len(cmd.Request) == 0 {
		return errors.New("either offer or request must be present")
	}
	return nil
}
