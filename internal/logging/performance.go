package logging

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// PerformanceLogger представляет логгер производительности
type PerformanceLogger struct {
	logger *slog.Logger
	level  slog.Level
}

// PerformanceConfig представляет конфигурацию логгера производительности
type PerformanceConfig struct {
	Level      slog.Level
	Output     string
	Format     string
	IncludeMem bool
	IncludeCPU bool
}

// NewPerformanceLogger создает новый логгер производительности
func NewPerformanceLogger(config PerformanceConfig) *PerformanceLogger {
	var handler slog.Handler

	// Настраиваем вывод
	var output *os.File
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stdout
	}

	// Настраиваем формат
	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: config.Level,
		})
	case "text":
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: config.Level,
		})
	default:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: config.Level,
		})
	}

	return &PerformanceLogger{
		logger: slog.New(handler),
		level:  config.Level,
	}
}

// LogPerformance записывает лог производительности
func (pl *PerformanceLogger) LogPerformance(operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs, slog.String("operation", operation))
	attrs = append(attrs, slog.Duration("duration", duration))

	// Добавляем информацию о памяти
	if pl.includeMem() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		attrs = append(attrs,
			slog.Uint64("mem_alloc_bytes", m.Alloc),
			slog.Uint64("mem_total_alloc_bytes", m.TotalAlloc),
			slog.Uint64("mem_sys_bytes", m.Sys),
			slog.Int("mem_num_gc", int(m.NumGC)),
		)
	}

	// Добавляем информацию о CPU
	if pl.includeCPU() {
		attrs = append(attrs, slog.Int("num_goroutines", runtime.NumGoroutine()))
	}

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Performance log", attrs...)
}

// LogSlowOperation записывает лог медленной операции
func (pl *PerformanceLogger) LogSlowOperation(operation string, duration time.Duration, threshold time.Duration, attrs ...slog.Attr) {
	if duration > threshold {
		attrs = append(attrs,
			slog.String("operation", operation),
			slog.Duration("duration", duration),
			slog.Duration("threshold", threshold),
		)

		pl.logger.LogAttrs(nil, slog.LevelWarn, "Slow operation detected", attrs...)
	}
}

// LogMemoryUsage записывает лог использования памяти
func (pl *PerformanceLogger) LogMemoryUsage(component string, attrs ...slog.Attr) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	attrs = append(attrs,
		slog.String("component", component),
		slog.Uint64("alloc_bytes", m.Alloc),
		slog.Uint64("total_alloc_bytes", m.TotalAlloc),
		slog.Uint64("sys_bytes", m.Sys),
		slog.Uint64("heap_alloc_bytes", m.HeapAlloc),
		slog.Uint64("heap_sys_bytes", m.HeapSys),
		slog.Uint64("heap_idle_bytes", m.HeapIdle),
		slog.Uint64("heap_inuse_bytes", m.HeapInuse),
		slog.Uint64("heap_released_bytes", m.HeapReleased),
		slog.Uint64("heap_objects", m.HeapObjects),
		slog.Int("num_gc", int(m.NumGC)),
		slog.Duration("gc_pause_total", time.Duration(m.PauseTotalNs)),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Memory usage", attrs...)
}

// LogGoroutineCount записывает лог количества горутин
func (pl *PerformanceLogger) LogGoroutineCount(component string, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("component", component),
		slog.Int("num_goroutines", runtime.NumGoroutine()),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Goroutine count", attrs...)
}

// LogBlockchainPerformance записывает лог производительности блокчейна
func (pl *PerformanceLogger) LogBlockchainPerformance(operation string, duration time.Duration, height int64, txCount int, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.Int64("height", height),
		slog.Int("transaction_count", txCount),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Blockchain performance", attrs...)
}

// LogMiningPerformance записывает лог производительности майнинга
func (pl *PerformanceLogger) LogMiningPerformance(duration time.Duration, difficulty int, nonce int64, hash string, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.Duration("duration", duration),
		slog.Int("difficulty", difficulty),
		slog.Int64("nonce", nonce),
		slog.String("hash", hash),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Mining performance", attrs...)
}

// LogNetworkPerformance записывает лог производительности сети
func (pl *PerformanceLogger) LogNetworkPerformance(operation string, duration time.Duration, peerID string, messageSize int, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("peer_id", peerID),
		slog.Int("message_size", messageSize),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Network performance", attrs...)
}

// LogCachePerformance записывает лог производительности кэша
func (pl *PerformanceLogger) LogCachePerformance(operation string, duration time.Duration, hit bool, key string, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.Bool("hit", hit),
		slog.String("key", key),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Cache performance", attrs...)
}

// LogDatabasePerformance записывает лог производительности базы данных
func (pl *PerformanceLogger) LogDatabasePerformance(operation string, duration time.Duration, table string, recordCount int, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("table", table),
		slog.Int("record_count", recordCount),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Database performance", attrs...)
}

// LogError записывает лог ошибки с контекстом производительности
func (pl *PerformanceLogger) LogError(err error, operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("error", err.Error()),
	)

	pl.logger.LogAttrs(nil, slog.LevelError, "Performance error", attrs...)
}

