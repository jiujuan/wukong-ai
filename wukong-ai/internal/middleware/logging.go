package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		logger.Info(
			"http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

func RecoveryLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				file, line, code := panicLocation()
				stack := string(debug.Stack())

				logger.Error(
					"panic_recovered",
					"panic", fmt.Sprint(recovered),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"query", c.Request.URL.RawQuery,
					"client_ip", c.ClientIP(),
					"error_file", file,
					"error_line", line,
					"error_code", code,
					"runtime_stack", stack,
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

func panicLocation() (string, int, string) {
	callers := make([]uintptr, 64)
	n := runtime.Callers(3, callers)
	frames := runtime.CallersFrames(callers[:n])

	for {
		frame, more := frames.Next()
		file := filepathToSlash(frame.File)
		if !strings.Contains(file, "/runtime/") &&
			!strings.Contains(file, "/gin-gonic/") &&
			!strings.HasSuffix(file, "/internal/middleware/logging.go") &&
			frame.File != "" && frame.Line > 0 {
			return frame.File, frame.Line, readLine(frame.File, frame.Line)
		}

		if !more {
			break
		}
	}

	return "", 0, ""
}

func readLine(path string, target int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	index := target - 1
	if index < 0 || index >= len(lines) {
		return ""
	}

	return strings.TrimSpace(lines[index])
}

func filepathToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
