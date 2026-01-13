package server

import (
	"context"
	"net/http"

	"github.com/touka-aoi/paralle-vs-single/server/domain"
)

type Server struct {
	HTTP       *http.Server
	dispatcher domain.Dispatcher
}

func NewServer(addr string, mux *http.ServeMux, dispatcher domain.Dispatcher) domain.Server {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return &Server{
		HTTP:       httpServer,
		dispatcher: dispatcher,
	}
}

func (s *Server) Serve() error                       { return s.HTTP.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.HTTP.Shutdown(ctx) }
func (s *Server) Close() error                       { return s.HTTP.Close() }
func (s *Server) Addr() string                       { return s.HTTP.Addr }
