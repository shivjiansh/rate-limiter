package middleware

import (
	"net/http"
	"strconv"

	"github.com/example/go-rate-limiter/internal/service"
	"github.com/example/go-rate-limiter/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RateLimit(svc *service.RateLimiterService, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := utils.EnsureRequestID(c.GetHeader("X-Request-ID"))
		userID := c.GetHeader("X-User-ID")
		apiKey := c.GetHeader("X-API-Key")
		identifier := c.ClientIP()
		scope := service.ScopeIP
		if userID != "" {
			identifier = userID
			scope = service.ScopeUser
		} else if apiKey != "" {
			identifier = apiKey
			scope = service.ScopeAPIKey
		}

		allowed, remaining, err := svc.Allow(c.Request.Context(), service.Request{Scope: scope, Identifier: identifier, Endpoint: c.FullPath()})
		log.Info("rate_limit_check", zap.String("request_id", requestID), zap.String("user_id", userID), zap.String("path", c.FullPath()), zap.Bool("allowed", allowed), zap.Int("remaining", remaining))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "request_id": requestID})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded", "request_id": requestID})
			return
		}
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Next()
	}
}
