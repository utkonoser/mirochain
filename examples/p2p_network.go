//go:build p2p_network
// +build p2p_network

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"mirochain/internal/api"
	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/wallet"
)

func mainP2P() {
	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== MiroChain P2P Network Example ===")

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		slog.Error("Failed to create wallet", "error", err)
		return
	}

	fmt.Printf("Node wallet: %s\n", nodeWallet.GetAddress())

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 4)
	fmt.Printf("Blockchain created with height: %d\n", bc.GetHeight())

	// Создаем P2P сервер
	p2pServer := network.NewServer("127.0.0.1", 8080, bc)

	// Запускаем P2P сервер
	err = p2pServer.Start()
	if err != nil {
		slog.Error("Failed to start P2P server", "error", err)
		return
	}
	defer p2pServer.Stop()

	fmt.Printf("P2P server started on port 8080\n")

	// Создаем API сервер
	apiServer := api.NewServer(bc, walletManager, 8081)

	// Запускаем API сервер в отдельной горутине
	go func() {
		err := apiServer.Start()
		if err != nil {
			slog.Error("Failed to start API server", "error", err)
		}
	}()

	fmt.Printf("API server started on port 8081\n")

	// Создаем P2P клиент
	_ = network.NewClient(p2pServer)

	// Выводим информацию о сервере
	fmt.Printf("Node ID: %s\n", p2pServer.ID)
	fmt.Printf("Connected peers: %d\n", p2pServer.GetPeerCount())

	// Выводим статистику блокчейна
	stats := bc.GetStats()
	fmt.Printf("Blockchain stats: %+v\n", stats)

	// Выводим доступные API endpoints
	fmt.Println("\nAvailable API endpoints:")
	fmt.Println("  GET  /api/status          - Node status")
	fmt.Println("  GET  /api/blocks          - List blocks")
	fmt.Println("  GET  /api/blocks/{height} - Get block by height")
	fmt.Println("  GET  /api/wallets         - List wallets")
	fmt.Println("  GET  /api/balance/{addr}  - Get balance")
	fmt.Println("  GET  /api/utxos/{addr}    - Get UTXOs")

	fmt.Println("\nExample API calls:")
	fmt.Println("  curl http://localhost:8081/api/status")
	fmt.Println("  curl http://localhost:8081/api/blocks")
	fmt.Println("  curl http://localhost:8081/api/balance/" + nodeWallet.GetAddress())

	// Ждем некоторое время
	fmt.Println("\nServer is running for 10 seconds...")
	time.Sleep(10 * time.Second)

	fmt.Println("=== Example completed ===")
}
