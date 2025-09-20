package tests

import (
	"testing"
	"time"

	"mirochain/internal/metrics"
)

// TestCounterMetric тестирует счетчик
func TestCounterMetric(t *testing.T) {
	counter := metrics.NewCounterMetric("test_counter", map[string]string{"label1": "value1"})

	// Проверяем начальное значение
	if counter.GetValue() != int64(0) {
		t.Errorf("Expected initial value 0, got %v", counter.GetValue())
	}

	// Увеличиваем счетчик
	counter.Inc()
	if counter.GetValue() != int64(1) {
		t.Errorf("Expected value 1 after Inc(), got %v", counter.GetValue())
	}

	// Добавляем значение
	counter.Add(5)
	if counter.GetValue() != int64(6) {
		t.Errorf("Expected value 6 after Add(5), got %v", counter.GetValue())
	}

	// Устанавливаем значение
	counter.Set(10)
	if counter.GetValue() != int64(10) {
		t.Errorf("Expected value 10 after Set(10), got %v", counter.GetValue())
	}

	// Проверяем метаданные
	if counter.GetName() != "test_counter" {
		t.Errorf("Expected name 'test_counter', got %s", counter.GetName())
	}

	if counter.GetType() != metrics.Counter {
		t.Errorf("Expected type Counter, got %v", counter.GetType())
	}

	labels := counter.GetLabels()
	if labels["label1"] != "value1" {
		t.Errorf("Expected label 'value1', got %s", labels["label1"])
	}

	t.Logf("Counter metric test completed successfully")
}

// TestGaugeMetric тестирует датчик
func TestGaugeMetric(t *testing.T) {
	gauge := metrics.NewGaugeMetric("test_gauge", map[string]string{"label1": "value1"})

	// Проверяем начальное значение
	if gauge.GetValue() != int64(0) {
		t.Errorf("Expected initial value 0, got %v", gauge.GetValue())
	}

	// Устанавливаем значение
	gauge.Set(100)
	if gauge.GetValue() != int64(100) {
		t.Errorf("Expected value 100 after Set(100), got %v", gauge.GetValue())
	}

	// Добавляем значение
	gauge.Add(50)
	if gauge.GetValue() != int64(150) {
		t.Errorf("Expected value 150 after Add(50), got %v", gauge.GetValue())
	}

	// Уменьшаем значение
	gauge.Sub(25)
	if gauge.GetValue() != int64(125) {
		t.Errorf("Expected value 125 after Sub(25), got %v", gauge.GetValue())
	}

	// Проверяем метаданные
	if gauge.GetName() != "test_gauge" {
		t.Errorf("Expected name 'test_gauge', got %s", gauge.GetName())
	}

	if gauge.GetType() != metrics.Gauge {
		t.Errorf("Expected type Gauge, got %v", gauge.GetType())
	}

	t.Logf("Gauge metric test completed successfully")
}

// TestHistogramMetric тестирует гистограмму
func TestHistogramMetric(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0}
	histogram := metrics.NewHistogramMetric("test_histogram", map[string]string{"label1": "value1"}, buckets)

	// Проверяем начальное значение
	value := histogram.GetValue().(map[string]interface{})
	if value["total"] != int64(0) {
		t.Errorf("Expected initial total 0, got %v", value["total"])
	}

	// Добавляем наблюдения
	histogram.Observe(0.3)
	histogram.Observe(0.7)
	histogram.Observe(1.5)
	histogram.Observe(3.0)
	histogram.Observe(7.5)

	// Проверяем значение
	value = histogram.GetValue().(map[string]interface{})
	if value["total"] != int64(5) {
		t.Errorf("Expected total 5, got %v", value["total"])
	}

	// Проверяем метаданные
	if histogram.GetName() != "test_histogram" {
		t.Errorf("Expected name 'test_histogram', got %s", histogram.GetName())
	}

	if histogram.GetType() != metrics.Histogram {
		t.Errorf("Expected type Histogram, got %v", histogram.GetType())
	}

	t.Logf("Histogram metric test completed successfully")
}

// TestSummaryMetric тестирует сводку
func TestSummaryMetric(t *testing.T) {
	summary := metrics.NewSummaryMetric("test_summary", map[string]string{"label1": "value1"})

	// Проверяем начальное значение
	value := summary.GetValue().(map[string]interface{})
	if value["count"] != int64(0) {
		t.Errorf("Expected initial count 0, got %v", value["count"])
	}
	if value["sum"] != int64(0) {
		t.Errorf("Expected initial sum 0, got %v", value["sum"])
	}

	// Добавляем наблюдения
	summary.Observe(100)
	summary.Observe(200)
	summary.Observe(300)

	// Проверяем значение
	value = summary.GetValue().(map[string]interface{})
	if value["count"] != int64(3) {
		t.Errorf("Expected count 3, got %v", value["count"])
	}
	if value["sum"] != int64(600) {
		t.Errorf("Expected sum 600, got %v", value["sum"])
	}

	// Проверяем метаданные
	if summary.GetName() != "test_summary" {
		t.Errorf("Expected name 'test_summary', got %s", summary.GetName())
	}

	if summary.GetType() != metrics.Summary {
		t.Errorf("Expected type Summary, got %v", summary.GetType())
	}

	t.Logf("Summary metric test completed successfully")
}

