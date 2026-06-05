// Dev mode logging. Activated via GDA_DEV=1 environment variable.
// Writes timestamped logs to stderr for debugging user issues.
package devlog

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	if os.Getenv("GDA_DEV") == "1" {
		logger = log.New(os.Stderr, "", 0)
		Printf("dev mode enabled")
	}
}

// Enabled reports whether dev logging is active.
func Enabled() bool {
	return logger != nil
}

// Printf logs a formatted message with timestamp if dev mode is on.
func Printf(format string, args ...any) {
	if logger != nil {
		logger.Printf("[%s] %s", time.Now().Format("15:04:05.000"), sprintf(format, args...))
	}
}

// Error logs an error with context.
func Error(context string, err error) {
	if logger != nil {
		logger.Printf("[%s] ERROR [%s]: %v", time.Now().Format("15:04:05.000"), context, err)
	}
}

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
