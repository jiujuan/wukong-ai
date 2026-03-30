package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	globalLogger *slog.Logger
	once         sync.Once
)

// Init 根据配置初始化全局 slog Logger，支持 debug/info/warn/error 四级
func Init(level string) {
	once.Do(func() {
		var l slog.Level
		switch strings.ToLower(level) {
		case "debug":
			l = slog.LevelDebug
		case "warn":
			l = slog.LevelWarn
		case "error":
			l = slog.LevelError
		default:
			l = slog.LevelInfo
		}

		opts := &slog.HandlerOptions{
			Level: l,
		}

		// 生产环境使用 JSON 格式
		handler := slog.NewJSONHandler(os.Stdout, opts)
		globalLogger = slog.New(handler)
		slog.SetDefault(globalLogger)
	})
}

// Get 获取全局 Logger 实例
func Get() *slog.Logger {
	if globalLogger == nil {
		Init("info")
	}
	return globalLogger
}

// Debug 记录调试级别日志
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info 记录信息级别日志
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn 记录警告级别日志
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error 记录错误级别日志
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With 返回带有额外字段的 Logger
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}
