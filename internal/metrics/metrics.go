package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricType представляет тип метрики
type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
	Summary
)

// Metric представляет базовую метрику
type Metric interface {
	GetName() string
	GetType() MetricType
	GetValue() interface{}
	GetLabels() map[string]string
}

// CounterMetric представляет счетчик
type CounterMetric struct {
	name   string
	labels map[string]string
	value  int64
}

// NewCounterMetric создает новый счетчик
func NewCounterMetric(name string, labels map[string]string) *CounterMetric {
	return &CounterMetric{
		name:   name,
		labels: labels,
		value:  0,
	}
}

// GetName возвращает имя метрики
func (c *CounterMetric) GetName() string {
	return c.name
}

// GetType возвращает тип метрики
func (c *CounterMetric) GetType() MetricType {
	return Counter
}

// GetValue возвращает значение метрики
func (c *CounterMetric) GetValue() interface{} {
	return atomic.LoadInt64(&c.value)
}

// GetLabels возвращает метки метрики
func (c *CounterMetric) GetLabels() map[string]string {
	return c.labels
}

// Inc увеличивает счетчик на 1
func (c *CounterMetric) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add увеличивает счетчик на указанное значение
func (c *CounterMetric) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

// Set устанавливает значение счетчика
func (c *CounterMetric) Set(value int64) {
	atomic.StoreInt64(&c.value, value)
}

// GaugeMetric представляет датчик
type GaugeMetric struct {
	name   string
	labels map[string]string
	value  int64
}

// NewGaugeMetric создает новый датчик
func NewGaugeMetric(name string, labels map[string]string) *GaugeMetric {
	return &GaugeMetric{
		name:   name,
		labels: labels,
		value:  0,
	}
}

// GetName возвращает имя метрики
func (g *GaugeMetric) GetName() string {
	return g.name
}

// GetType возвращает тип метрики
func (g *GaugeMetric) GetType() MetricType {
	return Gauge
}

// GetValue возвращает значение метрики
func (g *GaugeMetric) GetValue() interface{} {
	return atomic.LoadInt64(&g.value)
}

// GetLabels возвращает метки метрики
func (g *GaugeMetric) GetLabels() map[string]string {
	return g.labels
}

// Set устанавливает значение датчика
func (g *GaugeMetric) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Add увеличивает датчик на указанное значение
func (g *GaugeMetric) Add(delta int64) {
	atomic.AddInt64(&g.value, delta)
}

// Sub уменьшает датчик на указанное значение
func (g *GaugeMetric) Sub(delta int64) {
	atomic.AddInt64(&g.value, -delta)
}

// HistogramMetric представляет гистограмму
type HistogramMetric struct {
	name    string
	labels  map[string]string
	buckets []float64
	counts  []int64
	total   int64
}

// NewHistogramMetric создает новую гистограмму
func NewHistogramMetric(name string, labels map[string]string, buckets []float64) *HistogramMetric {
	return &HistogramMetric{
		name:    name,
		labels:  labels,
		buckets: buckets,
		counts:  make([]int64, len(buckets)),
		total:   0,
	}
}

// GetName возвращает имя метрики
func (h *HistogramMetric) GetName() string {
	return h.name
}

// GetType возвращает тип метрики
func (h *HistogramMetric) GetType() MetricType {
	return Histogram
}

// GetValue возвращает значение метрики
func (h *HistogramMetric) GetValue() interface{} {
	return map[string]interface{}{
		"buckets": h.buckets,
		"counts":  h.counts,
		"total":   atomic.LoadInt64(&h.total),
	}
}

// GetLabels возвращает метки метрики
func (h *HistogramMetric) GetLabels() map[string]string {
	return h.labels
}

// Observe добавляет наблюдение в гистограмму
func (h *HistogramMetric) Observe(value float64) {
	atomic.AddInt64(&h.total, 1)

	for i, bucket := range h.buckets {
		if value <= bucket {
			atomic.AddInt64(&h.counts[i], 1)
		}
	}
}

// SummaryMetric представляет сводку
type SummaryMetric struct {
	name   string
	labels map[string]string
	count  int64
	sum    int64
}

// NewSummaryMetric создает новую сводку
func NewSummaryMetric(name string, labels map[string]string) *SummaryMetric {
	return &SummaryMetric{
		name:   name,
		labels: labels,
		count:  0,
		sum:    0,
	}
}

// GetName возвращает имя метрики
func (s *SummaryMetric) GetName() string {
	return s.name
}

// GetType возвращает тип метрики
func (s *SummaryMetric) GetType() MetricType {
	return Summary
}

