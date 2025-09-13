package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sandbox-dev-test-task/internal/config"
	"sandbox-dev-test-task/internal/http-server"
	"sandbox-dev-test-task/internal/pool"
	"syscall"
)

func main() {

	cfg := config.MustLoad()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	p := pool.New(cfg.WorkersNum, pool.WithTaskBufferSize(cfg.QueueSize))

	serv := http_server.Init(log, p)

	closeConnCh := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		p.Stop()

		if err := serv.Shutdown(context.Background()); err != nil {
			log.Error("server shutdown error: %v", "error message", err)
		}

		_ = p.Stop()

		close(closeConnCh)
	}()

	log.Info("server started on %s", "port", serv.Addr)
	if err := serv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("server error: %v", "error message", err)
	}

	<-closeConnCh
	log.Info("server stopped")
}