// TestMetricsCollector тестирует сборщик метрик
func TestMetricsCollector(t *testing.T) {
	collector := metrics.NewMetricsCollector()

	// Создаем метрики
	counter := collector.CreateCounter("test_counter", map[string]string{"label1": "value1"})
	gauge := collector.CreateGauge("test_gauge", map[string]string{"label1": "value1"})
	histogram := collector.CreateHistogram("test_histogram", map[string]string{"label1": "value1"}, []float64{0.1, 0.5, 1.0})
	summary := collector.CreateSummary("test_summary", map[string]string{"label1": "value1"})

	// Проверяем, что метрики зарегистрированы
	if _, exists := collector.GetMetric("test_counter"); !exists {
		t.Error("Counter metric should be registered")
	}

	if _, exists := collector.GetMetric("test_gauge"); !exists {
		t.Error("Gauge metric should be registered")
	}

	if _, exists := collector.GetMetric("test_histogram"); !exists {
		t.Error("Histogram metric should be registered")
	}

	if _, exists := collector.GetMetric("test_summary"); !exists {
		t.Error("Summary metric should be registered")
	}

	// Проверяем получение метрик по типу
	if _, exists := collector.GetCounter("test_counter"); !exists {
		t.Error("Counter metric should be retrievable")
	}

	if _, exists := collector.GetGauge("test_gauge"); !exists {
		t.Error("Gauge metric should be retrievable")
	}

	if _, exists := collector.GetHistogram("test_histogram"); !exists {
		t.Error("Histogram metric should be retrievable")
	}

	if _, exists := collector.GetSummary("test_summary"); !exists {
		t.Error("Summary metric should be retrievable")
	}

	// Проверяем несуществующую метрику
	if _, exists := collector.GetMetric("nonexistent"); exists {
		t.Error("Nonexistent metric should not exist")
	}

	// Проверяем статистику
	stats := collector.GetStats()
	if len(stats) != 4 {
		t.Errorf("Expected 4 metrics in stats, got %d", len(stats))
	}

	// Проверяем, что метрики работают
	counter.Inc()
	gauge.Set(100)
	histogram.Observe(0.5)
	summary.Observe(200)

	// Проверяем статистику после изменений
	stats = collector.GetStats()
	counterStats := stats["test_counter"].(map[string]interface{})
	if counterStats["value"] != int64(1) {
		t.Errorf("Expected counter value 1, got %v", counterStats["value"])
	}

	t.Logf("Metrics collector test completed successfully")
}

// TestTimer тестирует таймер
func TestTimer(t *testing.T) {
	timer := metrics.NewTimer()

	// Ждем немного
	time.Sleep(10 * time.Millisecond)

	// Проверяем прошедшее время
	elapsed := timer.Elapsed()
	if elapsed < 10*time.Millisecond {
		t.Errorf("Expected elapsed time >= 10ms, got %v", elapsed)
	}

	// Тестируем с метрикой
	summary := metrics.NewSummaryMetric("test_timer", map[string]string{})
	timer.Observe(summary)

	value := summary.GetValue().(map[string]interface{})
	if value["count"] != int64(1) {
		t.Errorf("Expected count 1, got %v", value["count"])
	}

	if value["sum"] == int64(0) {
		t.Error("Expected sum > 0")
	}

	// Тестируем с гистограммой
	histogram := metrics.NewHistogramMetric("test_timer_hist", map[string]string{}, []float64{0.001, 0.01, 0.1, 1.0})
	timer.ObserveHistogram(histogram)

	value = histogram.GetValue().(map[string]interface{})
	if value["total"] != int64(1) {
		t.Errorf("Expected total 1, got %v", value["total"])
	}

	t.Logf("Timer test completed successfully")
}

// TestMetricsIntegration тестирует интеграцию метрик
func TestMetricsIntegration(t *testing.T) {
	t.Run("Counter", TestCounterMetric)
	t.Run("Gauge", TestGaugeMetric)
	t.Run("Histogram", TestHistogramMetric)
	t.Run("Summary", TestSummaryMetric)
	t.Run("Collector", TestMetricsCollector)
	t.Run("Timer", TestTimer)
}
