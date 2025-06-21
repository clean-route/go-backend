package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var currentLevel = INFO

// Init initializes the logger with appropriate configuration
func Init() error {
	// Set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		currentLevel = DEBUG
	case "info":
		currentLevel = INFO
	case "warn":
		currentLevel = WARN
	case "error":
		currentLevel = ERROR
	default:
		currentLevel = INFO
	}

	// Configure standard logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	// No-op for simple logger
}

// shouldLog checks if the message should be logged at the current level
func shouldLog(level LogLevel) bool {
	return level >= currentLevel
}

// formatMessage formats a log message with timestamp and level
func formatMessage(level string, msg string, fields ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	if len(fields) == 0 {
		return fmt.Sprintf("[%s] %s: %s", timestamp, level, msg)
	}

	// Simple field formatting
	fieldStr := ""
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			if fieldStr != "" {
				fieldStr += " "
			}
			fieldStr += fmt.Sprintf("%v=%v", fields[i], fields[i+1])
		}
	}

	return fmt.Sprintf("[%s] %s: %s %s", timestamp, level, msg, fieldStr)
}

// Debug logs a debug message
func Debug(msg string, fields ...interface{}) {
	if shouldLog(DEBUG) {
		log.Println(formatMessage("DEBUG", msg, fields...))
	}
}

// Info logs an info message
func Info(msg string, fields ...interface{}) {
	if shouldLog(INFO) {
		log.Println(formatMessage("INFO", msg, fields...))
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...interface{}) {
	if shouldLog(WARN) {
		log.Println(formatMessage("WARN", msg, fields...))
	}
}

// Error logs an error message
func Error(msg string, fields ...interface{}) {
	if shouldLog(ERROR) {
		log.Println(formatMessage("ERROR", msg, fields...))
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...interface{}) {
	if shouldLog(FATAL) {
		log.Println(formatMessage("FATAL", msg, fields...))
		os.Exit(1)
	}
}

// WithContext creates a logger with additional context fields
func WithContext(fields ...interface{}) *ContextLogger {
	return &ContextLogger{fields: fields}
}

// ContextLogger provides context-aware logging
type ContextLogger struct {
	fields []interface{}
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(msg string, fields ...interface{}) {
	allFields := append(cl.fields, fields...)
	Debug(msg, allFields...)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(msg string, fields ...interface{}) {
	allFields := append(cl.fields, fields...)
	Info(msg, allFields...)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(msg string, fields ...interface{}) {
	allFields := append(cl.fields, fields...)
	Warn(msg, allFields...)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(msg string, fields ...interface{}) {
	allFields := append(cl.fields, fields...)
	Error(msg, allFields...)
}

// Fatal logs a fatal message with context and exits
func (cl *ContextLogger) Fatal(msg string, fields ...interface{}) {
	allFields := append(cl.fields, fields...)
	Fatal(msg, allFields...)
}
