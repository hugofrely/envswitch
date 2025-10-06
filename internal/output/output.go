package output

import (
	"fmt"
	"time"

	"github.com/hugofrely/envswitch/internal/config"
)

// Formatter handles output formatting based on config
type Formatter struct {
	useColors      bool
	showTimestamps bool
}

var globalFormatter *Formatter

// InitFormatter initializes the global formatter from config
func InitFormatter(cfg *config.Config) {
	globalFormatter = &Formatter{
		useColors:      cfg.ColorOutput,
		showTimestamps: cfg.ShowTimestamps,
	}
}

// GetFormatter returns the global formatter
func GetFormatter() *Formatter {
	if globalFormatter == nil {
		globalFormatter = &Formatter{
			useColors:      true,
			showTimestamps: true,
		}
	}
	return globalFormatter
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f := GetFormatter()
	if f.useColors {
		fmt.Printf("✅ %s\n", msg)
	} else {
		fmt.Printf("[OK] %s\n", msg)
	}
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f := GetFormatter()
	if f.useColors {
		fmt.Printf("❌ %s\n", msg)
	} else {
		fmt.Printf("[ERROR] %s\n", msg)
	}
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f := GetFormatter()
	if f.useColors {
		fmt.Printf("⚠️  %s\n", msg)
	} else {
		fmt.Printf("[WARN] %s\n", msg)
	}
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f := GetFormatter()
	if f.useColors {
		fmt.Printf("ℹ️  %s\n", msg)
	} else {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

// Progress prints a progress message
func Progress(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	f := GetFormatter()
	if f.useColors {
		fmt.Printf("🔄 %s\n", msg)
	} else {
		fmt.Printf("[PROGRESS] %s\n", msg)
	}
}

// Colorize returns a colored string if colors are enabled
func Colorize(color, text string) string {
	f := GetFormatter()
	if !f.useColors {
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
		"bold":    "\033[1m",
		"reset":   "\033[0m",
	}

	if code, ok := colors[color]; ok {
		return code + text + colors["reset"]
	}
	return text
}

// WithTimestamp adds a timestamp prefix if enabled
func WithTimestamp(msg string) string {
	f := GetFormatter()
	if f.showTimestamps {
		timestamp := time.Now().Format("15:04:05")
		return fmt.Sprintf("[%s] %s", timestamp, msg)
	}
	return msg
}

// Printf prints formatted output with color and timestamp support
func Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Print(WithTimestamp(msg))
}

// Println prints a line with color and timestamp support
func Println(msg string) {
	fmt.Println(WithTimestamp(msg))
}
