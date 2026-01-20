package server

import (
	"net/http"

	"github.com/touka-aoi/paralle-vs-single/server/domain"
	"github.com/touka-aoi/paralle-vs-single/server/handler"
)

func Route(pubsub domain.PubSub, roomManager domain.RoomManager) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/ws", handler.NewAcceptHandler(pubsub, roomManager))
	return mux
}
