package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/touka-aoi/paralle-vs-single/handler"
	"github.com/touka-aoi/paralle-vs-single/repository/state"
)

type ConnectService interface {
	Connect(ctx context.Context) (handler.ClientID, handler.RoomID, error)
}

type connectService struct {
	state state.InteractionState
}

func NewConnectService(state state.InteractionState) (ConnectService, error) {
	return &connectService{state: state}, nil
}

func (c connectService) Connect(ctx context.Context) (handler.ClientID, handler.RoomID, error) {
	clientID := handler.ClientID(uuid.NewString())
	roomID := handler.RoomID(uuid.NewString())
	slog.InfoContext(ctx, "player connected")
	return clientID, roomID, nil
}
