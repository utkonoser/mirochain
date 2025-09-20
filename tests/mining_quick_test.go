package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/wallet"
)

func TestMempoolQuick(t *testing.T) {
	// Создаем mempool
	mempool := mining.NewMempool(10)

	// Проверяем начальное состояние
	if !mempool.IsEmpty() {
		t.Error("Mempool should be empty initially")
	}

	if mempool.Size() != 0 {
		t.Errorf("Expected size 0, got %d", mempool.Size())
	}

	// Проверяем статистику
	stats := mempool.GetStats()
	if stats["size"].(int) != 0 {
		t.Errorf("Expected stats size 0, got %d", stats["size"])
	}

	t.Logf("Mempool quick test completed. Stats: %+v", stats)
}

func TestMinerCreation(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн с минимальной сложностью (0)
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Создаем майнера
	miner := mining.NewMiner(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), bc, mempool, nil, nodeWallet)

	// Проверяем создание майнера
	if miner.ID == "" {
		t.Error("Miner should have ID")
	}

	if miner.Address != nodeWallet.GetAddress() {
		t.Error("Miner address should match wallet address")
	}

	if miner.Blockchain == nil {
		t.Error("Miner should have blockchain reference")
	}

	if miner.Mempool == nil {
		t.Error("Miner should have mempool reference")
	}

	// Проверяем, что майнер не работает изначально
	if miner.IsRunning() {
		t.Error("Miner should not be running initially")
	}

	// Проверяем статистику
	stats := miner.GetStats()
	if stats.BlocksMined != 0 {
		t.Errorf("Expected 0 blocks mined, got %d", stats.BlocksMined)
	}

	t.Logf("Miner creation test completed. ID: %s, Address: %s", miner.ID, miner.Address)
}

func TestMiningManagerQuick(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн с минимальной сложностью (0)
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Создаем менеджер майнинга
	manager := mining.NewManager(bc, mempool, nil)

	// Проверяем начальное состояние
	if manager.IsRunning {
		t.Error("Manager should not be running initially")
	}

	// Запускаем менеджер
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start mining manager: %v", err)
	}

	// Проверяем, что менеджер запущен
	if !manager.IsRunning {
		t.Error("Manager should be running")
	}

	// Добавляем майнера
	miner, err := manager.AddMiner(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), nodeWallet)
	if err != nil {
		t.Fatalf("Failed to add miner: %v", err)
	}

	// Проверяем, что майнер добавлен
	if miner == nil {
		t.Error("Miner should be created")
	}

	// Запускаем майнинг
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	defer miner.StopMining()

	// Ждем немного для майнинга
	time.Sleep(50 * time.Millisecond)

	// Проверяем статистику
	stats := manager.GetMiningStats()
	if stats["total_miners"].(int) != 1 {
		t.Errorf("Expected 1 miner, got %d", stats["total_miners"])
	}

	if stats["active_miners"].(int) != 1 {
		t.Errorf("Expected 1 active miner, got %d", stats["active_miners"])
	}

	// Проверяем, что блоки были замайнены
	blockchainStats := bc.GetStats()
	if blockchainStats["height"].(int64) < 1 {
		t.Errorf("Expected at least 1 block (genesis + mined), got %d", blockchainStats["height"])
	}

	// Останавливаем менеджер
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop mining manager: %v", err)
	}

	// Проверяем, что менеджер остановлен
	if manager.IsRunning {
		t.Error("Manager should be stopped")
	}

	t.Logf("Mining manager quick test completed. Stats: %+v", stats)
}
