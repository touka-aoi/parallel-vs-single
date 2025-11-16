package parallel

import (
	"encoding/json"
	"net/http"

	"github.com/touka-aoi/paralle-vs-single/service"
)

type ConnectResponse struct {
	PlayerID string `json:"player_id"`
	RoomID   string `json:"room_id"`
}

type ConnectHandler struct {
	svc service.InteractionService
}

func NewConnectHandler(svc service.InteractionService) *ConnectHandler {
	return &ConnectHandler{svc: svc}
}

func (h *ConnectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	playerID, roomID, err := h.svc.Connect(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := ConnectResponse{
		PlayerID: playerID,
		RoomID:   roomID,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
	return
}
