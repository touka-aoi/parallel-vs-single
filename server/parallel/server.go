package parallel

import "net/http"

func NewServer(addr string, wsHandler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/ws", wsHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
