//go:build metrics_usage
// +build metrics_usage

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
	// Создаем сборщик метрик
	collector := metrics.NewMetricsCollector()

	// Создаем метрики для блокчейна
	blocksMined := collector.CreateCounter("blocks_mined_total", map[string]string{"type": "blockchain"})
	transactionsProcessed := collector.CreateCounter("transactions_processed_total", map[string]string{"type": "blockchain"})
	blockchainHeight := collector.CreateGauge("blockchain_height", map[string]string{"type": "blockchain"})
	utxoCount := collector.CreateGauge("utxo_count", map[string]string{"type": "blockchain"})
	blockMiningTime := collector.CreateHistogram("block_mining_time_seconds", map[string]string{"type": "mining"}, []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0})
	transactionProcessingTime := collector.CreateSummary("transaction_processing_time_nanoseconds", map[string]string{"type": "processing"})

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet", "error", err)
		return
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Устанавливаем начальную высоту
	blockchainHeight.Set(0)

	// Обрабатываем genesis блок
	genesisBlock := bc.GetBlockByHeight(0)
	transactionsProcessed.Add(int64(len(genesisBlock.Transactions)))
	blocksMined.Inc()
	blockchainHeight.Set(1)

	// Создаем несколько блоков с метриками
	for i := 1; i <= 5; i++ {
		// Измеряем время майнинга
		timer := metrics.NewTimer()

		// Создаем coinbase транзакцию
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)

		// Измеряем время обработки транзакции
		txTimer := metrics.NewTimer()
		// В реальной реализации здесь была бы обработка транзакции
		time.Sleep(1 * time.Millisecond) // Имитируем обработку
		txTimer.Observe(transactionProcessingTime)

		// Создаем блок
		block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, genesisBlock.Hash, int64(i), 0)

		// Добавляем блок в блокчейн
		err = bc.AddBlock(block)
		if err != nil {
			slog.Error("Failed to add block", "error", err, "height", i)
			continue
		}

		// Обновляем метрики
		blocksMined.Inc()
		transactionsProcessed.Add(int64(len(block.Transactions)))
		blockchainHeight.Set(int64(i + 1))

		// Обновляем время майнинга
		timer.ObserveHistogram(blockMiningTime)

		// Обновляем количество UTXO
		utxoCount.Set(int64(len(bc.UTXOSet.UTXOs)))

		slog.Info("Block added", "height", i, "transactions", len(block.Transactions))
	}

	// Выводим статистику метрик
	fmt.Println("\n=== Blockchain Metrics ===")
	stats := collector.GetStats()
	for name, metric := range stats {
		metricData := metric.(map[string]interface{})
		fmt.Printf("%s: %v (type: %v)\n", name, metricData["value"], metricData["type"])
	}

	// Выводим детальную статистику
	fmt.Println("\n=== Detailed Metrics ===")

	// Счетчики
	fmt.Printf("Blocks mined: %d\n", blocksMined.GetValue())
	fmt.Printf("Transactions processed: %d\n", transactionsProcessed.GetValue())

	// Датчики
	fmt.Printf("Blockchain height: %d\n", blockchainHeight.GetValue())
	fmt.Printf("UTXO count: %d\n", utxoCount.GetValue())

	// Гистограмма
	histogramValue := blockMiningTime.GetValue().(map[string]interface{})
	fmt.Printf("Block mining time histogram: %v\n", histogramValue)

	// Сводка
	summaryValue := transactionProcessingTime.GetValue().(map[string]interface{})
	fmt.Printf("Transaction processing time summary: %v\n", summaryValue)

	// Выводим статистику блокчейна
	fmt.Println("\n=== Blockchain Stats ===")
	bcStats := bc.GetStats()
	for key, value := range bcStats {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Демонстрируем работу таймера
	fmt.Println("\n=== Timer Demo ===")
	timer := metrics.NewTimer()
	time.Sleep(100 * time.Millisecond)
	elapsed := timer.Elapsed()
	fmt.Printf("Elapsed time: %v\n", elapsed)

	// Демонстрируем работу с метриками
	fmt.Println("\n=== Metric Operations Demo ===")

	// Увеличиваем счетчик
	blocksMined.Add(10)
	fmt.Printf("Blocks mined after adding 10: %d\n", blocksMined.GetValue())

	// Устанавливаем датчик
	blockchainHeight.Set(100)
	fmt.Printf("Blockchain height after setting to 100: %d\n", blockchainHeight.GetValue())

	// Добавляем наблюдения в гистограмму
	for i := 0; i < 10; i++ {
		blockMiningTime.Observe(float64(i) * 0.1)
	}

	// Добавляем наблюдения в сводку
	for i := 0; i < 10; i++ {
		transactionProcessingTime.Observe(int64(i * 1000000)) // наносекунды
	}

	// Выводим финальную статистику
	fmt.Println("\n=== Final Metrics ===")
	finalStats := collector.GetStats()
	for name, metric := range finalStats {
		metricData := metric.(map[string]interface{})
		fmt.Printf("%s: %v\n", name, metricData["value"])
	}

	fmt.Println("\nMetrics demo completed successfully!")
}
