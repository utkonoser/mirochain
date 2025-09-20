package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mirochain/internal/api"
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
	// Параметры командной строки
	var (
		port            = flag.Int("port", 8080, "Port to run the node on")
		difficulty      = flag.Int("difficulty", 4, "Mining difficulty")
		miningEnabled   = flag.Bool("mining", true, "Enable mining")
		address         = flag.String("address", "", "Wallet address for mining rewards")
		peers           = flag.String("peers", "", "Comma-separated list of peer addresses")
		dataDir         = flag.String("data", "./data", "Data directory for persistent storage")
		enableProfiling = flag.Bool("profiling", false, "Enable profiling")
		enableMetrics   = flag.Bool("metrics", true, "Enable metrics collection")
	)
	flag.Parse()

	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Создаем performance logger
	perfLogger := logging.NewPerformanceLogger(logging.PerformanceConfig{
		Level: slog.LevelInfo,
	})

	// Создаем metrics collector
	var metricsCollector *metrics.MetricsCollector
	if *enableMetrics {
		metricsCollector = metrics.NewMetricsCollector()
	}

	// Создаем profiler
	var profiler *profiling.Profiler
	if *enableProfiling {
		profiler = profiling.NewProfiler(profiling.ProfilerConfig{
			ProfileDir: filepath.Join(*dataDir, "profiles"),
		})
	}

	slog.Info("Starting MiroChain node",
		"port", *port,
		"difficulty", *difficulty,
		"mining", *miningEnabled,
		"data_dir", *dataDir,
		"profiling", *enableProfiling,
		"metrics", *enableMetrics)

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet", "error", err)
		os.Exit(1)
	}

	// Используем адрес из параметров или созданный
	if *address == "" {
		*address = nodeWallet.GetAddress()
	}

	slog.Info("Node wallet created", "address", *address)

	// Создаем директорию для данных
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		slog.Error("Failed to create data directory", "error", err)
		os.Exit(1)
	}

	// Создаем кэшированный персистентный блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain(*dataDir, *address, nodeWallet.GetPublicKeyBytes(), *difficulty)
	if err != nil {
		slog.Error("Failed to create persistent blockchain", "error", err)
		os.Exit(1)
	}
	defer bc.Close()

	height, _ := bc.GetHeight()
	slog.Info("Blockchain initialized",
		"height", height,
		"difficulty", *difficulty)

	// Запускаем profiler если включен
	if profiler != nil {
		if err := profiler.Start(); err != nil {
			slog.Error("Failed to start profiler", "error", err)
		} else {
			slog.Info("Profiler started", "port", ":6060")
		}
		defer profiler.Stop()
	}

	// Создаем адаптер для совместимости с network.NewServer
	// Пока что создаем обычный блокчейн для P2P сервера
	// TODO: Обновить network.NewServer для поддержки интерфейсов
	bcAdapter := &blockchain.Blockchain{}

	// Создаем P2P сервер
	p2pServer := network.NewServer("0.0.0.0", *port, bcAdapter)

	// Запускаем P2P сервер
	err = p2pServer.Start()
	if err != nil {
		slog.Error("Failed to start P2P server", "error", err)
		os.Exit(1)
	}
	defer p2pServer.Stop()

	// Создаем P2P клиент
	p2pClient := network.NewClient(p2pServer)

	// Подключаемся к другим узлам, если указаны
	if *peers != "" {
		peerAddresses := strings.Split(*peers, ",")
		p2pClient.ConnectToPeers(peerAddresses)
	}

	// Создаем mempool и менеджер майнинга
	mempool := mining.NewMempool(1000) // Максимум 1000 транзакций в mempool
	miningManager := mining.NewManager(bcAdapter, mempool, p2pServer)

	// Запускаем менеджер майнинга
	err = miningManager.Start()
	if err != nil {
		slog.Error("Failed to start mining manager", "error", err)
		os.Exit(1)
	}
	defer miningManager.Stop()

	// Добавляем майнера, если включен майнинг
	if *miningEnabled {
		miner, err := miningManager.AddMiner(*address, nodeWallet.GetPublicKeyBytes(), nodeWallet)
		if err != nil {
			slog.Error("Failed to add miner", "error", err)
		} else {
			// Запускаем майнинг
			err = miner.StartMining()
			if err != nil {
				slog.Error("Failed to start mining", "error", err)
			} else {
				slog.Info("Mining started", "miner_id", miner.ID, "address", *address)
			}
		}
	}

	// Запускаем API сервер
	startAPIServer(bc, walletManager, mempool, *port, metricsCollector, perfLogger)

	// Ожидаем завершения
	slog.Info("Node is running. Press Ctrl+C to stop.")
	select {}
}

// startAPIServer запускает API сервер
func startAPIServer(bc interface{}, wm *wallet.WalletManager, mempool interface{}, port int, metricsCollector *metrics.MetricsCollector, perfLogger *logging.PerformanceLogger) {
	slog.Info("Starting API server", "port", port)

	// Создаем и запускаем API сервер
	// Пока что создаем обычный блокчейн для API сервера
	// TODO: Обновить API сервер для поддержки интерфейсов
	bcForAPI := &blockchain.Blockchain{}
	apiServer := api.NewServerWithMempool(bcForAPI, wm, mempool, port)

	// Запускаем API сервер в отдельной горутине
	go func() {
		err := apiServer.Start()
		if err != nil {
			slog.Error("Failed to start API server", "error", err)
		}
	}()

	// Выводим статистику блокчейна
	if blockchain, ok := bc.(interface{ GetStats() map[string]interface{} }); ok {
		stats := blockchain.GetStats()
		slog.Info("Blockchain stats", "stats", stats)

		// Логируем производительность
		if perfLogger != nil {
			height := int64(0)
			txCount := 0
			if h, ok := stats["height"].(int); ok {
				height = int64(h)
			}
			if tc, ok := stats["tx_count"].(int); ok {
				txCount = tc
			}
			perfLogger.LogBlockchainPerformance("blockchain_stats", time.Since(time.Now()), height, txCount)
		}

		// Обновляем метрики
		if metricsCollector != nil {
			if height, ok := stats["height"].(int); ok {
				gauge := metricsCollector.CreateGauge("blockchain_height", nil)
				gauge.Set(int64(height))
			}
			if utxoCount, ok := stats["utxo_count"].(int); ok {
				gauge := metricsCollector.CreateGauge("blockchain_utxo_count", nil)
				gauge.Set(int64(utxoCount))
			}
		}
	}
}
