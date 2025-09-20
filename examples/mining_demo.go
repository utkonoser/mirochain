//go:build mining_demo
// +build mining_demo

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
	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== MiroChain Mining Demo ===")

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet", "error", err)
		return
	}

	fmt.Printf("Miner wallet: %s\n", nodeWallet.GetAddress())

	// Создаем блокчейн с низкой сложностью для демонстрации
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 1)
	fmt.Printf("Blockchain created with height: %d\n", bc.GetHeight())

	// Создаем mempool
	mempool := mining.NewMempool(1000)
	fmt.Printf("Mempool created with max size: %d\n", mempool.MaxSize)

	// Создаем майнера
	miner := mining.NewMiner(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), bc, mempool, nil, nodeWallet)
	fmt.Printf("Miner created: %s\n", miner.ID)

	// Пока что не создаем транзакции, только майним coinbase блоки
	fmt.Println("Mining coinbase blocks only...")

	// Запускаем майнинг
	err = miner.StartMining()
	if err != nil {
		slog.Error("Failed to start mining", "error", err)
		return
	}
	defer miner.StopMining()

	fmt.Println("Mining started...")

	// Майним в течение 10 секунд
	fmt.Println("Mining for 10 seconds...")
	startTime := time.Now()

	for time.Since(startTime) < 10*time.Second {
		time.Sleep(2 * time.Second)

		// Выводим статистику
		stats := miner.GetStats()
		blockchainHeight := bc.GetHeight()
		mempoolSize := mempool.Size()

		fmt.Printf("Time: %v, Height: %d, Mempool: %d, Blocks Mined: %d, Hash Rate: %.2f H/s\n",
			time.Since(startTime).Round(time.Second),
			blockchainHeight,
			mempoolSize,
			stats.BlocksMined,
			stats.HashRate)
	}

	// Останавливаем майнинг
	miner.StopMining()
	fmt.Println("Mining stopped")

	// Выводим финальную статистику
	finalStats := miner.GetStats()
	fmt.Printf("\nFinal Statistics:\n")
	fmt.Printf("  Blocks Mined: %d\n", finalStats.BlocksMined)
	fmt.Printf("  Total Hashes: %d\n", finalStats.TotalHashes)
	fmt.Printf("  Average Block Time: %v\n", finalStats.AverageBlockTime)
	fmt.Printf("  Hash Rate: %.2f H/s\n", finalStats.HashRate)
	fmt.Printf("  Blockchain Height: %d\n", bc.GetHeight())
	fmt.Printf("  Mempool Size: %d\n", mempool.Size())

	fmt.Println("=== Demo completed ===")
}