// GetValue возвращает значение метрики
func (s *SummaryMetric) GetValue() interface{} {
	return map[string]interface{}{
		"count": atomic.LoadInt64(&s.count),
		"sum":   atomic.LoadInt64(&s.sum),
	}
}

// GetLabels возвращает метки метрики
func (s *SummaryMetric) GetLabels() map[string]string {
	return s.labels
}

// Observe добавляет наблюдение в сводку
func (s *SummaryMetric) Observe(value int64) {
	atomic.AddInt64(&s.count, 1)
	atomic.AddInt64(&s.sum, value)
}

// MetricsCollector представляет сборщик метрик
type MetricsCollector struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewMetricsCollector создает новый сборщик метрик
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]Metric),
	}
}

// Register регистрирует метрику
func (mc *MetricsCollector) Register(metric Metric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics[metric.GetName()] = metric
}

// GetMetric возвращает метрику по имени
func (mc *MetricsCollector) GetMetric(name string) (Metric, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metric, exists := mc.metrics[name]
	return metric, exists
}

// GetAllMetrics возвращает все метрики
func (mc *MetricsCollector) GetAllMetrics() map[string]Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]Metric)
	for name, metric := range mc.metrics {
		result[name] = metric
	}
	return result
}

// GetCounter возвращает счетчик по имени
func (mc *MetricsCollector) GetCounter(name string) (*CounterMetric, bool) {
	metric, exists := mc.GetMetric(name)
	if !exists {
		return nil, false
	}

	counter, ok := metric.(*CounterMetric)
	return counter, ok
}

// GetGauge возвращает датчик по имени
func (mc *MetricsCollector) GetGauge(name string) (*GaugeMetric, bool) {
	metric, exists := mc.GetMetric(name)
	if !exists {
		return nil, false
	}

	gauge, ok := metric.(*GaugeMetric)
	return gauge, ok
}

// GetHistogram возвращает гистограмму по имени
func (mc *MetricsCollector) GetHistogram(name string) (*HistogramMetric, bool) {
	metric, exists := mc.GetMetric(name)
	if !exists {
		return nil, false
	}

	histogram, ok := metric.(*HistogramMetric)
	return histogram, ok
}

// GetSummary возвращает сводку по имени
func (mc *MetricsCollector) GetSummary(name string) (*SummaryMetric, bool) {
	metric, exists := mc.GetMetric(name)
	if !exists {
		return nil, false
	}

	summary, ok := metric.(*SummaryMetric)
	return summary, ok
}

// CreateCounter создает и регистрирует счетчик
func (mc *MetricsCollector) CreateCounter(name string, labels map[string]string) *CounterMetric {
	counter := NewCounterMetric(name, labels)
	mc.Register(counter)
	return counter
}

// CreateGauge создает и регистрирует датчик
func (mc *MetricsCollector) CreateGauge(name string, labels map[string]string) *GaugeMetric {
	gauge := NewGaugeMetric(name, labels)
	mc.Register(gauge)
	return gauge
}

// CreateHistogram создает и регистрирует гистограмму
func (mc *MetricsCollector) CreateHistogram(name string, labels map[string]string, buckets []float64) *HistogramMetric {
	histogram := NewHistogramMetric(name, labels, buckets)
	mc.Register(histogram)
	return histogram
}

// CreateSummary создает и регистрирует сводку
func (mc *MetricsCollector) CreateSummary(name string, labels map[string]string) *SummaryMetric {
	summary := NewSummaryMetric(name, labels)
	mc.Register(summary)
	return summary
}

// GetStats возвращает статистику всех метрик
func (mc *MetricsCollector) GetStats() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := make(map[string]interface{})
	for name, metric := range mc.metrics {
		stats[name] = map[string]interface{}{
			"type":   metric.GetType(),
			"value":  metric.GetValue(),
			"labels": metric.GetLabels(),
		}
	}
	return stats
}

// Clear очищает все метрики
func (mc *MetricsCollector) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = make(map[string]Metric)
}

// Timer представляет таймер для измерения времени выполнения
type Timer struct {
	start time.Time
}

// NewTimer создает новый таймер
func NewTimer() *Timer {
	return &Timer{
		start: time.Now(),
	}
}

// Elapsed возвращает прошедшее время
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// Observe добавляет прошедшее время в метрику
func (t *Timer) Observe(metric *SummaryMetric) {
	metric.Observe(int64(t.Elapsed().Nanoseconds()))
}

// ObserveHistogram добавляет прошедшее время в гистограмму
func (t *Timer) ObserveHistogram(metric *HistogramMetric) {
	metric.Observe(float64(t.Elapsed().Nanoseconds()))
}
