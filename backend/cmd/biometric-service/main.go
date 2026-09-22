package main

import (
	"fmt"

	biometricHTTP "github.com/cbe/kyc-biometric-gateway/internal/biometric/delivery/http"
	biometricRepo "github.com/cbe/kyc-biometric-gateway/internal/biometric/repository"
	biometricUsecase "github.com/cbe/kyc-biometric-gateway/internal/biometric/usecase"
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

	logger.Log.Info().Str("port", cfg.BiometricPort).Msg("Starting Biometric Microservice Engine")

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("Database connection warning in Biometric Service")
	}

	wp := workerpool.NewWorkerPool(cfg.MaxWorkerPoolSize, cfg.MaxQueueCapacity)
	wp.Start()
	defer wp.Stop()

	repo := biometricRepo.NewBiometricRepository(db)
	uc := biometricUsecase.NewBiometricUsecase(repo, wp)

	app := fiber.New(fiber.Config{
		ServerHeader:          "CBE-Biometric-Engine",
		DisableStartupMessage: true,
	})
	app.Use(recover.New())

	biometricHTTP.NewBiometricHandler(app, uc)

	addr := fmt.Sprintf(":%s", cfg.BiometricPort)
	if err := app.Listen(addr); err != nil {
		logger.Log.Fatal().Err(err).Msg("Biometric Service failed to start")
	}
}
