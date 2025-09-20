//go:build basic_usage
// +build basic_usage

package main

import (
	"fmt"
	"log/slog"
	"os"

	"mirochain/internal/blockchain"
	"mirochain/internal/wallet"
)

func main() {
	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== MiroChain Basic Usage Example ===")

	// Создаем кошельки
	walletManager := wallet.NewWalletManager()

	// Создаем первый кошелек (для genesis блока)
	wallet1, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet1", "error", err)
		return
	}

	// Создаем второй кошелек
	wallet2, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet2", "error", err)
		return
	}

	fmt.Printf("Wallet 1: %s\n", wallet1.GetAddress())
	fmt.Printf("Wallet 2: %s\n", wallet2.GetAddress())

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 4)

	fmt.Printf("Blockchain created with height: %d\n", bc.GetHeight())
	fmt.Printf("Wallet 1 balance: %d\n", bc.GetBalance(wallet1.GetAddress()))
	fmt.Printf("Wallet 2 balance: %d\n", bc.GetBalance(wallet2.GetAddress()))

	// Создаем транзакцию
	tx, err := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 100, wallet1.GetPrivateKeyBytes())
	if err != nil {
		slog.Error("Failed to create transaction", "error", err)
		return
	}

	fmt.Printf("Transaction created: %x\n", tx.ID)
	fmt.Printf("Transaction valid: %t\n", tx.IsValid())

	// Получаем статистику блокчейна
	stats := bc.GetStats()
	fmt.Printf("Blockchain stats: %+v\n", stats)

	fmt.Println("=== Example completed ===")
}
