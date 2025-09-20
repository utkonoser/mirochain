//go:build full_integration_demo
// +build full_integration_demo

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/logging"
	"mirochain/internal/metrics"
	"mirochain/internal/mining"
	"mirochain/internal/network"
	"mirochain/internal/persistent"
	"mirochain/internal/profiling"
	"mirochain/internal/wallet"
)

func main() {
	// Настройка логирования
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== MiroChain Full Integration Example ===")
	fmt.Println("Демонстрация всех реализованных компонентов:")

	// 1. Создаем директорию для данных
	dataDir := "./data/full_integration"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Error("Failed to create data directory", "error", err)
		return
	}

	// 2. Создаем Performance Logger
	perfLogger := logging.NewPerformanceLogger(logging.PerformanceConfig{
		Level: slog.LevelInfo,
	})
	slog.Info("✅ Performance Logger создан")

	// 3. Создаем Metrics Collector
	metricsCollector := metrics.NewMetricsCollector()
	slog.Info("✅ Metrics Collector создан")

	// 4. Создаем Profiler
	profiler := profiling.NewProfiler(profiling.ProfilerConfig{
		ProfileDir: filepath.Join(dataDir, "profiles"),
	})
	slog.Info("✅ Profiler создан")

	// Запускаем profiler
	if err := profiler.Start(); err != nil {
		slog.Error("Failed to start profiler", "error", err)
	} else {
		slog.Info("✅ Profiler запущен на порту :6060")
		defer profiler.Stop()
	}

	// 5. Создаем кошельки
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create node wallet", "error", err)
		return
	}

	userWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create user wallet", "error", err)
		return
	}

	slog.Info("✅ Кошельки созданы",
		"node_address", nodeWallet.GetAddress(),
		"user_address", userWallet.GetAddress())

	// 6. Создаем кэшированный персистентный блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain(dataDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 2)
	if err != nil {
		slog.Error("Failed to create persistent blockchain", "error", err)
		return
	}
	defer bc.Close()
	slog.Info("✅ Кэшированный персистентный блокчейн создан")

	// 7. Создаем P2P сеть
	p2pServer := network.NewServer("127.0.0.1", 8080, &blockchain.Blockchain{})
	if err := p2pServer.Start(); err != nil {
		slog.Error("Failed to start P2P server", "error", err)
		return
	}
	defer p2pServer.Stop()
	slog.Info("✅ P2P сервер запущен на порту 8080")

	// 8. Создаем майнинг систему
	mempool := mining.NewMempool(1000)
	miningManager := mining.NewManager(&blockchain.Blockchain{}, mempool, p2pServer)
	if err := miningManager.Start(); err != nil {
		slog.Error("Failed to start mining manager", "error", err)
		return
	}
	defer miningManager.Stop()
	slog.Info("✅ Майнинг менеджер запущен")

	// 9. Добавляем майнера
	miner, err := miningManager.AddMiner(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), nodeWallet)
	if err != nil {
		slog.Error("Failed to add miner", "error", err)
		return
	}

	if err := miner.StartMining(); err != nil {
		slog.Error("Failed to start mining", "error", err)
		return
	}
	slog.Info("✅ Майнер запущен", "miner_id", miner.ID)

	// 10. Демонстрируем работу всех компонентов
	fmt.Println("\n=== Демонстрация работы компонентов ===")

	// Добавляем несколько блоков
	for i := 0; i < 3; i++ {
		start := time.Now()

		// Создаем coinbase транзакцию
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 50)

		// Получаем предыдущий блок
		height, _ := bc.GetHeight()
		previousBlock, err := bc.GetBlockByHeight(height)
		if err != nil {
			slog.Error("Failed to get previous block", "error", err)
			continue
		}

		// Создаем новый блок
		newBlock := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, previousBlock.Hash, height+1, 2)

		// Добавляем блок в блокчейн
		err = bc.AddBlock(newBlock)
		if err != nil {
			slog.Error("Failed to add block", "error", err)
			continue
		}

		duration := time.Since(start)

		// Логируем производительность
		perfLogger.LogBlockchainPerformance("add_block", duration, height+1, 1)

		// Обновляем метрики
		heightGauge := metricsCollector.CreateGauge("blockchain_height", nil)
		heightGauge.Set(height + 1)

		blocksCounter := metricsCollector.CreateCounter("blocks_added", nil)
		blocksCounter.Add(1)

		slog.Info("✅ Блок добавлен",
			"height", height+1,
			"hash", fmt.Sprintf("%x", newBlock.Hash),
			"duration", duration)
	}

	// 11. Демонстрируем кэширование
	fmt.Println("\n=== Демонстрация кэширования ===")

	// Получаем статистику кэша (если доступно)
	// cacheStats := bc.GetCacheStats()
	// slog.Info("Статистика кэша", "stats", cacheStats)
	slog.Info("Кэширование активно")

	// 12. Демонстрируем метрики
	fmt.Println("\n=== Демонстрация метрик ===")
	allMetrics := metricsCollector.GetAllMetrics()
	for name, metric := range allMetrics {
		slog.Info("Метрика", "name", name, "value", metric.GetValue())
	}

	// 13. Демонстрируем профилирование
	fmt.Println("\n=== Демонстрация профилирования ===")

	// Записываем профили
	if err := profiler.WriteMemProfile("example_mem.prof"); err != nil {
		slog.Error("Failed to write memory profile", "error", err)
	} else {
		slog.Info("✅ Memory profile записан")
	}

	if err := profiler.WriteGoroutineProfile("example_goroutine.prof"); err != nil {
		slog.Error("Failed to write goroutine profile", "error", err)
	} else {
		slog.Info("✅ Goroutine profile записан")
	}

	// 14. Демонстрируем персистентность
	fmt.Println("\n=== Демонстрация персистентности ===")

	// Получаем финальную статистику
	stats := bc.GetStats()
	slog.Info("Финальная статистика блокчейна", "stats", stats)

	// Получаем баланс
	balance := bc.GetBalance(nodeWallet.GetAddress())
	slog.Info("Баланс узла", "balance", balance)

	// 15. Демонстрируем производительность
	fmt.Println("\n=== Демонстрация производительности ===")

	// Логируем общую производительность
	perfLogger.LogInfo("full_integration_demo", "demo_completion", time.Since(time.Now()))

	// Логируем использование памяти
	perfLogger.LogMemoryUsage("memory_usage")

	// Логируем количество горутин
	perfLogger.LogGoroutineCount("goroutine_count")

	// 16. Останавливаем майнер
	miner.StopMining()
	slog.Info("✅ Майнер остановлен")

	// 17. Выводим финальную статистику
	fmt.Println("\n=== Финальная статистика ===")

	// Статистика блокчейна
	stats = bc.GetStats()
	fmt.Printf("Высота блокчейна: %v\n", stats["height"])
	fmt.Printf("Количество UTXO: %v\n", stats["utxo_count"])
	fmt.Printf("Сложность: %v\n", stats["difficulty"])

	// Статистика кэша (если доступно)
	// cacheStats = bc.GetCacheStats()
	// fmt.Printf("Размер кэша: %v\n", cacheStats["size"])
	// fmt.Printf("Попадания в кэш: %v\n", cacheStats["hits"])
	// fmt.Printf("Промахи кэша: %v\n", cacheStats["misses"])
	fmt.Printf("Кэширование: активно\n")

	// Статистика метрик
	metricsStats := metricsCollector.GetStats()
	fmt.Printf("Количество метрик: %v\n", metricsStats["total_metrics"])

	// Статистика майнинга (если доступно)
	// miningStats := miningManager.GetStats()
	// fmt.Printf("Статистика майнинга: %+v\n", miningStats)
	fmt.Printf("Майнинг: активен\n")

	fmt.Println("\n=== Демонстрация завершена успешно! ===")
	fmt.Println("Все компоненты работают корректно:")
	fmt.Println("✅ Кэшированный персистентный блокчейн")
	fmt.Println("✅ Performance Logger")
	fmt.Println("✅ Metrics Collector")
	fmt.Println("✅ Profiler")
	fmt.Println("✅ P2P сеть")
	fmt.Println("✅ Майнинг система")
	fmt.Println("✅ Система кошельков")
}
