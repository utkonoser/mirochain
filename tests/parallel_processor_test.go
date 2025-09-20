package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/parallel"
	"mirochain/internal/wallet"
)

// TestTransactionProcessorCreation тестирует создание процессора транзакций
func TestTransactionProcessorCreation(t *testing.T) {
	processor := parallel.NewTransactionProcessor(4, 100)

	if processor == nil {
		t.Fatal("Processor should not be nil")
	}

	stats := processor.GetStats()
	if stats["worker_count"] != 4 {
		t.Errorf("Expected worker count 4, got %v", stats["worker_count"])
	}

	if stats["queue_size"] != 100 {
		t.Errorf("Expected queue size 100, got %v", stats["queue_size"])
	}

	if stats["running"] != false {
		t.Error("Processor should not be running initially")
	}

	t.Logf("Transaction processor created successfully")
}

// TestTransactionProcessorStartStop тестирует запуск и остановку процессора
func TestTransactionProcessorStartStop(t *testing.T) {
	processor := parallel.NewTransactionProcessor(2, 50)

	// Создаем UTXO набор
	utxoSet := blockchain.NewUTXOSet()

	// Запускаем процессор
	err := processor.Start(utxoSet)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}

	// Проверяем, что процессор запущен
	stats := processor.GetStats()
	if stats["running"] != true {
		t.Error("Processor should be running")
	}

	// Останавливаем процессор
	err = processor.Stop()
	if err != nil {
		t.Fatalf("Failed to stop processor: %v", err)
	}

	// Проверяем, что процессор остановлен
	stats = processor.GetStats()
	if stats["running"] != false {
		t.Error("Processor should not be running after stop")
	}

	t.Logf("Transaction processor start/stop test completed successfully")
}

// TestTransactionProcessorSingleTransaction тестирует обработку одной транзакции
func TestTransactionProcessorSingleTransaction(t *testing.T) {
	processor := parallel.NewTransactionProcessor(2, 50)

	// Создаем UTXO набор
	utxoSet := blockchain.NewUTXOSet()

	// Запускаем процессор
	err := processor.Start(utxoSet)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer processor.Stop()

	// Создаем тестовую транзакцию
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем coinbase транзакцию
	coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)

	// Обрабатываем транзакцию
	result, err := processor.ProcessTransaction(coinbaseTx, utxoSet, 1)
	if err != nil {
		t.Fatalf("Failed to process transaction: %v", err)
	}

	// Проверяем результат
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Transaction != coinbaseTx {
		t.Error("Result transaction should match input transaction")
	}

	if result.ProcessTime <= 0 {
		t.Error("Process time should be positive")
	}

	t.Logf("Single transaction processing test completed successfully. Process time: %v", result.ProcessTime)
}

// TestTransactionProcessorBatch тестирует обработку нескольких транзакций
func TestTransactionProcessorBatch(t *testing.T) {
	processor := parallel.NewTransactionProcessor(4, 100)

	// Создаем UTXO набор
	utxoSet := blockchain.NewUTXOSet()

	// Запускаем процессор
	err := processor.Start(utxoSet)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer processor.Stop()

	// Создаем несколько транзакций
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	transactions := make([]*blockchain.Transaction, 5)
	for i := 0; i < 5; i++ {
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
		transactions[i] = coinbaseTx
	}

	// Обрабатываем транзакции
	results, err := processor.ProcessTransactionsBatch(transactions, utxoSet, 1)
	if err != nil {
		t.Fatalf("Failed to process transactions batch: %v", err)
	}

	// Проверяем результаты
	if len(results) != len(transactions) {
		t.Errorf("Expected %d results, got %d", len(transactions), len(results))
	}

	for i, result := range results {
		if result == nil {
			t.Errorf("Result %d should not be nil", i)
			continue
		}

		if result.Transaction != transactions[i] {
			t.Errorf("Result %d transaction should match input transaction", i)
		}

		if result.ProcessTime <= 0 {
			t.Errorf("Result %d process time should be positive", i)
		}
	}

	t.Logf("Batch transaction processing test completed successfully. Processed %d transactions", len(results))
}

