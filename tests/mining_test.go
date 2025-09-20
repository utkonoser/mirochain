package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/wallet"
)

func TestMempool(t *testing.T) {
	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Проверяем, что mempool пуст
	if !mempool.IsEmpty() {
		t.Error("Mempool should be empty")
	}

	// Создаем кошельки
	walletManager := wallet.NewWalletManager()
	wallet1, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet1: %v", err)
	}

	wallet2, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet2: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 4)

	// Создаем транзакцию
	tx, err := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 100, wallet1.GetPrivateKeyBytes())
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Добавляем транзакцию в mempool
	err = mempool.AddTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to add transaction to mempool: %v", err)
	}

	// Проверяем, что транзакция добавлена
	if mempool.Size() != 1 {
		t.Errorf("Expected mempool size 1, got %d", mempool.Size())
	}

	// Проверяем получение транзакции
	retrievedTx, exists := mempool.GetTransaction(string(tx.ID))
	if !exists {
		t.Error("Transaction should exist in mempool")
	}

	if retrievedTx.ID == nil {
		t.Error("Retrieved transaction should have ID")
	}

	t.Logf("Mempool test completed. Size: %d", mempool.Size())
}

// TestMiner тестирует базовую функциональность майнера
func TestMiner(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн с СЛОЖНОСТЬЮ 0 (мгновенный майнинг)
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Создаем майнер
	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil, // без сети
		nodeWallet,
	)

	// Проверяем, что майнер создан
	if miner == nil {
		t.Fatal("Miner should be created")
	}

	// Проверяем начальную статистику
	stats := miner.GetStats()
	if stats == nil {
		t.Fatal("Miner should have stats")
	}

	// Запускаем майнинг
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}

	// Ждем немного для майнинга
	time.Sleep(50 * time.Millisecond)

	// Останавливаем майнинг
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Проверяем, что блок был замайнен
	blockchainStats := bc.GetStats()
	if blockchainStats["height"].(int64) < 1 {
		t.Errorf("Expected at least 1 block (genesis + mined), got %d", blockchainStats["height"])
	}

	// Проверяем статистику майнера
	minerStats := miner.GetStats()
	if minerStats.BlocksMined == 0 {
		t.Error("Expected at least 1 mined block")
	}

	t.Logf("Miner test completed. Height: %d, Blocks Mined: %d",
		blockchainStats["height"], minerStats.BlocksMined)
}

func TestMiningManager(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 4)

	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Создаем менеджер майнинга
	manager := mining.NewManager(bc, mempool, nil)

	// Запускаем менеджер
	err = manager.Start()
	if err != nil {
		t.Fatalf("Failed to start mining manager: %v", err)
	}
	defer func() {
		if err := manager.Stop(); err != nil {
			t.Logf("Warning: failed to stop mining manager: %v", err)
		}
	}()

	// Добавляем майнера
	miner, err := manager.AddMiner(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), nodeWallet)
	if err != nil {
		t.Fatalf("Failed to add miner: %v", err)
	}

	// Проверяем, что майнер добавлен
	if miner == nil {
		t.Error("Miner should be created")
	}

	// Проверяем статистику
	stats := manager.GetMiningStats()
	if stats["total_miners"].(int) != 1 {
		t.Errorf("Expected 1 miner, got %d", stats["total_miners"])
	}

	t.Logf("Mining manager test completed. Stats: %+v", stats)
}

// TestMiningLowDifficulty тестирует майнинг на самой низкой сложности (0)
func TestMiningLowDifficulty(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн с СЛОЖНОСТЬЮ 0 (мгновенный майнинг)
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Создаем майнер
	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil, // без сети
		nodeWallet,
	)

	// Запускаем майнинг
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	defer miner.StopMining()

	// Ждем немного для майнинга
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что блок был замайнен
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Errorf("Expected at least 1 block (genesis + mined), got %d", stats["height"])
	}

	// Проверяем статистику майнера
	minerStats := miner.GetStats()
	if minerStats.BlocksMined == 0 {
		t.Error("Expected at least 1 mined block")
	}

	t.Logf("Mining test completed. Height: %d, Blocks Mined: %d",
		stats["height"], minerStats.BlocksMined)
}

// TestMiningDifficulty1 тестирует майнинг на сложности 1 (быстрый майнинг)
func TestMiningDifficulty1(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн с СЛОЖНОСТЬЮ 1 (быстрый майнинг)
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 1)

	// Создаем mempool
	mempool := mining.NewMempool(100)

	// Создаем майнер
	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil, // без сети
		nodeWallet,
	)

	// Запускаем майнинг
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	defer miner.StopMining()

	// Ждем немного для майнинга
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что блок был замайнен
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Errorf("Expected at least 1 block (genesis + mined), got %d", stats["height"])
	}

	// Проверяем статистику майнера
	minerStats := miner.GetStats()
	if minerStats.BlocksMined == 0 {
		t.Error("Expected at least 1 mined block")
	}

	t.Logf("Mining difficulty 1 test completed. Height: %d, Blocks Mined: %d",
		stats["height"], minerStats.BlocksMined)
}
