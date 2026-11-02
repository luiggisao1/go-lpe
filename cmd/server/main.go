package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/luiggisao/auto-golpe/internal/config"
	"github.com/luiggisao/auto-golpe/internal/gemini"
	apphttp "github.com/luiggisao/auto-golpe/internal/http"
	"github.com/luiggisao/auto-golpe/internal/queue"
	"github.com/luiggisao/auto-golpe/internal/slack"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()
	queueClient := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisAddr})
	defer func(queueClient *asynq.Client) {
		err := queueClient.Close()
		if err != nil {
			slog.Error("failed to close queue client", "error", err)
		}
	}(queueClient)

	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{Concurrency: 5},
	)

	geminiClient, err := gemini.NewClient(context.Background(), cfg.GoogleApiKey)
	if err != nil {
		slog.Error("failed to create gemini client", "error", err)
		os.Exit(1)
	}

	slackClient := slack.NewClient(cfg.SlackBotToken)

	workerMux := asynq.NewServeMux()
	workerMux.HandleFunc(queue.TypeSlackProcessEvent, func(ctx context.Context, t *asynq.Task) error {
		return queue.HandleSlackProcessEventTask(ctx, t, geminiClient, slackClient)
	})

	go func() {
		slog.Info("starting Asynq worker")
		if err := asynqServer.Run(workerMux); err != nil {
			slog.Error("asynq worker failed to start", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("configuration loaded", "env", cfg.Environment, "port", cfg.Port)

	router := apphttp.NewRouter(cfg, queueClient)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting HTTP server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	slog.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited cleanly")
	asynqServer.Shutdown()
}
