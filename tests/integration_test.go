package tests

import (
	"testing"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/network"
	"mirochain/internal/wallet"
)

func TestFullBlockchainIntegration(t *testing.T) {
	// Создаем кошельки
	wm := wallet.NewWalletManager()
	wallet1, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet1: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)

	// Создаем mempool (пока что пустой, только coinbase транзакции)
	mempool := mining.NewMempool(100)

	// Создаем майнера (без запуска реального майнинга)
	miner := mining.NewMiner("miner_001", wallet1.GetPublicKeyBytes(), bc, mempool, nil, wallet1)

	// Проверяем, что майнер создан
	if miner == nil {
		t.Fatalf("Failed to create miner")
	}

	// Проверяем статистику майнера
	minerStats := miner.GetStats()
	if minerStats == nil {
		t.Fatalf("Miner should have stats")
	}

	// Проверяем, что блокчейн имеет genesis блок
	stats := bc.GetStats()
	if stats["height"].(int64) < 0 {
		t.Fatalf("Expected at least 0 blocks (genesis), got %d", stats["height"])
	}

	t.Logf("Full blockchain integration test completed. Height: %d, Miner ID: %s", stats["height"], miner.ID)
}

func TestP2PNetworkIntegration(t *testing.T) {
	// Создаем кошельки
	wm1 := wallet.NewWalletManager()
	wallet1, _ := wm1.CreateWallet()

	wm2 := wallet.NewWalletManager()
	wallet2, _ := wm2.CreateWallet()

	// Создаем блокчейны
	bc1 := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)
	bc2 := blockchain.NewBlockchain(wallet2.GetAddress(), wallet2.GetPublicKeyBytes(), 1)

	// Создаем серверы
	server1 := network.NewServer("127.0.0.1", 0, bc1) // Используем порт 0 для автоматического выбора
	server2 := network.NewServer("127.0.0.1", 0, bc2) // Используем порт 0 для автоматического выбора

	// Проверяем, что серверы созданы
	if server1 == nil {
		t.Error("Server1 should be created")
	}

	if server2 == nil {
		t.Error("Server2 should be created")
	}

	// Проверяем, что серверы не запущены изначально
	if server1.IsRunning {
		t.Error("Server1 should not be running initially")
	}

	if server2.IsRunning {
		t.Error("Server2 should not be running initially")
	}

	// Создаем клиент для подключения
	_ = network.NewClient(server1)

	t.Logf("P2P network integration test completed. Server1 created: %v, Server2 created: %v", server1 != nil, server2 != nil)
}

func TestMiningAndNetworkIntegration(t *testing.T) {
	// Создаем кошелек
	wm := wallet.NewWalletManager()
	wallet1, _ := wm.CreateWallet()

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 0, bc) // Используем порт 0 для автоматического выбора

	// Создаем mempool и майнера
	mempool := mining.NewMempool(100)
	miner := mining.NewMiner("miner_001", wallet1.GetPublicKeyBytes(), bc, mempool, server, wallet1)

	// Проверяем, что майнер создан
	if miner == nil {
		t.Fatalf("Failed to create miner")
	}

	// Проверяем, что сервер создан
	if server == nil {
		t.Fatalf("Failed to create server")
	}

	// Проверяем, что сервер не запущен изначально
	if server.IsRunning {
		t.Error("Server should not be running initially")
	}

	// Проверяем результат (только genesis блок)
	stats := bc.GetStats()
	if stats["height"].(int64) < 0 {
		t.Fatalf("Expected at least 0 blocks (genesis), got %d", stats["height"])
	}

	t.Logf("Mining and network integration test completed. Height: %d, Miner ID: %s", stats["height"], miner.ID)
}

func TestWalletAndBlockchainIntegration(t *testing.T) {
	// Создаем кошельки
	wm := wallet.NewWalletManager()
	wallet1, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet1: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)

	// Создаем второй кошелек для транзакции
	wallet2, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet2: %v", err)
	}

	// Создаем транзакцию
	tx, err := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 50, wallet1.GetPrivateKeyBytes())
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Проверяем, что транзакция валидна
	if !tx.IsValid() {
		t.Fatal("Transaction should be valid")
	}

	// Создаем блок с транзакцией
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 1)

	// Добавляем блок в блокчейн
	err = bc.AddBlock(block)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	// Проверяем, что блок добавлен
	stats := bc.GetStats()
	if stats["height"].(int64) != 1 {
		t.Fatalf("Expected height 1, got %d", stats["height"])
	}

	t.Logf("Wallet and blockchain integration test completed. Height: %d", stats["height"])
}
