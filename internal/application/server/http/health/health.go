package health

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
}

func New() *Server {
	return &Server{}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.healthHandler)
	mux.Handle("/metrics", promhttp.Handler())
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}
