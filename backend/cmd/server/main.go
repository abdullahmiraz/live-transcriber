// Command server is the entrypoint and composition root: it wires config, logging,
// metrics, the database, domain services, providers, the WS hub, and the HTTP server.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"meetingplatform/internal/chat"
	"meetingplatform/internal/config"
	"meetingplatform/internal/httpapi"
	"meetingplatform/internal/meeting"
	"meetingplatform/internal/observability"
	"meetingplatform/internal/pubsub"
	"meetingplatform/internal/storage/postgres"
	"meetingplatform/internal/transcription"
	"meetingplatform/internal/translation"
	"meetingplatform/internal/ws"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.LogLevel)
	metrics := observability.NewMetrics()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- Database ---
	logger.Info("connecting to database")
	pool, err := postgres.ConnectWithRetry(rootCtx, cfg.DatabaseURL, 60*time.Second)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("running migrations")
	if err := postgres.Migrate(rootCtx, pool); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	// --- Realtime broker (Redis in containers, in-memory locally) ---
	broker, err := pubsub.New(rootCtx, cfg.RedisURL)
	if err != nil {
		logger.Error("broker init failed", "error", err)
		os.Exit(1)
	}
	defer broker.Close()
	logger.Info("realtime broker initialized", "broker", broker.Name())

	// --- Domain services ---
	meetingRepo := postgres.NewMeetingRepo(pool)
	meetingSvc := meeting.NewService(meetingRepo)
	messageRepo := postgres.NewMessageRepo(pool)
	chatSvc := chat.NewService(messageRepo, broker, logger)

	// --- AI providers (swappable via env) ---
	stt, err := transcription.New(cfg.STTProvider)
	if err != nil {
		logger.Error("invalid STT provider", "error", err)
		os.Exit(1)
	}
	tr, err := translation.New(cfg.TranslationProvider)
	if err != nil {
		logger.Error("invalid translation provider", "error", err)
		os.Exit(1)
	}
	logger.Info("ai providers initialized", "stt", stt.Name(), "translation", tr.Name())

	// --- Realtime hub ---
	hub := ws.NewHub(stt, tr, chatSvc, meetingSvc, broker, cfg.DefaultTargetLang, metrics, logger)
	go hub.Run(rootCtx)
	wsHandler := ws.NewHandler(hub, meetingSvc, cfg.CORSOrigins)

	// --- HTTP server ---
	router := httpapi.NewRouter(httpapi.Deps{
		Cfg:      cfg,
		Logger:   logger,
		Metrics:  metrics,
		Meetings: meetingSvc,
		Chat:     chatSvc,
		Ready: func(ctx context.Context) error {
			return pool.Ping(ctx)
		},
		WSHandler: wsHandler,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			stop()
		}
	}()

	<-rootCtx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("stopped")
}
