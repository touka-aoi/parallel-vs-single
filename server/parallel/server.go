package parallel

import "net/http"

func NewServer(addr string, wsHandler http.Handler, connectHandler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /ws", wsHandler)
	mux.Handle("POST /connect", connectHandler)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
