package handlers

import (
	"net/http"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
)

type PingHandler struct {
	pingUseCase ping.UseCase
}

func NewPingHandler(pingUseCase ping.UseCase) *PingHandler {
	return &PingHandler{
		pingUseCase: pingUseCase,
	}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	response := h.pingUseCase.Execute(ctx)
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}