//go:build profiling_usage
// +build profiling_usage

package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/logging"
	"mirochain/internal/profiling"
	"mirochain/internal/wallet"
)

func main() {
	// Создаем профилировщик
	profiler := profiling.NewProfiler(profiling.ProfilerConfig{
		Enabled:      true,
		HTTPPort:     ":6060",
		ProfileDir:   "./profiles",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	})

	// Запускаем профилировщик
	if err := profiler.Start(); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop()

	// Создаем логгер производительности
	perfLogger := logging.NewPerformanceLogger(logging.PerformanceConfig{
		Level:      slog.LevelInfo,
		Output:     "stdout",
		Format:     "text",
		IncludeMem: true,
		IncludeCPU: true,
	})

	// Создаем кошельки
	wallet1, err := wallet.NewWallet()
	if err != nil {
		log.Fatalf("Failed to create wallet1: %v", err)
	}

	wallet2, err := wallet.NewWallet()
	if err != nil {
		log.Fatalf("Failed to create wallet2: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1) // Низкая сложность для демонстрации

	// Создаем майнер (упрощенная версия для примера)
	_ = struct{}{} // Создаем пустой майнер для демонстрации

	// Логируем начальное состояние
	perfLogger.LogMemoryUsage("initial")
	perfLogger.LogGoroutineCount("initial")

	// Создаем транзакции для профилирования
	fmt.Println("Creating transactions...")
	for i := 0; i < 100; i++ {
		// Создаем транзакцию
		tx, err := wallet1.CreateTransaction(wallet2.GetAddress(), 10, bc)
		if err != nil {
			log.Printf("Failed to create transaction %d: %v", i, err)
			continue
		}

		// Добавляем транзакцию в блокчейн (упрощенная версия)
		_ = tx // Игнорируем транзакцию для демонстрации

		// Логируем производительность каждые 10 транзакций
		if i%10 == 0 {
			perfLogger.LogMemoryUsage(fmt.Sprintf("after_%d_transactions", i))
			perfLogger.LogGoroutineCount(fmt.Sprintf("after_%d_transactions", i))
		}
	}

	// Ждем некоторое время для майнинга
	fmt.Println("Waiting for mining...")
	time.Sleep(2 * time.Second)

	// Логируем производительность после майнинга
	perfLogger.LogMemoryUsage("after_mining")
	perfLogger.LogGoroutineCount("after_mining")

	// Записываем профили
	fmt.Println("Writing profiles...")

	// Memory профиль
	if err := profiler.WriteMemProfile("example_mem.prof"); err != nil {
		log.Printf("Failed to write memory profile: %v", err)
	}

	// Block профиль
	if err := profiler.WriteBlockProfile("example_block.prof"); err != nil {
		log.Printf("Failed to write block profile: %v", err)
	}

	// Mutex профиль
	if err := profiler.WriteMutexProfile("example_mutex.prof"); err != nil {
		log.Printf("Failed to write mutex profile: %v", err)
	}

	// Goroutine профиль
	if err := profiler.WriteGoroutineProfile("example_goroutine.prof"); err != nil {
		log.Printf("Failed to write goroutine profile: %v", err)
	}

	// Записываем все профили сразу
	if err := profiler.WriteAllProfiles("example_all"); err != nil {
		log.Printf("Failed to write all profiles: %v", err)
	}

	// Получаем статистику рантайма
	stats := profiler.GetRuntimeStats()
	fmt.Printf("Runtime stats: %+v\n", stats)

	// Получаем статистику блокчейна
	blockchainStats := bc.GetStats()
	height := int64(0)
	txCount := 0
	if h, ok := blockchainStats["height"]; ok {
		if hInt, ok := h.(int); ok {
			height = int64(hInt)
		} else if hInt64, ok := h.(int64); ok {
			height = hInt64
		}
	}
	if tc, ok := blockchainStats["tx_count"]; ok {
		if tcInt, ok := tc.(int); ok {
			txCount = tcInt
		} else if tcInt64, ok := tc.(int64); ok {
			txCount = int(tcInt64)
		}
	}
	perfLogger.LogBlockchainPerformance("final", 100*time.Millisecond, height, txCount)

	// Получаем статистику майнера (упрощенная версия)
	perfLogger.LogMiningPerformance(100*time.Millisecond, 0, 0, "0 H/s")

	fmt.Println("Profiling example completed!")
	fmt.Println("Check the ./profiles directory for generated profile files")
	fmt.Println("You can analyze them using:")
	fmt.Println("  go tool pprof example_cpu.prof")
	fmt.Println("  go tool pprof example_mem.prof")
	fmt.Println("  go tool pprof example_block.prof")
	fmt.Println("  go tool pprof example_mutex.prof")
	fmt.Println("  go tool pprof example_goroutine.prof")
}
