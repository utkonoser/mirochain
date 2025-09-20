//go:build transaction_processor_demo
// +build transaction_processor_demo

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/wallet"
)

func main() {
	// Настройка логирования
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== TransactionProcessor Integration Demo ===")
	fmt.Println("Демонстрация параллельной обработки транзакций:")

	// 1. Создаем кошельки
	walletManager := wallet.NewWalletManager()

	wallet1, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet1", "error", err)
		return
	}

	wallet2, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet2", "error", err)
		return
	}

	wallet3, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet3", "error", err)
		return
	}

	slog.Info("✅ Кошельки созданы",
		"wallet1", wallet1.GetAddress(),
		"wallet2", wallet2.GetAddress(),
		"wallet3", wallet3.GetAddress())

	// 2. Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 2)
	slog.Info("✅ Блокчейн создан", "height", bc.GetHeight())

	// 3. Создаем mempool с TransactionProcessor
	mempool := mining.NewMempool(100)
	slog.Info("✅ Mempool создан")

	// 4. Запускаем TransactionProcessor для mempool
	if err := mempool.StartProcessor(); err != nil {
		slog.Error("Failed to start mempool processor", "error", err)
		return
	}
	defer mempool.StopProcessor()
	slog.Info("✅ TransactionProcessor для mempool запущен")

	// 5. Создаем несколько транзакций
	transactions := make([]*blockchain.Transaction, 5)

	for i := 0; i < 5; i++ {
		// Создаем транзакцию от wallet1 к wallet2
		tx, err := bc.CreateTransaction(
			wallet1.GetAddress(),
			wallet2.GetAddress(),
			100,
			wallet1.GetPrivateKeyBytes(),
		)
		if err != nil {
			slog.Error("Failed to create transaction", "error", err, "index", i)
			continue
		}
		transactions[i] = tx
	}

	slog.Info("✅ Создано 5 транзакций")

	// 6. Демонстрируем асинхронную обработку одной транзакции
	fmt.Println("\n=== Асинхронная обработка одной транзакции ===")

	utxoSet := blockchain.NewUTXOSet()
	start := time.Now()

	resultChan := mempool.ValidateTransactionAsync(transactions[0], utxoSet)

	select {
	case result := <-resultChan:
		duration := time.Since(start)
		slog.Info("Транзакция обработана",
			"valid", result.Valid,
			"error", result.Error,
			"duration", duration,
			"process_time", result.ProcessTime)
	case <-time.After(5 * time.Second):
		slog.Error("Timeout waiting for transaction result")
	}

	// 7. Демонстрируем пакетную обработку транзакций
	fmt.Println("\n=== Пакетная обработка транзакций ===")

	start = time.Now()
	results := mempool.ValidateTransactionsBatch(transactions, utxoSet)
	duration := time.Since(start)

	validCount := 0
	errorCount := 0

	for i, result := range results {
		if result.Valid {
			validCount++
		} else {
			errorCount++
			slog.Warn("Transaction validation failed",
				"index", i,
				"error", result.Error)
		}
	}

	slog.Info("Пакетная обработка завершена",
		"total", len(results),
		"valid", validCount,
		"errors", errorCount,
		"total_duration", duration)

	// 8. Демонстрируем работу с майнинг менеджером
	fmt.Println("\n=== Интеграция с майнинг менеджером ===")

	// Создаем отдельный mempool для менеджера
	managerMempool := mining.NewMempool(100)

	// Создаем майнинг менеджер (без P2P сервера для простоты)
	miningManager := mining.NewManager(bc, managerMempool, nil)

	// Запускаем менеджер
	if err := miningManager.Start(); err != nil {
		slog.Error("Failed to start mining manager", "error", err)
		return
	}
	defer miningManager.Stop()
	slog.Info("✅ Майнинг менеджер запущен")

	// 9. Демонстрируем асинхронную обработку через менеджер
	fmt.Println("\n=== Асинхронная обработка через менеджер ===")

	start = time.Now()
	managerResultChan := miningManager.ProcessTransactionAsync(transactions[1], utxoSet)

	select {
	case result := <-managerResultChan:
		duration = time.Since(start)
		slog.Info("Транзакция обработана через менеджер",
			"valid", result.Valid,
			"error", result.Error,
			"duration", duration,
			"process_time", result.ProcessTime)
	case <-time.After(5 * time.Second):
		slog.Error("Timeout waiting for manager transaction result")
	}

	// 10. Демонстрируем пакетную обработку через менеджер
	fmt.Println("\n=== Пакетная обработка через менеджер ===")

	start = time.Now()
	managerResults := miningManager.ProcessTransactionsBatch(transactions[2:], utxoSet)
	duration = time.Since(start)

	validCount = 0
	errorCount = 0

	for i, result := range managerResults {
		if result.Valid {
			validCount++
		} else {
			errorCount++
			slog.Warn("Manager transaction validation failed",
				"index", i,
				"error", result.Error)
		}
	}

	slog.Info("Пакетная обработка через менеджер завершена",
		"total", len(managerResults),
		"valid", validCount,
		"errors", errorCount,
		"total_duration", duration)

	// 11. Показываем статистику TransactionProcessor
	fmt.Println("\n=== Статистика TransactionProcessor ===")

	// Статистика mempool processor
	mempoolStats := mempool.GetProcessorStats()
	slog.Info("Статистика mempool processor", "stats", mempoolStats)

	// Статистика manager processor
	managerStats := miningManager.GetTransactionProcessorStats()
	slog.Info("Статистика manager processor", "stats", managerStats)

	// 12. Демонстрируем производительность
	fmt.Println("\n=== Тест производительности ===")

	// Создаем много транзакций для теста производительности
	performanceTransactions := make([]*blockchain.Transaction, 20)

	for i := 0; i < 20; i++ {
		tx, err := bc.CreateTransaction(
			wallet1.GetAddress(),
			wallet2.GetAddress(),
			50,
			wallet1.GetPrivateKeyBytes(),
		)
		if err != nil {
			slog.Error("Failed to create performance transaction", "error", err, "index", i)
			continue
		}
		performanceTransactions[i] = tx
	}

	start = time.Now()
	performanceResults := mempool.ValidateTransactionsBatch(performanceTransactions, utxoSet)
	duration = time.Since(start)

	validCount = 0
	for _, result := range performanceResults {
		if result.Valid {
			validCount++
		}
	}

	slog.Info("Тест производительности завершен",
		"transactions", len(performanceResults),
		"valid", validCount,
		"total_duration", duration,
		"avg_per_transaction", duration/time.Duration(len(performanceResults)))

	fmt.Println("\n=== Демонстрация завершена успешно! ===")
	fmt.Println("TransactionProcessor успешно интегрирован в:")
	fmt.Println("✅ Mempool для параллельной валидации транзакций")
	fmt.Println("✅ Mining Manager для асинхронной обработки")
	fmt.Println("✅ Пакетная обработка транзакций")
	fmt.Println("✅ Статистика и мониторинг производительности")
}
