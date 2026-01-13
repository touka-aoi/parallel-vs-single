package server

import (
	"net/http"

	"github.com/touka-aoi/paralle-vs-single/server/handler"
)

func (s *Server) Route() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/ws", handler.NewAcceptHandler(s.dispatcher))
	return mux
}
