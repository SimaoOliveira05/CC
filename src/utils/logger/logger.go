package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Component string                 `json:"component"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LogDestination defines where logs should be sent
type LogDestination int

const (
	DestConsole LogDestination = 1 << iota
	DestFile
	DestFrontend
	DestAll = DestConsole | DestFile | DestFrontend
)

// APIPublisher interface for publishing updates to frontend
type APIPublisher interface {
	PublishUpdate(event string, data interface{})
}

// Logger is the centralized logging system
type Logger struct {
	mu               sync.Mutex
	file             *os.File
	destinations     LogDestination
	minLevel         LogLevel
	minLevelConsole  LogLevel
	minLevelFile     LogLevel
	minLevelFrontend LogLevel
	apiPublisher     APIPublisher // API server for WebSocket broadcast
}

// NewLogger creates a new logger instance
func NewLogger(logFilePath string, destinations LogDestination, minLevel LogLevel, apiPublisher APIPublisher) (*Logger, error) {
	logger := &Logger{
		destinations:     destinations,
		minLevel:         minLevel,
		minLevelConsole:  minLevel,
		minLevelFile:     DEBUG, // File always logs everything
		minLevelFrontend: INFO,  // Frontend only INFO and above
		apiPublisher:     apiPublisher,
	}

	// Open log file if file destination is enabled
	if destinations&DestFile != 0 {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.file = file
	}

	return logger, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// shouldLogToDestination checks if a log should go to a specific destination
func (l *Logger) shouldLogToDestination(level LogLevel, dest LogDestination) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}

	switch dest {
	case DestConsole:
		return levels[level] >= levels[l.minLevelConsole]
	case DestFile:
		return levels[level] >= levels[l.minLevelFile]
	case DestFrontend:
		return levels[level] >= levels[l.minLevelFrontend]
	default:
		return levels[level] >= levels[l.minLevel]
	}
}

// shouldLog checks if a log level should be logged based on minimum level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[l.minLevel]
}

// Log logs a message with the given level, component, message and optional metadata
func (l *Logger) Log(level LogLevel, component, message string, metadata map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
		Metadata:  metadata,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Terminal output
	if l.destinations&DestConsole != 0 && l.shouldLogToDestination(level, DestConsole) {
		l.writeToTerminal(entry)
	}

	// File output
	if l.destinations&DestFile != 0 && l.file != nil && l.shouldLogToDestination(level, DestFile) {
		l.writeToFile(entry)
	}

	// Frontend output via WebSocket
	if l.destinations&DestFrontend != 0 && l.apiPublisher != nil && l.shouldLogToDestination(level, DestFrontend) {
		// Convert LogEntry to Event format expected by frontend
		l.apiPublisher.PublishUpdate("log", map[string]interface{}{
			"timestamp": entry.Timestamp,
			"level":     string(entry.Level),
			"source":    entry.Component,
			"message":   entry.Message,
			"meta":      entry.Metadata,
		})
	}
}

// writeToTerminal writes colored output to terminal
func (l *Logger) writeToTerminal(entry LogEntry) {
	// ANSI color codes
	colors := map[LogLevel]string{
		DEBUG: "\033[36m", // Cyan
		INFO:  "\033[32m", // Green
		WARN:  "\033[33m", // Yellow
		ERROR: "\033[31m", // Red
	}
	reset := "\033[0m"

	color := colors[entry.Level]
	timestamp := entry.Timestamp.Format("15:04:05.000")

	metaStr := ""
	if len(entry.Metadata) > 0 {
		metaBytes, _ := json.Marshal(entry.Metadata)
		metaStr = " " + string(metaBytes)
	}

	fmt.Printf("%s[%s] [%s] [%s]%s %s%s\n",
		color,
		timestamp,
		entry.Level,
		entry.Component,
		reset,
		entry.Message,
		metaStr,
	)
}

// writeToFile writes JSON formatted logs to file
func (l *Logger) writeToFile(entry LogEntry) {
	encoder := json.NewEncoder(l.file)
	encoder.Encode(entry)
}

// Convenience methods for different log levels
func (l *Logger) Debug(component, message string, metadata map[string]interface{}) {
	l.Log(DEBUG, component, message, metadata)
}

func (l *Logger) Info(component, message string, metadata map[string]interface{}) {
	l.Log(INFO, component, message, metadata)
}

func (l *Logger) Warn(component, message string, metadata map[string]interface{}) {
	l.Log(WARN, component, message, metadata)
}

func (l *Logger) Error(component, message string, metadata map[string]interface{}) {
	l.Log(ERROR, component, message, metadata)
}

// Printf-style convenience methods
func (l *Logger) Debugf(component, format string, args ...interface{}) {
	l.Debug(component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Infof(component, format string, args ...interface{}) {
	l.Info(component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Warnf(component, format string, args ...interface{}) {
	l.Warn(component, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Errorf(component, format string, args ...interface{}) {
	l.Error(component, fmt.Sprintf(format, args...), nil)
}

// CreateLogCallback creates a logging callback function for use with packet operations
// This allows external packages to log through the logger without direct coupling
func (l *Logger) CreateLogCallback(component string) func(string, string, any) {
	return func(level, msg string, meta any) {
		switch level {
		case "ERROR":
			l.Errorf(component, "%s: %+v", msg, meta)
		case "WARN":
			l.Warnf(component, "%s: %+v", msg, meta)
		case "DEBUG":
			l.Debugf(component, "%s: %+v", msg, meta)
		default:
			l.Infof(component, "%s: %+v", msg, meta)
		}
	}
}
