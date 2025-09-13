package http_server

import (
	"log/slog"
	"net/http"
	"sandbox-dev-test-task/internal/http-server/middleware/logger"
	"sandbox-dev-test-task/internal/pool"
)

func Init(log *slog.Logger, p pool.Pool) http.Server {
	mux := http.NewServeMux()

	mux.Handle("POST /enqueue", enqueuePostHandler(p))
	mux.Handle("GET /healthcheck", http.HandlerFunc(healthCheckHandler))

	server := http.Server{
		Addr:    ":8080",
		Handler: logger.New(log)(mux),
	}

	return server
}
