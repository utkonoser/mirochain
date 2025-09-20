package tests

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"mirochain/internal/logging"
)

// TestPerformanceLoggerCreation тестирует создание логгера производительности
func TestPerformanceLoggerCreation(t *testing.T) {
	config := logging.PerformanceConfig{
		Level:      slog.LevelInfo,
		Output:     "stdout",
		Format:     "text",
		IncludeMem: true,
		IncludeCPU: true,
	}

	logger := logging.NewPerformanceLogger(config)

	if logger == nil {
		t.Fatal("Performance logger should not be nil")
	}

	t.Logf("Performance logger created successfully")
}

// TestPerformanceLoggerLogging тестирует логирование производительности
func TestPerformanceLoggerLogging(t *testing.T) {
	config := logging.PerformanceConfig{
		Level:      slog.LevelInfo,
		Output:     "stdout",
		Format:     "text",
		IncludeMem: true,
		IncludeCPU: true,
	}

	logger := logging.NewPerformanceLogger(config)

	// Тестируем различные типы логов
	logger.LogPerformance("test_operation", 100*time.Millisecond,
		slog.String("component", "test"),
		slog.String("status", "success"))

	logger.LogSlowOperation("slow_operation", 2*time.Second, 1*time.Second,
		slog.String("component", "test"))

	logger.LogMemoryUsage("test_component",
		slog.String("component", "test"))

	logger.LogGoroutineCount("test_component",
		slog.String("component", "test"))

	logger.LogBlockchainPerformance("add_block", 500*time.Millisecond, 100, 5,
		slog.String("component", "blockchain"))

	logger.LogMiningPerformance(1*time.Second, 5, 12345, "abc123",
		slog.String("component", "mining"))

	logger.LogNetworkPerformance("send_message", 50*time.Millisecond, "peer_001", 1024,
		slog.String("component", "network"))

	logger.LogCachePerformance("get", 1*time.Millisecond, true, "key_001",
		slog.String("component", "cache"))

	logger.LogDatabasePerformance("insert", 10*time.Millisecond, "blocks", 1,
		slog.String("component", "database"))

	logger.LogError(errors.New("test error"), "test_operation", 100*time.Millisecond,
		slog.String("component", "test"))

	logger.LogWarning("test warning", "test_operation", 100*time.Millisecond,
		slog.String("component", "test"))

	logger.LogInfo("test info", "test_operation", 100*time.Millisecond,
		slog.String("component", "test"))

	logger.LogDebug("test debug", "test_operation", 100*time.Millisecond,
		slog.String("component", "test"))

	t.Logf("Performance logging test completed successfully")
}

// TestPerformanceTimer тестирует таймер производительности
func TestPerformanceTimer(t *testing.T) {
	config := logging.PerformanceConfig{
		Level:      slog.LevelInfo,
		Output:     "stdout",
		Format:     "text",
		IncludeMem: true,
		IncludeCPU: true,
	}

	logger := logging.NewPerformanceLogger(config)

	// Тестируем таймер
	timer := logging.NewPerformanceTimer(logger, "test_operation",
		slog.String("component", "test"))

	// Ждем немного
	time.Sleep(100 * time.Millisecond)

	// Проверяем прошедшее время
	elapsed := timer.Elapsed()
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected elapsed time >= 100ms, got %v", elapsed)
	}

	// Тестируем различные методы логирования
	timer.Log()
	timer.LogSlow(50 * time.Millisecond)
	timer.LogBlockchain(100, 5)
	timer.LogMining(5, 12345, "abc123")
	timer.LogNetwork("peer_001", 1024)
	timer.LogCache(true, "key_001")
	timer.LogDatabase("blocks", 1)
	timer.LogError(errors.New("test error"))
	timer.LogWarning("test warning")
	timer.LogInfo("test info")
	timer.LogDebug("test debug")

	t.Logf("Performance timer test completed successfully")
}

// TestContextLogger тестирует логгер с контекстом
func TestContextLogger(t *testing.T) {
	config := logging.PerformanceConfig{
		Level:      slog.LevelInfo,
		Output:     "stdout",
		Format:     "text",
		IncludeMem: true,
		IncludeCPU: true,
	}

	logger := logging.NewPerformanceLogger(config)
	ctx := context.Background()

	contextLogger := logging.NewContextLogger(logger, ctx)

	// Тестируем логирование с контекстом
	contextLogger.LogPerformance("test_operation", 100*time.Millisecond,
		slog.String("component", "test"))

	contextLogger.LogError(errors.New("test error"), "test_operation", 100*time.Millisecond,
		slog.String("component", "test"))

	t.Logf("Context logger test completed successfully")
}

// TestPerformanceLoggerIntegration тестирует интеграцию логгера производительности
func TestPerformanceLoggerIntegration(t *testing.T) {
	t.Run("Creation", TestPerformanceLoggerCreation)
	t.Run("Logging", TestPerformanceLoggerLogging)
	t.Run("Timer", TestPerformanceTimer)
	t.Run("Context", TestContextLogger)
}
