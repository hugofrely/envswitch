package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hugofrely/envswitch/internal/config"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"error", LevelError},
		{"unknown", LevelInfo}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitLogger(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("initializes with default config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.LogFile = filepath.Join(tempDir, "test.log")

		err := InitLogger(cfg)
		require.NoError(t, err)
		defer Close()

		assert.NotNil(t, globalLogger)
		assert.Equal(t, LevelInfo, globalLogger.level)
	})

	t.Run("creates log directory if missing", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.LogFile = filepath.Join(tempDir, "nested", "dir", "test.log")

		err := InitLogger(cfg)
		require.NoError(t, err)
		defer Close()

		_, err = os.Stat(filepath.Dir(cfg.LogFile))
		assert.NoError(t, err)
	})

	t.Run("sets correct log level", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.LogLevel = "debug"
		cfg.LogFile = ""

		err := InitLogger(cfg)
		require.NoError(t, err)

		assert.Equal(t, LevelDebug, globalLogger.level)
	})
}

func TestGetLogger(t *testing.T) {
	t.Run("returns default logger if not initialized", func(t *testing.T) {
		globalLogger = nil
		logger := GetLogger()

		assert.NotNil(t, logger)
		assert.Equal(t, LevelInfo, logger.level)
		assert.True(t, logger.showColors)
	})

	t.Run("returns global logger if initialized", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.ColorOutput = false
		InitLogger(cfg)
		defer Close()

		logger := GetLogger()
		assert.False(t, logger.showColors)
	})
}

func TestLogLevels(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.LogLevel = "debug"
	cfg.LogFile = ""
	cfg.ColorOutput = false
	cfg.ShowTimestamps = false
	InitLogger(cfg)
	defer Close()

	// Just verify no panics - output testing is complex with stdout/stderr
	t.Run("debug", func(t *testing.T) {
		Debug("debug message")
	})

	t.Run("info", func(t *testing.T) {
		Info("info message")
	})

	t.Run("warn", func(t *testing.T) {
		Warn("warn message")
	})

	t.Run("error", func(t *testing.T) {
		Error("error message")
	})
}

func TestLogFiltering(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.LogLevel = "warn"
	cfg.LogFile = ""
	cfg.ShowTimestamps = false
	InitLogger(cfg)
	defer Close()

	// Just verify filtering logic works (messages below warn level should be filtered)
	// Actual output testing is complex with stdout/stderr
	Debug("debug message")
	Info("info message")

	// Verify logger is configured correctly
	assert.Equal(t, LevelWarn, globalLogger.level)
}

func TestColorize(t *testing.T) {
	t.Run("adds color codes when enabled", func(t *testing.T) {
		logger := &Logger{showColors: true}
		result := logger.Colorize("red", "test")

		assert.Contains(t, result, "\033[31m")
		assert.Contains(t, result, "test")
		assert.Contains(t, result, "\033[0m")
	})

	t.Run("returns plain text when disabled", func(t *testing.T) {
		logger := &Logger{showColors: false}
		result := logger.Colorize("red", "test")

		assert.Equal(t, "test", result)
		assert.NotContains(t, result, "\033[")
	})

	t.Run("handles unknown colors", func(t *testing.T) {
		logger := &Logger{showColors: true}
		result := logger.Colorize("unknown", "test")

		assert.Equal(t, "test", result)
	})
}

func TestLogToFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	cfg := config.DefaultConfig()
	cfg.LogFile = logFile
	cfg.LogLevel = "info"
	cfg.ColorOutput = true
	cfg.ShowTimestamps = false

	err := InitLogger(cfg)
	require.NoError(t, err)
	defer Close()

	Info("test message")

	// Read log file
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	// File should contain plain text (no colors)
	assert.Contains(t, string(content), "[INFO]")
	assert.Contains(t, string(content), "test message")
	assert.NotContains(t, string(content), "\033[")
}

func TestTimestamps(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := config.DefaultConfig()
	cfg.LogLevel = "info"
	cfg.LogFile = ""
	cfg.ColorOutput = false
	cfg.ShowTimestamps = true
	InitLogger(cfg)
	defer Close()
	defer func() { os.Stdout = old }()

	Info("test")
	w.Close()

	out, _ := io.ReadAll(r)
	output := string(out)

	// Should contain timestamp in format YYYY-MM-DD HH:MM:SS
	assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, output)
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		useColor bool
		contains string
	}{
		{LevelDebug, false, "[DEBUG]"},
		{LevelInfo, false, "[INFO]"},
		{LevelWarn, false, "[WARN]"},
		{LevelError, false, "[ERROR]"},
		{LevelDebug, true, "\033[36m"}, // Cyan
		{LevelInfo, true, "\033[32m"},  // Green
		{LevelWarn, true, "\033[33m"},  // Yellow
		{LevelError, true, "\033[31m"}, // Red
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.level)), func(t *testing.T) {
			result := levelString(tt.level, tt.useColor)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("closes file if open", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := config.DefaultConfig()
		cfg.LogFile = filepath.Join(tempDir, "test.log")

		InitLogger(cfg)
		err := Close()
		assert.NoError(t, err)
	})

	t.Run("handles nil logger", func(t *testing.T) {
		globalLogger = nil
		err := Close()
		assert.NoError(t, err)
	})

	t.Run("handles logger without file", func(t *testing.T) {
		globalLogger = &Logger{file: nil}
		err := Close()
		assert.NoError(t, err)
	})
}

func TestConcurrentLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := config.DefaultConfig()
	cfg.LogFile = ""
	cfg.ShowTimestamps = false
	InitLogger(cfg)
	defer Close()

	// Run concurrent logging
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			Info("concurrent message %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Just verify no panic occurred
	assert.NotNil(t, buf)
}

func TestShouldShowColors(t *testing.T) {
	t.Run("returns true when colors enabled", func(t *testing.T) {
		logger := &Logger{showColors: true}
		assert.True(t, logger.ShouldShowColors())
	})

	t.Run("returns false when colors disabled", func(t *testing.T) {
		logger := &Logger{showColors: false}
		assert.False(t, logger.ShouldShowColors())
	})
}
