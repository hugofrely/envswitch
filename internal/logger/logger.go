package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hugofrely/envswitch/internal/config"
)

// LogLevel represents logging levels
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger handles application logging
type Logger struct {
	level      LogLevel
	file       *os.File
	showColors bool
	showTime   bool
}

var (
	globalLogger *Logger
)

// InitLogger initializes the global logger from config
func InitLogger(cfg *config.Config) error {
	level := parseLogLevel(cfg.LogLevel)

	var file *os.File
	var err error

	if cfg.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(cfg.LogFile)
		if mkdirErr := os.MkdirAll(logDir, 0755); mkdirErr != nil {
			return fmt.Errorf("failed to create log directory: %w", mkdirErr)
		}

		// Open log file in append mode
		file, err = os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
	}

	globalLogger = &Logger{
		level:      level,
		file:       file,
		showColors: cfg.ColorOutput,
		showTime:   cfg.ShowTimestamps,
	}

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Return a default logger if not initialized
		globalLogger = &Logger{
			level:      LevelInfo,
			showColors: true,
			showTime:   true,
		}
	}
	return globalLogger
}

// Close closes the log file if open
func Close() error {
	if globalLogger != nil && globalLogger.file != nil {
		return globalLogger.file.Close()
	}
	return nil
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	GetLogger().log(LevelDebug, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	GetLogger().log(LevelInfo, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	GetLogger().log(LevelWarn, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	GetLogger().log(LevelError, format, args...)
}

// log performs the actual logging
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := ""

	if l.showTime {
		timestamp = time.Now().Format("2006-01-02 15:04:05") + " "
	}

	levelStr := levelString(level, l.showColors)
	output := fmt.Sprintf("%s%s %s\n", timestamp, levelStr, msg)

	// Write to stdout/stderr
	writer := l.getWriter(level)
	fmt.Fprint(writer, output)

	// Write to file if configured
	if l.file != nil {
		// Strip colors for file output
		fileOutput := fmt.Sprintf("%s%s %s\n", timestamp, levelStringPlain(level), msg)
		l.file.WriteString(fileOutput)
	}
}

// getWriter returns the appropriate output writer for the log level
func (l *Logger) getWriter(level LogLevel) io.Writer {
	if level >= LevelWarn {
		return os.Stderr
	}
	return os.Stdout
}

// levelString returns a colored level string
func levelString(level LogLevel, useColor bool) string {
	if !useColor {
		return levelStringPlain(level)
	}

	switch level {
	case LevelDebug:
		return "\033[36m[DEBUG]\033[0m" // Cyan
	case LevelInfo:
		return "\033[32m[INFO]\033[0m" // Green
	case LevelWarn:
		return "\033[33m[WARN]\033[0m" // Yellow
	case LevelError:
		return "\033[31m[ERROR]\033[0m" // Red
	default:
		return "[UNKNOWN]"
	}
}

// levelStringPlain returns a plain level string
func levelStringPlain(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "[DEBUG]"
	case LevelInfo:
		return "[INFO]"
	case LevelWarn:
		return "[WARN]"
	case LevelError:
		return "[ERROR]"
	default:
		return "[UNKNOWN]"
	}
}

// parseLogLevel converts a string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Colorize returns a colored string if colors are enabled
func (l *Logger) Colorize(color, text string) string {
	if !l.showColors {
		return text
	}

	colors := map[string]string{
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
		"reset":   "\033[0m",
	}

	if code, ok := colors[color]; ok {
		return code + text + colors["reset"]
	}
	return text
}

// ShouldShowColors returns whether colors should be shown
func (l *Logger) ShouldShowColors() bool {
	return l.showColors
}
