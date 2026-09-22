package http

import (
	"github.com/cbe/kyc-biometric-gateway/internal/platform/config"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/metrics"
	"github.com/cbe/kyc-biometric-gateway/internal/platform/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func SetupRouter(app *fiber.App, cfg *config.Config, handler *GatewayHandler, rdb *redis.Client) {
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Correlation-ID",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(middleware.RequestID())
	app.Use(metrics.PrometheusMiddleware())

	app.Get("/health", handler.HealthCheck)
	app.Get("/metrics", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())(c.Context())
		return nil
	})

	v1 := app.Group("/api/v1")
	v1.Post("/auth/login", handler.Login)

	protected := v1.Group("")
	protected.Use(middleware.RedisRateLimiter(rdb, cfg.RateLimitRequests, cfg.RateLimitWindowSec))
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))

	protected.Post("/onboarding/full", handler.ProcessFullOnboarding)
	protected.All("/kyc/*", handler.ProxyToKYC)
	protected.All("/biometric/*", handler.ProxyToBiometric)
}
