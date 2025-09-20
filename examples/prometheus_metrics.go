//go:build prometheus_metrics
// +build prometheus_metrics

package main

import (
	"fmt"
	"log/slog"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/metrics"
	"mirochain/internal/wallet"
)

func main() {
	// Создаем Prometheus коллектор
	prometheusCollector := metrics.NewPrometheusCollector()

	// Запускаем HTTP сервер для метрик
	err := prometheusCollector.StartHTTPServer(":8080")
	if err != nil {
		slog.Error("Failed to start metrics server", "error", err)
		return
	}

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet", "error", err)
		return
	}

	nodeID := "node_001"

	// Создаем метрики блокчейна
	blockchainMetrics := metrics.NewBlockchainMetrics(prometheusCollector, nodeID)
	networkMetrics := metrics.NewNetworkMetrics(prometheusCollector, nodeID)

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Устанавливаем начальные метрики
	blockchainMetrics.OnBlockchainUpdated(0, int64(len(bc.UTXOSet.UTXOs)))

	// Обрабатываем genesis блок
	genesisBlock := bc.GetBlockByHeight(0)
	blockchainMetrics.OnTransactionProcessed("coinbase", 0)
	genesisData, _ := genesisBlock.Serialize()
	blockchainMetrics.OnBlockMined(0, 0, int64(len(genesisData)))

	// Создаем несколько блоков с метриками
	for i := 1; i <= 5; i++ {
		// Измеряем время майнинга
		miningStart := time.Now()

		// Создаем coinbase транзакцию
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)

		// Измеряем время обработки транзакции
		txStart := time.Now()
		// В реальной реализации здесь была бы обработка транзакции
		time.Sleep(1 * time.Millisecond) // Имитируем обработку
		blockchainMetrics.OnTransactionProcessed("coinbase", time.Since(txStart))

		// Создаем блок
		block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, genesisBlock.Hash, int64(i), 0)

		// Измеряем время валидации блока
		validationStart := time.Now()
		// В реальной реализации здесь была бы валидация блока
		time.Sleep(100 * time.Microsecond) // Имитируем валидацию
		blockchainMetrics.OnBlockValidated(time.Since(validationStart))

		// Добавляем блок в блокчейн
		err = bc.AddBlock(block)
		if err != nil {
			slog.Error("Failed to add block", "error", err, "height", i)
			blockchainMetrics.OnError("block_validation")
			continue
		}

		// Обновляем метрики
		miningTime := time.Since(miningStart)
		blockData, _ := block.Serialize()
		blockSize := int64(len(blockData))
		blockchainMetrics.OnBlockMined(0, miningTime, blockSize)
		blockchainMetrics.OnBlockchainUpdated(int64(i+1), int64(len(bc.UTXOSet.UTXOs)))

		slog.Info("Block added with metrics",
			"height", i,
			"transactions", len(block.Transactions),
			"mining_time", miningTime,
			"block_size", blockSize)
	}

	// Имитируем сетевые события
	networkMetrics.OnConnectionEstablished("tcp")
	networkMetrics.OnConnectionEstablished("udp")
	networkMetrics.OnNetworkLatency("peer_001", 10*time.Millisecond)
	networkMetrics.OnNetworkLatency("peer_002", 15*time.Millisecond)

	// Имитируем ошибки сети
	networkMetrics.OnError("connection_timeout")
	networkMetrics.OnError("invalid_message")

	// Имитируем использование памяти
	prometheusCollector.SetMemoryUsage(nodeID, "blockchain", 1024*1024)
	prometheusCollector.SetMemoryUsage(nodeID, "utxo", 512*1024)
	prometheusCollector.SetMemoryUsage(nodeID, "cache", 256*1024)

	// Выводим статистику
	fmt.Println("\n=== Prometheus Metrics Demo ===")
	fmt.Printf("Metrics server running on http://localhost:8080/metrics\n")
	fmt.Printf("Node ID: %s\n", nodeID)

	// Выводим статистику блокчейна
	bcStats := bc.GetStats()
	fmt.Println("\n=== Blockchain Stats ===")
	for key, value := range bcStats {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Демонстрируем работу с метриками
	fmt.Println("\n=== Metrics Operations Demo ===")

	// Увеличиваем счетчики
	prometheusCollector.IncBlocksMined(nodeID, "5")
	prometheusCollector.AddTransactionsProcessed(nodeID, "coinbase", 10)
	prometheusCollector.IncErrors(nodeID, "validation")

	// Устанавливаем датчики
	prometheusCollector.SetBlockchainHeight(nodeID, 100)
	prometheusCollector.SetUTXOCount(nodeID, 500)
	prometheusCollector.SetActiveConnections(nodeID, "tcp", 5)

	// Записываем наблюдения
	prometheusCollector.ObserveBlockMiningTime(nodeID, "5", 1*time.Second)
	prometheusCollector.ObserveTransactionProcessingTime(nodeID, "coinbase", 100*time.Millisecond)
	prometheusCollector.ObserveBlockSize(nodeID, 1024)
	prometheusCollector.ObserveBlockValidationTime(nodeID, 50*time.Millisecond)
	prometheusCollector.ObserveNetworkLatency(nodeID, "peer_001", 10*time.Millisecond)

	fmt.Println("Metrics operations completed successfully!")

	// Ждем некоторое время, чтобы можно было проверить метрики
	fmt.Println("\nWaiting for 30 seconds to check metrics at http://localhost:8080/metrics...")
	time.Sleep(30 * time.Second)

	fmt.Println("Prometheus metrics demo completed!")
}
