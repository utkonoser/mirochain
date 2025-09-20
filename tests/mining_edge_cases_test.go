package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/wallet"
)

// TestMiningWithEmptyMempool тестирует майнинг с пустым mempool
func TestMiningWithEmptyMempool(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	mempool := mining.NewMempool(100)

	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Запускаем майнинг с пустым mempool
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Проверяем, что блокчейн имеет блоки (genesis + mined)
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", stats["height"])
	}

	t.Logf("Mining with empty mempool completed. Height: %d", stats["height"])
}

// TestMiningWithFullMempool тестирует майнинг с полным mempool
func TestMiningWithFullMempool(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	mempool := mining.NewMempool(10) // Маленький mempool

	// Заполняем mempool
	for i := 0; i < 15; i++ { // Больше чем размер mempool
		tx := &blockchain.Transaction{
			Inputs: []*blockchain.TransactionInput{},
			Outputs: []*blockchain.TransactionOutput{
				{
					Value:     100,
					Address:   "recipient_address",
					PublicKey: []byte("recipient_public_key"),
				},
			},
			Timestamp: time.Now().Unix(),
			Fee:       1,
		}
		mempool.AddTransaction(tx)
	}

	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Запускаем майнинг с полным mempool
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Проверяем, что блокчейн имеет блоки
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", stats["height"])
	}

	t.Logf("Mining with full mempool completed. Height: %d", stats["height"])
}

// TestMiningWithInvalidTransactions тестирует майнинг с невалидными транзакциями
func TestMiningWithInvalidTransactions(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	mempool := mining.NewMempool(100)

	// Добавляем невалидную транзакцию
	invalidTx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     -100, // Отрицательная сумма
				Address:   "recipient_address",
				PublicKey: []byte("recipient_public_key"),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}
	mempool.AddTransaction(invalidTx)

	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Запускаем майнинг с невалидными транзакциями
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Проверяем, что блокчейн имеет блоки (genesis + mined)
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", stats["height"])
	}

	t.Logf("Mining with invalid transactions completed. Height: %d", stats["height"])
}

// TestMiningWithHighDifficulty тестирует майнинг с высокой сложностью
func TestMiningWithHighDifficulty(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 3) // Высокая сложность
	mempool := mining.NewMempool(100)

	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Запускаем майнинг с высокой сложностью
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	time.Sleep(200 * time.Millisecond) // Больше времени для высокой сложности
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Проверяем, что блокчейн имеет блоки
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", stats["height"])
	}

	t.Logf("Mining with high difficulty completed. Height: %d", stats["height"])
}

// TestMiningWithZeroDifficulty тестирует майнинг с нулевой сложностью
func TestMiningWithZeroDifficulty(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0) // Нулевая сложность
	mempool := mining.NewMempool(100)

	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Запускаем майнинг с нулевой сложностью
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}
	time.Sleep(50 * time.Millisecond) // Уменьшаем время ожидания
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Проверяем, что блокчейн имеет блоки
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", stats["height"])
	}

	// Проверяем, что все блоки имеют nonce=0 (для сложности 0)
	for _, block := range bc.Chain[1:] { // Пропускаем genesis блок
		if block.Nonce != 0 {
			t.Errorf("Expected nonce 0 for block at height %d, got %d", block.Height, block.Nonce)
		}
	}

	t.Logf("Mining with zero difficulty completed. Height: %d", stats["height"])
}

// TestMiningWithConcurrentMiners тестирует майнинг с несколькими майнерами
func TestMiningWithConcurrentMiners(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	mempool := mining.NewMempool(100)

	// Создаем несколько майнеров
	miner1 := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	miner2 := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Запускаем майнинг на обоих майнерах
	err = miner1.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining on miner1: %v", err)
	}
	err = miner2.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining on miner2: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Останавливаем майнинг
	err = miner1.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining on miner1: %v", err)
	}
	err = miner2.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining on miner2: %v", err)
	}

	// Проверяем, что блокчейн имеет блоки
	stats := bc.GetStats()
	if stats["height"].(int64) < 1 {
		t.Fatalf("Expected at least 1 block, got %d", stats["height"])
	}

	t.Logf("Mining with concurrent miners completed. Height: %d", stats["height"])
}

// TestMiningEdgeCases тестирует граничные случаи майнинга
func TestMiningEdgeCases(t *testing.T) {
	t.Run("EmptyMempool", TestMiningWithEmptyMempool)
	t.Run("FullMempool", TestMiningWithFullMempool)
	t.Run("InvalidTransactions", TestMiningWithInvalidTransactions)
	t.Run("HighDifficulty", TestMiningWithHighDifficulty)
	t.Run("ZeroDifficulty", TestMiningWithZeroDifficulty)
	t.Run("ConcurrentMiners", TestMiningWithConcurrentMiners)
}
