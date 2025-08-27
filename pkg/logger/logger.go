package logger

import (
	"log"
	"os"
)

// Logger encapsulates logging functionality
type Logger struct {
	logger *log.Logger
}

// New creates a new logger instance
func New(prefix string) *Logger {
	return &Logger{
		logger: log.New(os.Stdout, prefix+": ", log.LstdFlags),
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.logger.Printf("INFO: "+format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.logger.Printf("ERROR: "+format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.logger.Printf("FATAL: "+format, v...)
	os.Exit(1)
}

// Printf logs a formatted message
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}