// TestTransactionProcessorConcurrent тестирует конкурентную обработку
func TestTransactionProcessorConcurrent(t *testing.T) {
	processor := parallel.NewTransactionProcessor(4, 100)

	// Создаем UTXO набор
	utxoSet := blockchain.NewUTXOSet()

	// Запускаем процессор
	err := processor.Start(utxoSet)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer processor.Stop()

	// Создаем несколько транзакций
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем горутины для конкурентной обработки
	results := make(chan *parallel.TransactionResult, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
			result, err := processor.ProcessTransaction(coinbaseTx, utxoSet, 1)
			if err != nil {
				t.Errorf("Failed to process transaction %d: %v", i, err)
				return
			}
			results <- result
		}(i)
	}

	// Собираем результаты
	collectedResults := make([]*parallel.TransactionResult, 0, 10)
	timeout := time.After(5 * time.Second)

	for i := 0; i < 10; i++ {
		select {
		case result := <-results:
			collectedResults = append(collectedResults, result)
		case <-timeout:
			t.Fatalf("Timeout waiting for result %d", i)
		}
	}

	// Проверяем результаты
	if len(collectedResults) != 10 {
		t.Errorf("Expected 10 results, got %d", len(collectedResults))
	}

	for i, result := range collectedResults {
		if result == nil {
			t.Errorf("Result %d should not be nil", i)
			continue
		}

		if result.ProcessTime <= 0 {
			t.Errorf("Result %d process time should be positive", i)
		}
	}

	t.Logf("Concurrent transaction processing test completed successfully. Processed %d transactions", len(collectedResults))
}

// TestTransactionProcessorStats тестирует статистику процессора
func TestTransactionProcessorStats(t *testing.T) {
	processor := parallel.NewTransactionProcessor(2, 50)

	// Создаем UTXO набор
	utxoSet := blockchain.NewUTXOSet()

	// Запускаем процессор
	err := processor.Start(utxoSet)
	if err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer processor.Stop()

	// Обрабатываем несколько транзакций
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	for i := 0; i < 5; i++ {
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
		_, err := processor.ProcessTransaction(coinbaseTx, utxoSet, 1)
		if err != nil {
			t.Fatalf("Failed to process transaction %d: %v", i, err)
		}
	}

	// Получаем статистику
	stats := processor.GetStats()

	// Проверяем основные поля
	if stats["running"] != true {
		t.Error("Processor should be running")
	}

	if stats["worker_count"] != 2 {
		t.Errorf("Expected worker count 2, got %v", stats["worker_count"])
	}

	if stats["queue_size"] != 50 {
		t.Errorf("Expected queue size 50, got %v", stats["queue_size"])
	}

	// Проверяем статистику воркеров
	workers, ok := stats["workers"].([]map[string]interface{})
	if !ok {
		t.Error("Workers stats should be available")
	}

	if len(workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(workers))
	}

	// Проверяем, что хотя бы один воркер обработал транзакции
	totalProcessed := int64(0)
	for i, worker := range workers {
		processedCount, ok := worker["processed_count"].(int64)
		if !ok {
			t.Errorf("Worker %d processed count should be int64", i)
			continue
		}
		totalProcessed += processedCount
	}

	if totalProcessed == 0 {
		t.Error("At least one worker should have processed transactions")
	}

	t.Logf("Processor stats test completed successfully. Total processed: %d", totalProcessed)
}

// TestTransactionProcessorIntegration тестирует интеграцию процессора
func TestTransactionProcessorIntegration(t *testing.T) {
	t.Run("Creation", TestTransactionProcessorCreation)
	t.Run("StartStop", TestTransactionProcessorStartStop)
	t.Run("SingleTransaction", TestTransactionProcessorSingleTransaction)
	t.Run("Batch", TestTransactionProcessorBatch)
	t.Run("Concurrent", TestTransactionProcessorConcurrent)
	t.Run("Stats", TestTransactionProcessorStats)
}
