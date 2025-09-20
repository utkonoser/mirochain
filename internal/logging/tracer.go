package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TraceID представляет ID трассировки
type TraceID string

// SpanID представляет ID span'а
type SpanID string

// TraceContext представляет контекст трассировки
type TraceContext struct {
	TraceID TraceID `json:"trace_id"`
	SpanID  SpanID  `json:"span_id"`
}

// Span представляет span трассировки
type Span struct {
	TraceID   TraceID           `json:"trace_id"`
	SpanID    SpanID            `json:"span_id"`
	ParentID  SpanID            `json:"parent_id,omitempty"`
	Operation string            `json:"operation"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time,omitempty"`
	Duration  time.Duration     `json:"duration,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	Logs      []SpanLogEntry    `json:"logs,omitempty"`
	Error     string            `json:"error,omitempty"`
	Component string            `json:"component"`
	Status    string            `json:"status"` // "started", "completed", "error"
}

// SpanLogEntry представляет запись в span
type SpanLogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Tracer представляет трассировщик
type Tracer struct {
	spans    map[SpanID]*Span
	spansMux sync.RWMutex
	logger   *Logger
}

// NewTracer создает новый трассировщик
func NewTracer(logger *Logger) *Tracer {
	return &Tracer{
		spans:  make(map[SpanID]*Span),
		logger: logger,
	}
}

// StartSpan начинает новый span
func (t *Tracer) StartSpan(operation string, component string, parentID ...SpanID) SpanID {
	spanID := SpanID(generateID())
	traceID := TraceID(generateID())

	// Если есть parent span, используем его trace ID
	if len(parentID) > 0 {
		if parentSpan, exists := t.getSpan(parentID[0]); exists {
			traceID = parentSpan.TraceID
		}
	}

	span := &Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Operation: operation,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
		Logs:      make([]SpanLogEntry, 0),
		Component: component,
		Status:    "started",
	}

	if len(parentID) > 0 {
		span.ParentID = parentID[0]
	}

	t.setSpan(spanID, span)

	// Логируем начало span'а
	t.logger.Info("tracer", fmt.Sprintf("Started span: %s", operation), map[string]interface{}{
		"trace_id":  traceID,
		"span_id":   spanID,
		"operation": operation,
		"component": component,
	})

	return spanID
}

// FinishSpan завершает span
func (t *Tracer) FinishSpan(spanID SpanID, err error) {
	span, exists := t.getSpan(spanID)
	if !exists {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	span.Status = "completed"

	if err != nil {
		span.Error = err.Error()
		span.Status = "error"
	}

	// Логируем завершение span'а
	t.logger.Info("tracer", fmt.Sprintf("Finished span: %s", span.Operation), map[string]interface{}{
		"trace_id":  span.TraceID,
		"span_id":   span.SpanID,
		"operation": span.Operation,
		"duration":  span.Duration,
		"status":    span.Status,
		"error":     span.Error,
	})

	// Удаляем span из памяти после некоторого времени
	go func() {
		time.Sleep(5 * time.Minute)
		t.removeSpan(spanID)
	}()
}

// AddTag добавляет тег к span'у
func (t *Tracer) AddTag(spanID SpanID, key, value string) {
	span, exists := t.getSpan(spanID)
	if !exists {
		return
	}

	span.Tags[key] = value
}

// AddLog добавляет лог к span'у
func (t *Tracer) AddLog(spanID SpanID, message string, fields map[string]interface{}) {
	span, exists := t.getSpan(spanID)
	if !exists {
		return
	}

	logEntry := SpanLogEntry{
		Timestamp: time.Now(),
		Message:   message,
		Fields:    fields,
	}

	span.Logs = append(span.Logs, logEntry)
}

// GetSpan возвращает span по ID
func (t *Tracer) GetSpan(spanID SpanID) (*Span, bool) {
	return t.getSpan(spanID)
}

// GetSpansByTraceID возвращает все span'ы по trace ID
func (t *Tracer) GetSpansByTraceID(traceID TraceID) []*Span {
	t.spansMux.RLock()
	defer t.spansMux.RUnlock()

	var result []*Span
	for _, span := range t.spans {
		if span.TraceID == traceID {
			result = append(result, span)
		}
	}

	return result
}

// GetAllSpans возвращает все активные span'ы
func (t *Tracer) GetAllSpans() []*Span {
	t.spansMux.RLock()
	defer t.spansMux.RUnlock()

	var result []*Span
	for _, span := range t.spans {
		result = append(result, span)
	}

	return result
}

// getSpan возвращает span по ID (внутренний метод)
func (t *Tracer) getSpan(spanID SpanID) (*Span, bool) {
	t.spansMux.RLock()
	defer t.spansMux.RUnlock()

	span, exists := t.spans[spanID]
	return span, exists
}

// setSpan устанавливает span (внутренний метод)
func (t *Tracer) setSpan(spanID SpanID, span *Span) {
	t.spansMux.Lock()
	defer t.spansMux.Unlock()

	t.spans[spanID] = span
}

// removeSpan удаляет span (внутренний метод)
func (t *Tracer) removeSpan(spanID SpanID) {
	t.spansMux.Lock()
	defer t.spansMux.Unlock()

	delete(t.spans, spanID)
}

// TraceTransaction трассирует транзакцию
func (t *Tracer) TraceTransaction(txID string, operation string) SpanID {
	spanID := t.StartSpan(operation, "transaction")
	t.AddTag(spanID, "transaction_id", txID)
	return spanID
}

// TraceBlock трассирует блок
func (t *Tracer) TraceBlock(blockHash string, operation string) SpanID {
	spanID := t.StartSpan(operation, "block")
	t.AddTag(spanID, "block_hash", blockHash)
	return spanID
}

// TraceNetwork трассирует сетевую операцию
func (t *Tracer) TraceNetwork(peerID string, operation string) SpanID {
	spanID := t.StartSpan(operation, "network")
	t.AddTag(spanID, "peer_id", peerID)
	return spanID
}

// TraceMining трассирует майнинг
func (t *Tracer) TraceMining(blockHeight int64, operation string) SpanID {
	spanID := t.StartSpan(operation, "mining")
	t.AddTag(spanID, "block_height", fmt.Sprintf("%d", blockHeight))
	return spanID
}

// ExportSpans экспортирует span'ы в JSON
func (t *Tracer) ExportSpans() ([]byte, error) {
	spans := t.GetAllSpans()
	return json.MarshalIndent(spans, "", "  ")
}

// ExportSpansByTraceID экспортирует span'ы по trace ID
func (t *Tracer) ExportSpansByTraceID(traceID TraceID) ([]byte, error) {
	spans := t.GetSpansByTraceID(traceID)
	return json.MarshalIndent(spans, "", "  ")
}

// Global tracer instance
var globalTracer *Tracer

// InitTracer инициализирует глобальный трассировщик
func InitTracer(logger *Logger) {
	globalTracer = NewTracer(logger)
}

// GetTracer возвращает глобальный трассировщик
func GetTracer() *Tracer {
	if globalTracer == nil {
		globalTracer = NewTracer(GetLogger())
	}
	return globalTracer
}

// StartSpan начинает span через глобальный трассировщик
func StartSpan(operation string, component string, parentID ...SpanID) SpanID {
	return GetTracer().StartSpan(operation, component, parentID...)
}

// FinishSpan завершает span через глобальный трассировщик
func FinishSpan(spanID SpanID, err error) {
	GetTracer().FinishSpan(spanID, err)
}

// AddTag добавляет тег через глобальный трассировщик
func AddTag(spanID SpanID, key, value string) {
	GetTracer().AddTag(spanID, key, value)
}

// AddLog добавляет лог через глобальный трассировщик
func AddLog(spanID SpanID, message string, fields map[string]interface{}) {
	GetTracer().AddLog(spanID, message, fields)
}

// TraceTransaction трассирует транзакцию через глобальный трассировщик
func TraceTransaction(txID string, operation string) SpanID {
	return GetTracer().TraceTransaction(txID, operation)
}

// TraceBlock трассирует блок через глобальный трассировщик
func TraceBlock(blockHash string, operation string) SpanID {
	return GetTracer().TraceBlock(blockHash, operation)
}

// TraceNetwork трассирует сетевую операцию через глобальный трассировщик
func TraceNetwork(peerID string, operation string) SpanID {
	return GetTracer().TraceNetwork(peerID, operation)
}

// TraceMining трассирует майнинг через глобальный трассировщик
func TraceMining(blockHeight int64, operation string) SpanID {
	return GetTracer().TraceMining(blockHeight, operation)
}

// generateID генерирует уникальный ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ContextWithTrace добавляет trace context в контекст
func ContextWithTrace(ctx context.Context, traceID TraceID, spanID SpanID) context.Context {
	return context.WithValue(ctx, "trace_context", TraceContext{
		TraceID: traceID,
		SpanID:  spanID,
	})
}

// TraceFromContext извлекает trace context из контекста
func TraceFromContext(ctx context.Context) (TraceContext, bool) {
	traceCtx, ok := ctx.Value("trace_context").(TraceContext)
	return traceCtx, ok
}