// LogWarning записывает лог предупреждения с контекстом производительности
func (pl *PerformanceLogger) LogWarning(message string, operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("message", message),
	)

	pl.logger.LogAttrs(nil, slog.LevelWarn, "Performance warning", attrs...)
}

// LogInfo записывает информационный лог с контекстом производительности
func (pl *PerformanceLogger) LogInfo(message string, operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("message", message),
	)

	pl.logger.LogAttrs(nil, slog.LevelInfo, "Performance info", attrs...)
}

// LogDebug записывает отладочный лог с контекстом производительности
func (pl *PerformanceLogger) LogDebug(message string, operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("message", message),
	)

	pl.logger.LogAttrs(nil, slog.LevelDebug, "Performance debug", attrs...)
}

// Вспомогательные методы

func (pl *PerformanceLogger) includeMem() bool {
	// В реальной реализации здесь была бы проверка конфигурации
	return true
}

func (pl *PerformanceLogger) includeCPU() bool {
	// В реальной реализации здесь была бы проверка конфигурации
	return true
}

// PerformanceTimer представляет таймер для измерения производительности
type PerformanceTimer struct {
	start     time.Time
	operation string
	logger    *PerformanceLogger
	attrs     []slog.Attr
}

// NewPerformanceTimer создает новый таймер производительности
func NewPerformanceTimer(logger *PerformanceLogger, operation string, attrs ...slog.Attr) *PerformanceTimer {
	return &PerformanceTimer{
		start:     time.Now(),
		operation: operation,
		logger:    logger,
		attrs:     attrs,
	}
}

// Elapsed возвращает прошедшее время
func (pt *PerformanceTimer) Elapsed() time.Duration {
	return time.Since(pt.start)
}

// Log записывает лог производительности
func (pt *PerformanceTimer) Log() {
	pt.logger.LogPerformance(pt.operation, pt.Elapsed(), pt.attrs...)
}

// LogSlow записывает лог медленной операции
func (pt *PerformanceTimer) LogSlow(threshold time.Duration) {
	pt.logger.LogSlowOperation(pt.operation, pt.Elapsed(), threshold, pt.attrs...)
}

// LogBlockchain записывает лог производительности блокчейна
func (pt *PerformanceTimer) LogBlockchain(height int64, txCount int) {
	pt.logger.LogBlockchainPerformance(pt.operation, pt.Elapsed(), height, txCount, pt.attrs...)
}

// LogMining записывает лог производительности майнинга
func (pt *PerformanceTimer) LogMining(difficulty int, nonce int64, hash string) {
	pt.logger.LogMiningPerformance(pt.Elapsed(), difficulty, nonce, hash, pt.attrs...)
}

// LogNetwork записывает лог производительности сети
func (pt *PerformanceTimer) LogNetwork(peerID string, messageSize int) {
	pt.logger.LogNetworkPerformance(pt.operation, pt.Elapsed(), peerID, messageSize, pt.attrs...)
}

// LogCache записывает лог производительности кэша
func (pt *PerformanceTimer) LogCache(hit bool, key string) {
	pt.logger.LogCachePerformance(pt.operation, pt.Elapsed(), hit, key, pt.attrs...)
}

// LogDatabase записывает лог производительности базы данных
func (pt *PerformanceTimer) LogDatabase(table string, recordCount int) {
	pt.logger.LogDatabasePerformance(pt.operation, pt.Elapsed(), table, recordCount, pt.attrs...)
}

// LogError записывает лог ошибки
func (pt *PerformanceTimer) LogError(err error) {
	pt.logger.LogError(err, pt.operation, pt.Elapsed(), pt.attrs...)
}

// LogWarning записывает лог предупреждения
func (pt *PerformanceTimer) LogWarning(message string) {
	pt.logger.LogWarning(message, pt.operation, pt.Elapsed(), pt.attrs...)
}

// LogInfo записывает информационный лог
func (pt *PerformanceTimer) LogInfo(message string) {
	pt.logger.LogInfo(message, pt.operation, pt.Elapsed(), pt.attrs...)
}

// LogDebug записывает отладочный лог
func (pt *PerformanceTimer) LogDebug(message string) {
	pt.logger.LogDebug(message, pt.operation, pt.Elapsed(), pt.attrs...)
}

// ContextLogger представляет логгер с контекстом
type ContextLogger struct {
	logger *PerformanceLogger
	ctx    context.Context
}

// NewContextLogger создает новый логгер с контекстом
func NewContextLogger(logger *PerformanceLogger, ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: logger,
		ctx:    ctx,
	}
}

// LogPerformance записывает лог производительности с контекстом
func (cl *ContextLogger) LogPerformance(operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs, slog.String("operation", operation))
	attrs = append(attrs, slog.Duration("duration", duration))

	cl.logger.logger.LogAttrs(cl.ctx, slog.LevelInfo, "Performance log", attrs...)
}

// LogError записывает лог ошибки с контекстом
func (cl *ContextLogger) LogError(err error, operation string, duration time.Duration, attrs ...slog.Attr) {
	attrs = append(attrs,
		slog.String("operation", operation),
		slog.Duration("duration", duration),
		slog.String("error", err.Error()),
	)

	cl.logger.logger.LogAttrs(cl.ctx, slog.LevelError, "Performance error", attrs...)
}
