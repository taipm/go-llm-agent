package agent

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/types"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Color returns ANSI color code for the log level
func (l LogLevel) Color() string {
	switch l {
	case LogLevelDebug:
		return "\033[36m" // Cyan
	case LogLevelInfo:
		return "\033[32m" // Green
	case LogLevelWarn:
		return "\033[33m" // Yellow
	case LogLevelError:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}

// Logger defines the interface for agent logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(level LogLevel)
}

// ConsoleLogger implements colored console logging
type ConsoleLogger struct {
	level      LogLevel
	output     io.Writer
	timestamps bool
	colors     bool
}

// NewConsoleLogger creates a new console logger
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		level:      LogLevelInfo,
		output:     os.Stdout,
		timestamps: true,
		colors:     true,
	}
}

// SetLevel sets the minimum log level
func (l *ConsoleLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the output writer
func (l *ConsoleLogger) SetOutput(w io.Writer) {
	l.output = w
}

// SetTimestamps enables/disables timestamps
func (l *ConsoleLogger) SetTimestamps(enabled bool) {
	l.timestamps = enabled
}

// SetColors enables/disables colors
func (l *ConsoleLogger) SetColors(enabled bool) {
	l.colors = enabled
}

// log writes a log message
func (l *ConsoleLogger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	var parts []string

	// Add timestamp if enabled
	if l.timestamps {
		timestamp := time.Now().Format("15:04:05")
		parts = append(parts, timestamp)
	}

	// Add level with color
	levelStr := fmt.Sprintf("[%s]", level.String())
	if l.colors {
		levelStr = level.Color() + levelStr + "\033[0m"
	}
	parts = append(parts, levelStr)

	// Add message
	message := fmt.Sprintf(format, args...)
	parts = append(parts, message)

	// Write to output
	fmt.Fprintln(l.output, strings.Join(parts, " "))
}

// Debug logs a debug message
func (l *ConsoleLogger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info logs an info message
func (l *ConsoleLogger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *ConsoleLogger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *ConsoleLogger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// NoopLogger is a logger that does nothing (for disabling logging)
type NoopLogger struct{}

func (l *NoopLogger) Debug(format string, args ...interface{}) {}
func (l *NoopLogger) Info(format string, args ...interface{})  {}
func (l *NoopLogger) Warn(format string, args ...interface{})  {}
func (l *NoopLogger) Error(format string, args ...interface{}) {}
func (l *NoopLogger) SetLevel(level LogLevel)                  {}

// Helper functions for logging agent actions

// LogToolCall logs when a tool is being called
func LogToolCall(logger Logger, toolName string, params map[string]interface{}) {
	logger.Info("🔧 Calling tool: %s", toolName)
	if len(params) > 0 {
		logger.Debug("   Parameters: %v", params)
	}
}

// LogToolResult logs the result of a tool call
func LogToolResult(logger Logger, toolName string, success bool, result interface{}) {
	if success {
		logger.Info("✓ Tool %s completed successfully", toolName)
		logger.Debug("   Result: %v", result)
	} else {
		logger.Error("✗ Tool %s failed: %v", toolName, result)
	}
}

// LogThinking logs when the agent is thinking (LLM processing)
func LogThinking(logger Logger) {
	logger.Info("🤔 Agent thinking...")
}

// LogResponse logs the agent's response
func LogResponse(logger Logger, response string) {
	logger.Info("💬 Agent response:")
	// Indent multi-line responses
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if line != "" {
			logger.Info("   %s", line)
		}
	}
}

// LogMemory logs memory operations
func LogMemory(logger Logger, operation string, count int) {
	logger.Debug("💾 Memory %s: %d messages", operation, count)
}

// LogUserMessage logs user input
func LogUserMessage(logger Logger, message string) {
	logger.Info("👤 User: %s", message)
}

// LogIteration logs agent iteration in tool calling loop
func LogIteration(logger Logger, iteration int, maxIterations int) {
	logger.Debug("🔄 Iteration %d/%d", iteration+1, maxIterations)
}

// FormatToolCalls formats tool calls for logging
func FormatToolCalls(toolCalls []types.ToolCall) string {
	if len(toolCalls) == 0 {
		return "none"
	}

	names := make([]string, len(toolCalls))
	for i, tc := range toolCalls {
		names[i] = tc.Function.Name
	}
	return strings.Join(names, ", ")
}
