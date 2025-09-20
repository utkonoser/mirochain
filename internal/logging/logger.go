package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// LogLevel представляет уровень логирования
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogFormat представляет формат логирования
type LogFormat string

const (
	FormatJSON LogFormat = "json"
	FormatText LogFormat = "text"
)

// LoggerEntry представляет запись лога
type LoggerEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Component string                 `json:"component,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger представляет логгер
type Logger struct {
	level    LogLevel
	format   LogFormat
	output   io.Writer
	handlers map[string]*slog.Logger
}

// NewLogger создает новый логгер
func NewLogger(level LogLevel, format LogFormat, output io.Writer) *Logger {
	return &Logger{
		level:    level,
		format:   format,
		output:   output,
		handlers: make(map[string]*slog.Logger),
	}
}

// GetHandler возвращает handler для компонента
func (l *Logger) GetHandler(component string) *slog.Logger {
	if handler, exists := l.handlers[component]; exists {
		return handler
	}

	// Создаем новый handler для компонента
	handler := l.createHandler(component)
	l.handlers[component] = handler
	return handler
}

// createHandler создает handler для компонента
func (l *Logger) createHandler(component string) *slog.Logger {
	switch l.format {
	case FormatJSON:
		return slog.New(slog.NewJSONHandler(l.output, &slog.HandlerOptions{
			Level:     l.getSlogLevel(),
			AddSource: true,
		}))
	case FormatText:
		return slog.New(slog.NewTextHandler(l.output, &slog.HandlerOptions{
			Level:     l.getSlogLevel(),
			AddSource: true,
		}))
	default:
		return slog.New(slog.NewTextHandler(l.output, &slog.HandlerOptions{
			Level: l.getSlogLevel(),
		}))
	}
}

// getSlogLevel возвращает уровень slog
func (l *Logger) getSlogLevel() slog.Level {
	switch l.level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// SetLevel устанавливает уровень логирования
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
	// Пересоздаем все handlers с новым уровнем
	l.handlers = make(map[string]*slog.Logger)
}

// SetFormat устанавливает формат логирования
func (l *Logger) SetFormat(format LogFormat) {
	l.format = format
	// Пересоздаем все handlers
	l.handlers = make(map[string]*slog.Logger)
}

// SetOutput устанавливает выходной поток
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
	// Пересоздаем все handlers
	l.handlers = make(map[string]*slog.Logger)
}

// Log записывает лог
func (l *Logger) Log(level LogLevel, component, message string, fields map[string]interface{}) {
	entry := LoggerEntry{
		Timestamp: time.Now(),
		Level:     string(level),
		Message:   message,
		Component: component,
		Fields:    fields,
	}

	switch l.format {
	case FormatJSON:
		l.logJSON(entry)
	case FormatText:
		l.logText(entry)
	}
}

// logJSON записывает лог в JSON формате
func (l *Logger) logJSON(entry LoggerEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, "{\"error\":\"failed to marshal log entry\",\"message\":\"%s\"}\n", err.Error())
		return
	}
	fmt.Fprintln(l.output, string(data))
}

// logText записывает лог в текстовом формате
func (l *Logger) logText(entry LoggerEntry) {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	fmt.Fprintf(l.output, "[%s] %s [%s] %s", timestamp, entry.Level, entry.Component, entry.Message)

	if len(entry.Fields) > 0 {
		fmt.Fprint(l.output, " {")
		first := true
		for key, value := range entry.Fields {
			if !first {
				fmt.Fprint(l.output, ", ")
			}
			fmt.Fprintf(l.output, "%s=%v", key, value)
			first = false
		}
		fmt.Fprint(l.output, "}")
	}
	fmt.Fprintln(l.output)
}

// Debug записывает debug лог
func (l *Logger) Debug(component, message string, fields ...map[string]interface{}) {
	if l.level == LevelDebug {
		var mergedFields map[string]interface{}
		if len(fields) > 0 {
			mergedFields = fields[0]
		}
		l.Log(LevelDebug, component, message, mergedFields)
	}
}

// Info записывает info лог
func (l *Logger) Info(component, message string, fields ...map[string]interface{}) {
	if l.level == LevelDebug || l.level == LevelInfo {
		var mergedFields map[string]interface{}
		if len(fields) > 0 {
			mergedFields = fields[0]
		}
		l.Log(LevelInfo, component, message, mergedFields)
	}
}

// Warn записывает warn лог
func (l *Logger) Warn(component, message string, fields ...map[string]interface{}) {
	if l.level == LevelDebug || l.level == LevelInfo || l.level == LevelWarn {
		var mergedFields map[string]interface{}
		if len(fields) > 0 {
			mergedFields = fields[0]
		}
		l.Log(LevelWarn, component, message, mergedFields)
	}
}

// Error записывает error лог
func (l *Logger) Error(component, message string, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	if len(fields) > 0 {
		mergedFields = fields[0]
	}
	l.Log(LevelError, component, message, mergedFields)
}

// Global logger instance
var globalLogger *Logger

// InitLogger инициализирует глобальный логгер
func InitLogger(level LogLevel, format LogFormat, output io.Writer) {
	globalLogger = NewLogger(level, format, output)
}

// GetLogger возвращает глобальный логгер
func GetLogger() *Logger {
	if globalLogger == nil {
		// Создаем логгер по умолчанию
		globalLogger = NewLogger(LevelInfo, FormatText, os.Stdout)
	}
	return globalLogger
}

// SetLevel устанавливает уровень глобального логгера
func SetLevel(level LogLevel) {
	GetLogger().SetLevel(level)
}

// SetFormat устанавливает формат глобального логгера
func SetFormat(format LogFormat) {
	GetLogger().SetFormat(format)
}

// SetOutput устанавливает выходной поток глобального логгера
func SetOutput(output io.Writer) {
	GetLogger().SetOutput(output)
}

// Debug записывает debug лог через глобальный логгер
func Debug(component, message string, fields ...map[string]interface{}) {
	GetLogger().Debug(component, message, fields...)
}

// Info записывает info лог через глобальный логгер
func Info(component, message string, fields ...map[string]interface{}) {
	GetLogger().Info(component, message, fields...)
}

// Warn записывает warn лог через глобальный логгер
func Warn(component, message string, fields ...map[string]interface{}) {
	GetLogger().Warn(component, message, fields...)
}

// Error записывает error лог через глобальный логгер
func Error(component, message string, fields ...map[string]interface{}) {
	GetLogger().Error(component, message, fields...)
}

// CreateFileLogger создает логгер с выводом в файл
func CreateFileLogger(filename string, level LogLevel, format LogFormat) (*Logger, error) {
	// Создаем директорию если не существует
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Открываем файл для записи
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return NewLogger(level, format, file), nil
}

// RotateLogger создает ротирующий логгер
func RotateLogger(baseFilename string, level LogLevel, format LogFormat, maxSize int, maxBackups int, maxAge int) (*Logger, error) {
	// В реальной реализации здесь должна быть ротация логов
	// Пока что просто создаем обычный файловый логгер
	return CreateFileLogger(baseFilename, level, format)
}
