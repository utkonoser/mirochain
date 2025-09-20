package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusCollector представляет сборщик метрик Prometheus
type PrometheusCollector struct {
	registry *prometheus.Registry

	// Счетчики
	blocksMinedTotal           *prometheus.CounterVec
	transactionsProcessedTotal *prometheus.CounterVec
	errorsTotal                *prometheus.CounterVec

	// Датчики
	blockchainHeight  *prometheus.GaugeVec
	utxoCount         *prometheus.GaugeVec
	activeConnections *prometheus.GaugeVec
	memoryUsage       *prometheus.GaugeVec

	// Гистограммы
	blockMiningTime           *prometheus.HistogramVec
	transactionProcessingTime *prometheus.HistogramVec
	blockSize                 *prometheus.HistogramVec

	// Сводки
	blockValidationTime *prometheus.SummaryVec
	networkLatency      *prometheus.SummaryVec
}

// NewPrometheusCollector создает новый сборщик Prometheus метрик
func NewPrometheusCollector() *PrometheusCollector {
	registry := prometheus.NewRegistry()

	pc := &PrometheusCollector{
		registry: registry,

		// Счетчики
		blocksMinedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mirochain_blocks_mined_total",
				Help: "Total number of blocks mined",
			},
			[]string{"node_id", "difficulty"},
		),

		transactionsProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mirochain_transactions_processed_total",
				Help: "Total number of transactions processed",
			},
			[]string{"node_id", "type"},
		),

		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mirochain_errors_total",
				Help: "Total number of errors",
			},
			[]string{"node_id", "error_type"},
		),

		// Датчики
		blockchainHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mirochain_blockchain_height",
				Help: "Current blockchain height",
			},
			[]string{"node_id"},
		),

		utxoCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mirochain_utxo_count",
				Help: "Current number of UTXOs",
			},
			[]string{"node_id"},
		),

		activeConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mirochain_active_connections",
				Help: "Number of active network connections",
			},
			[]string{"node_id", "connection_type"},
		),

		memoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mirochain_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"node_id", "component"},
		),

		// Гистограммы
		blockMiningTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mirochain_block_mining_time_seconds",
				Help:    "Time taken to mine a block",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
			},
			[]string{"node_id", "difficulty"},
		),

		transactionProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mirochain_transaction_processing_time_seconds",
				Help:    "Time taken to process a transaction",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
			},
			[]string{"node_id", "transaction_type"},
		),

		blockSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "mirochain_block_size_bytes",
				Help:    "Size of blocks in bytes",
				Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
			},
			[]string{"node_id"},
		),

		// Сводки
		blockValidationTime: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "mirochain_block_validation_time_seconds",
				Help:       "Time taken to validate a block",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"node_id"},
		),

		networkLatency: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "mirochain_network_latency_seconds",
				Help:       "Network latency in seconds",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"node_id", "peer_id"},
		),
	}

	// Регистрируем метрики
	registry.MustRegister(
		pc.blocksMinedTotal,
		pc.transactionsProcessedTotal,
		pc.errorsTotal,
		pc.blockchainHeight,
		pc.utxoCount,
		pc.activeConnections,
		pc.memoryUsage,
		pc.blockMiningTime,
		pc.transactionProcessingTime,
		pc.blockSize,
		pc.blockValidationTime,
		pc.networkLatency,
	)

	return pc
}

// StartHTTPServer запускает HTTP сервер для метрик
func (pc *PrometheusCollector) StartHTTPServer(addr string) error {
	http.Handle("/metrics", promhttp.HandlerFor(pc.registry, promhttp.HandlerOpts{}))

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Failed to start metrics server: %v\n", err)
		}
	}()

	fmt.Printf("Prometheus metrics server started on %s/metrics\n", addr)
	return nil
}

// Методы для работы с метриками

// IncBlocksMined увеличивает счетчик добытых блоков
func (pc *PrometheusCollector) IncBlocksMined(nodeID, difficulty string) {
	pc.blocksMinedTotal.WithLabelValues(nodeID, difficulty).Inc()
}

// AddTransactionsProcessed увеличивает счетчик обработанных транзакций
func (pc *PrometheusCollector) AddTransactionsProcessed(nodeID, txType string, count int) {
	pc.transactionsProcessedTotal.WithLabelValues(nodeID, txType).Add(float64(count))
}

// IncErrors увеличивает счетчик ошибок
func (pc *PrometheusCollector) IncErrors(nodeID, errorType string) {
	pc.errorsTotal.WithLabelValues(nodeID, errorType).Inc()
}

// SetBlockchainHeight устанавливает высоту блокчейна
func (pc *PrometheusCollector) SetBlockchainHeight(nodeID string, height int64) {
	pc.blockchainHeight.WithLabelValues(nodeID).Set(float64(height))
}

// SetUTXOCount устанавливает количество UTXO
func (pc *PrometheusCollector) SetUTXOCount(nodeID string, count int64) {
	pc.utxoCount.WithLabelValues(nodeID).Set(float64(count))
}

// SetActiveConnections устанавливает количество активных соединений
func (pc *PrometheusCollector) SetActiveConnections(nodeID, connType string, count int) {
	pc.activeConnections.WithLabelValues(nodeID, connType).Set(float64(count))
}

