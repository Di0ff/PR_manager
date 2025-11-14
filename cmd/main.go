package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"mPR/db/migrations"
	"mPR/internal/api/handlers"
	"mPR/internal/api/routers"
	"mPR/internal/config"
	"mPR/internal/logger"
	"mPR/internal/service"
	"mPR/internal/storage/postgres"
	"mPR/internal/storage/repository"
)

func main() {
	cfg := config.Load()

	log := logger.New(*cfg)
	defer func() { _ = log.Sync() }()

	migrations.Run(cfg.Postgres, log)

	db := postgres.New(cfg.Postgres, log)

	repos := repository.New(db)
	services := service.New(repos, cfg.App.MaxReviewers)
	api := handlers.New(log, services)
	router := routers.Init(api, cfg.App.AdminToken)

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Сервис упал", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Ошибка завершения работы сервиса", zap.Error(err))
	}
}
