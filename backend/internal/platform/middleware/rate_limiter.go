package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cbe/kyc-biometric-gateway/pkg/response"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func RedisRateLimiter(rdb *redis.Client, limit int, windowSec int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rdb == nil {
			return c.Next()
		}

		clientIP := c.IP()
		key := fmt.Sprintf("ratelimit:%s:%d", clientIP, time.Now().Unix()/int64(windowSec))

		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}

		if count == 1 {
			rdb.Expire(ctx, key, time.Duration(windowSec)*time.Second)
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, limit-int(count))))

		if int(count) > limit {
			return response.Error(c, fiber.StatusTooManyRequests, "Rate limit exceeded. Try again later.", nil)
		}

		return c.Next()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
