package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/api"
	"server/auth"
	"server/config"
	"server/database"
	"server/logging"
	"server/repository"
	"server/services"
	"server/websocket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		println("failed to load configuration: " + err.Error())
		os.Exit(1)
	}

	logger := logging.New(cfg.LogLevel)
	logger.Info("starting Corvus server",
		"environment", cfg.Environment,
		"port", cfg.HTTPPort,
		"db", cfg.DatabasePath,
	)

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	if err := database.MigrateDB(db); err != nil {
		logger.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("database migrations complete")

	repos := repository.New(db)
	svcs := services.New(repos, cfg.ChatRequestCooldown)
	authSvc := auth.NewService(repos.Users, cfg.JWTSecret, cfg.JWTExpiration)

	wsServer := websocket.NewServer(repos, logger)
	router := api.NewRouter(logger, authSvc, svcs, wsServer, cfg.CORSOrigin, wsServer)

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownDone := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
		close(shutdownDone)
	}()

	logger.Info("HTTP server listening", "addr", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}

	<-shutdownDone
	logger.Info("server exited cleanly")
}
