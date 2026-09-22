package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TotalRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyc_gateway_requests_total",
			Help: "Total number of HTTP requests processed by gateway",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kyc_gateway_request_duration_seconds",
			Help:    "Histogram of response latency for HTTP requests",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"method", "path"},
	)

	ActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kyc_gateway_active_requests",
			Help: "Current number of in-flight requests",
		},
	)

	BiometricVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "biometric_verifications_total",
			Help: "Total biometric verification requests processed",
		},
		[]string{"type", "status"},
	)

	KYCVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kyc_verifications_total",
			Help: "Total KYC verification requests processed",
		},
		[]string{"type", "status"},
	)
)

func PrometheusMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		ActiveRequests.Inc()
		defer ActiveRequests.Dec()

		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		duration := time.Since(start).Seconds()
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}

		TotalRequests.WithLabelValues(c.Method(), path, strconv.Itoa(status)).Inc()
		RequestDuration.WithLabelValues(c.Method(), path).Observe(duration)

		return err
	}
}