// SetMemoryUsage устанавливает использование памяти
func (pc *PrometheusCollector) SetMemoryUsage(nodeID, component string, bytes int64) {
	pc.memoryUsage.WithLabelValues(nodeID, component).Set(float64(bytes))
}

// ObserveBlockMiningTime записывает время майнинга блока
func (pc *PrometheusCollector) ObserveBlockMiningTime(nodeID, difficulty string, duration time.Duration) {
	pc.blockMiningTime.WithLabelValues(nodeID, difficulty).Observe(duration.Seconds())
}

// ObserveTransactionProcessingTime записывает время обработки транзакции
func (pc *PrometheusCollector) ObserveTransactionProcessingTime(nodeID, txType string, duration time.Duration) {
	pc.transactionProcessingTime.WithLabelValues(nodeID, txType).Observe(duration.Seconds())
}

// ObserveBlockSize записывает размер блока
func (pc *PrometheusCollector) ObserveBlockSize(nodeID string, size int64) {
	pc.blockSize.WithLabelValues(nodeID).Observe(float64(size))
}

// ObserveBlockValidationTime записывает время валидации блока
func (pc *PrometheusCollector) ObserveBlockValidationTime(nodeID string, duration time.Duration) {
	pc.blockValidationTime.WithLabelValues(nodeID).Observe(duration.Seconds())
}

// ObserveNetworkLatency записывает задержку сети
func (pc *PrometheusCollector) ObserveNetworkLatency(nodeID, peerID string, duration time.Duration) {
	pc.networkLatency.WithLabelValues(nodeID, peerID).Observe(duration.Seconds())
}

// GetRegistry возвращает реестр Prometheus
func (pc *PrometheusCollector) GetRegistry() *prometheus.Registry {
	return pc.registry
}

// BlockchainMetrics представляет метрики блокчейна
type BlockchainMetrics struct {
	collector *PrometheusCollector
	nodeID    string
}

// NewBlockchainMetrics создает новые метрики блокчейна
func NewBlockchainMetrics(collector *PrometheusCollector, nodeID string) *BlockchainMetrics {
	return &BlockchainMetrics{
		collector: collector,
		nodeID:    nodeID,
	}
}

// OnBlockMined вызывается при добыче блока
func (bm *BlockchainMetrics) OnBlockMined(difficulty int, miningTime time.Duration, blockSize int64) {
	bm.collector.IncBlocksMined(bm.nodeID, fmt.Sprintf("%d", difficulty))
	bm.collector.ObserveBlockMiningTime(bm.nodeID, fmt.Sprintf("%d", difficulty), miningTime)
	bm.collector.ObserveBlockSize(bm.nodeID, blockSize)
}

// OnTransactionProcessed вызывается при обработке транзакции
func (bm *BlockchainMetrics) OnTransactionProcessed(txType string, processingTime time.Duration) {
	bm.collector.AddTransactionsProcessed(bm.nodeID, txType, 1)
	bm.collector.ObserveTransactionProcessingTime(bm.nodeID, txType, processingTime)
}

// OnBlockchainUpdated вызывается при обновлении блокчейна
func (bm *BlockchainMetrics) OnBlockchainUpdated(height int64, utxoCount int64) {
	bm.collector.SetBlockchainHeight(bm.nodeID, height)
	bm.collector.SetUTXOCount(bm.nodeID, utxoCount)
}

// OnError вызывается при ошибке
func (bm *BlockchainMetrics) OnError(errorType string) {
	bm.collector.IncErrors(bm.nodeID, errorType)
}

// OnBlockValidated вызывается при валидации блока
func (bm *BlockchainMetrics) OnBlockValidated(validationTime time.Duration) {
	bm.collector.ObserveBlockValidationTime(bm.nodeID, validationTime)
}

// NetworkMetrics представляет метрики сети
type NetworkMetrics struct {
	collector *PrometheusCollector
	nodeID    string
}

// NewNetworkMetrics создает новые метрики сети
func NewNetworkMetrics(collector *PrometheusCollector, nodeID string) *NetworkMetrics {
	return &NetworkMetrics{
		collector: collector,
		nodeID:    nodeID,
	}
}

// OnConnectionEstablished вызывается при установке соединения
func (nm *NetworkMetrics) OnConnectionEstablished(connType string) {
	// Получаем текущее количество соединений и увеличиваем на 1
	// В реальной реализации здесь была бы более сложная логика
	nm.collector.SetActiveConnections(nm.nodeID, connType, 1)
}

// OnConnectionClosed вызывается при закрытии соединения
func (nm *NetworkMetrics) OnConnectionClosed(connType string) {
	// Получаем текущее количество соединений и уменьшаем на 1
	// В реальной реализации здесь была бы более сложная логика
	nm.collector.SetActiveConnections(nm.nodeID, connType, 0)
}

// OnNetworkLatency вызывается при измерении задержки сети
func (nm *NetworkMetrics) OnNetworkLatency(peerID string, latency time.Duration) {
	nm.collector.ObserveNetworkLatency(nm.nodeID, peerID, latency)
}

// OnError вызывается при ошибке сети
func (nm *NetworkMetrics) OnError(errorType string) {
	nm.collector.IncErrors(nm.nodeID, errorType)
}
