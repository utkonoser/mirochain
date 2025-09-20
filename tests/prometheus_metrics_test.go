package tests

import (
	"testing"
	"time"

	"mirochain/internal/metrics"
)

// TestPrometheusCollectorCreation тестирует создание Prometheus коллектора
func TestPrometheusCollectorCreation(t *testing.T) {
	collector := metrics.NewPrometheusCollector()

	if collector == nil {
		t.Fatal("Prometheus collector should not be nil")
	}

	if collector.GetRegistry() == nil {
		t.Fatal("Registry should not be nil")
	}

	t.Logf("Prometheus collector created successfully")
}

// TestPrometheusCollectorMetrics тестирует работу с метриками
func TestPrometheusCollectorMetrics(t *testing.T) {
	collector := metrics.NewPrometheusCollector()
	nodeID := "test_node"

	// Тестируем счетчики
	collector.IncBlocksMined(nodeID, "5")
	collector.AddTransactionsProcessed(nodeID, "coinbase", 10)
	collector.IncErrors(nodeID, "validation")

	// Тестируем датчики
	collector.SetBlockchainHeight(nodeID, 100)
	collector.SetUTXOCount(nodeID, 500)
	collector.SetActiveConnections(nodeID, "tcp", 5)
	collector.SetMemoryUsage(nodeID, "blockchain", 1024*1024)

	// Тестируем гистограммы
	collector.ObserveBlockMiningTime(nodeID, "5", 1*time.Second)
	collector.ObserveTransactionProcessingTime(nodeID, "coinbase", 100*time.Millisecond)
	collector.ObserveBlockSize(nodeID, 1024)

	// Тестируем сводки
	collector.ObserveBlockValidationTime(nodeID, 50*time.Millisecond)
	collector.ObserveNetworkLatency(nodeID, "peer1", 10*time.Millisecond)

	t.Logf("Prometheus metrics test completed successfully")
}

// TestBlockchainMetrics тестирует метрики блокчейна
func TestBlockchainMetrics(t *testing.T) {
	collector := metrics.NewPrometheusCollector()
	nodeID := "test_node"

	bm := metrics.NewBlockchainMetrics(collector, nodeID)

	// Тестируем события блокчейна
	bm.OnBlockMined(5, 1*time.Second, 1024)
	bm.OnTransactionProcessed("coinbase", 100*time.Millisecond)
	bm.OnBlockchainUpdated(100, 500)
	bm.OnError("validation")
	bm.OnBlockValidated(50 * time.Millisecond)

	t.Logf("Blockchain metrics test completed successfully")
}

// TestNetworkMetrics тестирует метрики сети
func TestNetworkMetrics(t *testing.T) {
	collector := metrics.NewPrometheusCollector()
	nodeID := "test_node"

	nm := metrics.NewNetworkMetrics(collector, nodeID)

	// Тестируем события сети
	nm.OnConnectionEstablished("tcp")
	nm.OnConnectionClosed("tcp")
	nm.OnNetworkLatency("peer1", 10*time.Millisecond)
	nm.OnError("connection")

	t.Logf("Network metrics test completed successfully")
}

// TestPrometheusCollectorIntegration тестирует интеграцию Prometheus коллектора
func TestPrometheusCollectorIntegration(t *testing.T) {
	t.Run("Creation", TestPrometheusCollectorCreation)
	t.Run("Metrics", TestPrometheusCollectorMetrics)
	t.Run("Blockchain", TestBlockchainMetrics)
	t.Run("Network", TestNetworkMetrics)
}
