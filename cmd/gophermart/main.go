package main

import (
	"log"
	"net/http"
	"time"

	"go-gophermart/internal/config"
	"go-gophermart/internal/database"
	"go-gophermart/internal/handler"
	"go-gophermart/internal/logger"
	"go-gophermart/internal/repository"
	"go-gophermart/internal/router"
	"go-gophermart/internal/service"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLogger, err := logger.New()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	db, err := database.New(cfg.DatabaseDSN)
	if err != nil {
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	userRepo := repository.NewUserPostgresRepository(db.DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, cfg.CookieSecret, cfg.CookieSecure, zapLogger)

	orderRepo := repository.NewOrderRepository(db.DB)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService, zapLogger)

	balanceRepo := repository.NewBalanceRepository(db.DB)
	balanceService := service.NewBalanceService(balanceRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService, zapLogger)

	router := router.New(router.RouterDeps{
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		BalanceHandler: balanceHandler,
		CookieSecret:   cfg.CookieSecret,
		Logger:         zapLogger,
	})

	srv := &http.Server{
		Addr:         cfg.RunAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	zapLogger.Info("Server running at",
		zap.String("address", cfg.RunAddr),
		zap.Bool("cookie_secure", cfg.CookieSecure),
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zapLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
