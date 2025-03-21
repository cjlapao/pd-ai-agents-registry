package middleware

import (
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/logger"
	"github.com/gin-gonic/gin"
)

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Infow("Request",
			"method", method,
			"path", path,
			"status", statusCode,
			"latency", latency,
			"ip", c.ClientIP(),
		)
	}
}
