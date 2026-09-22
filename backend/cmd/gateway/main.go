package main

import (
	"fmt"

	gatewayHTTP "github.com/cbe/kyc-biometric-gateway/internal/gateway/delivery/http"
	gatewayUsecase "github.com/cbe/kyc-biometric-gateway/internal/gateway/usecase"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/circuitbreaker"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/config"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/logger"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/redis"
	"github.com/gofiber/fiber/v2"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Environment)

	logger.Log.Info().Str("port", cfg.GatewayPort).Msg("Starting CBE High-Concurrency API Gateway")

	rdb, err := redis.NewRedisClient(cfg)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("Redis connection failed, continuing without distributed rate limiter")
	}

	cb := circuitbreaker.NewCircuitBreaker(cfg.CircuitBreakerThresh, 10*time.Second)

	app := fiber.New(fiber.Config{
		Prefork:               false,
		ServerHeader:          "CBE-KYC-Gateway",
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
	})

	uc := gatewayUsecase.NewGatewayUsecase(cfg, cb)
	handler := gatewayHTTP.NewGatewayHandler(cfg, uc)
	gatewayHTTP.SetupRouter(app, cfg, handler, rdb)

	addr := fmt.Sprintf(":%s", cfg.GatewayPort)
	if err := app.Listen(addr); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start API Gateway server")
	}
}
