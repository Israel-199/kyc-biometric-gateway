package main

import (
	"fmt"

	kycHTTP "github.com/cbe/kyc-biometric-gateway/internal/kyc/delivery/http"
	kycRepo "github.com/cbe/kyc-biometric-gateway/internal/kyc/repository"
	kycUsecase "github.com/cbe/kyc-biometric-gateway/internal/kyc/usecase"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/config"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/database"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/logger"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/workerpool"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg := config.LoadConfig()
	logger.InitLogger(cfg.Environment)

	logger.Log.Info().Str("port", cfg.KYCPort).Msg("Starting KYC Microservice Engine")

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("Database connection warning in KYC Service")
	}

	wp := workerpool.NewWorkerPool(cfg.MaxWorkerPoolSize, cfg.MaxQueueCapacity)
	wp.Start()
	defer wp.Stop()

	repo := kycRepo.NewKYCRepository(db)
	uc := kycUsecase.NewKYCUsecase(repo, wp)

	app := fiber.New(fiber.Config{
		ServerHeader:          "CBE-KYC-Engine",
		DisableStartupMessage: true,
	})
	app.Use(recover.New())

	kycHTTP.NewKYCHandler(app, uc)

	addr := fmt.Sprintf(":%s", cfg.KYCPort)
	if err := app.Listen(addr); err != nil {
		logger.Log.Fatal().Err(err).Msg("KYC Service failed to start")
	}
}
