package handlers

import (
	"net/http"

	"github.com/liwaisi-tech/iot-server-smart-irrigation/backend/go-soc-consumer/internal/usecases/ping"
)

type PingHandler struct {
	pingUseCase ping.PingUseCase
}

func NewPingHandler(pingUseCase ping.PingUseCase) *PingHandler {
	return &PingHandler{
		pingUseCase: pingUseCase,
	}
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	response := h.pingUseCase.Ping(ctx)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(response)); err != nil {
		// Log the error if we can't write the response
		// In a real application, you might want to use a proper logger here
		http.Error(w, "failed to write response", http.StatusInternalServerError)
		return
	}
}
