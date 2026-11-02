package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/luiggisao/auto-golpe/internal/config"
)

type Server struct {
	cfg         *config.Config
	queueClient *asynq.Client
}

type HealthResponse struct {
	Status    string    `json:"status"`
	App       string    `json:"app"`
	Timestamp time.Time `json:"timestamp"`
}

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Received  string    `json:"received"`
	Timestamp time.Time `json:"timestamp"`
}

func NewRouter(cfg *config.Config, queueClient *asynq.Client) http.Handler {
	mux := http.NewServeMux()
	s := &Server{cfg: cfg, queueClient: queueClient}

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/v1/echo", handleEcho)
	mux.HandleFunc("GET /api/v1/panic", handlePanic)
	mux.HandleFunc("POST /api/v1/webhooks/slack", s.handleSlackWebhook)

	var handler http.Handler = mux
	handler = Logging(handler)
	handler = Recovery(handler)

	return handler
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "ok",
		App:       "golpe",
		Timestamp: time.Now().UTC(),
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	var req EchoRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON body",
		})
		return
	}

	if req.Message == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"error": "field 'message' is required",
		})
		return
	}

	resp := EchoResponse{
		Received:  req.Message,
		Timestamp: time.Now().UTC(),
	}

	writeJSON(w, http.StatusOK, resp)
}

func handlePanic(w http.ResponseWriter, r *http.Request) {
	panic("unexpected simulated disaster in handler")
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
