package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/wallet"
)

// TestOptimizedProofOfWorkCreation тестирует создание оптимизированного PoW
func TestOptimizedProofOfWorkCreation(t *testing.T) {
	// Создаем тестовый блок
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
	block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, 0, 1)

	// Создаем PoW
	pow := mining.NewOptimizedProofOfWork(block, 1)

	if pow == nil {
		t.Fatal("PoW should not be nil")
	}

	if pow.Difficulty != 1 {
		t.Errorf("Expected difficulty 1, got %d", pow.Difficulty)
	}

	if pow.Nonce != 0 {
		t.Errorf("Expected nonce 0, got %d", pow.Nonce)
	}

	if pow.Valid {
		t.Error("PoW should not be valid initially")
	}

	t.Logf("Optimized PoW created successfully")
}

// TestOptimizedProofOfWorkMining тестирует майнинг с оптимизированным PoW
func TestOptimizedProofOfWorkMining(t *testing.T) {
	// Создаем тестовый блок
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
	block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, 0, 1)

	// Создаем PoW с очень низкой сложностью (1 бит)
	pow := mining.NewOptimizedProofOfWork(block, 1)

	// Выполняем майнинг
	start := time.Now()
	nonce, hash, success := pow.Mine()
	elapsed := time.Since(start)

	if !success {
		t.Fatal("Mining should succeed with difficulty 1")
	}

	if nonce < 0 {
		t.Errorf("Expected nonce >= 0, got %d", nonce)
	}

	if hash == nil {
		t.Fatal("Hash should not be nil")
	}

	if !pow.Valid {
		t.Error("PoW should be valid after successful mining")
	}

	t.Logf("Mining completed successfully. Nonce: %d, Time: %v", nonce, elapsed)
}

// TestOptimizedProofOfWorkParallelMining тестирует параллельный майнинг
func TestOptimizedProofOfWorkParallelMining(t *testing.T) {
	// Создаем тестовый блок
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
	block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, 0, 1)

	// Создаем PoW с очень низкой сложностью (1 бит)
	pow := mining.NewOptimizedProofOfWork(block, 1)

	// Выполняем параллельный майнинг
	start := time.Now()
	nonce, hash, success := pow.MineParallel(2)
	elapsed := time.Since(start)

	if !success {
		t.Fatal("Parallel mining should succeed with difficulty 1")
	}

	if nonce < 0 {
		t.Errorf("Expected nonce >= 0, got %d", nonce)
	}

	if hash == nil {
		t.Fatal("Hash should not be nil")
	}

	if !pow.Valid {
		t.Error("PoW should be valid after successful parallel mining")
	}

	t.Logf("Parallel mining completed successfully. Nonce: %d, Time: %v", nonce, elapsed)
}

// TestOptimizedMinerCreation тестирует создание оптимизированного майнера
func TestOptimizedMinerCreation(t *testing.T) {
	miner := mining.NewOptimizedMiner(4)

	if miner == nil {
		t.Fatal("Miner should not be nil")
	}

	if miner.IsRunning() {
		t.Error("Miner should not be running initially")
	}

	stats := miner.GetStats()
	if stats.BlocksMined != 0 {
		t.Errorf("Expected blocks mined 0, got %d", stats.BlocksMined)
	}

	t.Logf("Optimized miner created successfully")
}

// TestOptimizedMinerStartStop тестирует запуск и остановку майнера
func TestOptimizedMinerStartStop(t *testing.T) {
	miner := mining.NewOptimizedMiner(2)

	// Запускаем майнер
	err := miner.Start()
	if err != nil {
		t.Fatalf("Failed to start miner: %v", err)
	}

	if !miner.IsRunning() {
		t.Error("Miner should be running after start")
	}

	// Останавливаем майнер
	err = miner.Stop()
	if err != nil {
		t.Fatalf("Failed to stop miner: %v", err)
	}

	if miner.IsRunning() {
		t.Error("Miner should not be running after stop")
	}

	t.Logf("Miner start/stop test completed successfully")
}

// TestOptimizedMinerMining тестирует майнинг блока
func TestOptimizedMinerMining(t *testing.T) {
	miner := mining.NewOptimizedMiner(2)

	// Запускаем майнер
	err := miner.Start()
	if err != nil {
		t.Fatalf("Failed to start miner: %v", err)
	}
	defer miner.Stop()

	// Создаем тестовый блок
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
	block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, 0, 1)

	// Майним блок
	start := time.Now()
	nonce, hash, success := miner.MineBlock(block)
	elapsed := time.Since(start)

	if !success {
		t.Fatal("Mining should succeed")
	}

	if nonce < 0 {
		t.Errorf("Expected nonce >= 0, got %d", nonce)
	}

	if hash == nil {
		t.Fatal("Hash should not be nil")
	}

	// Проверяем статистику
	stats := miner.GetStats()
	if stats.BlocksMined != 1 {
		t.Errorf("Expected blocks mined 1, got %d", stats.BlocksMined)
	}

	if stats.TotalTime <= 0 {
		t.Error("Total time should be positive")
	}

	if stats.AverageTime <= 0 {
		t.Error("Average time should be positive")
	}

	t.Logf("Block mining completed successfully. Nonce: %d, Time: %v, Stats: %+v", nonce, elapsed, stats)
}

// TestOptimizedMinerStats тестирует статистику майнера
func TestOptimizedMinerStats(t *testing.T) {
	miner := mining.NewOptimizedMiner(2)

	// Запускаем майнер
	err := miner.Start()
	if err != nil {
		t.Fatalf("Failed to start miner: %v", err)
	}
	defer miner.Stop()

	// Создаем и майним несколько блоков
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	for i := 0; i < 3; i++ {
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
		block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, int64(i), 1)

		_, _, success := miner.MineBlock(block)
		if !success {
			t.Fatalf("Failed to mine block %d", i)
		}
	}

	// Проверяем статистику
	stats := miner.GetStats()

	if stats.BlocksMined != 3 {
		t.Errorf("Expected blocks mined 3, got %d", stats.BlocksMined)
	}

	if stats.TotalTime <= 0 {
		t.Error("Total time should be positive")
	}

	if stats.AverageTime <= 0 {
		t.Error("Average time should be positive")
	}

	if stats.LastMined.IsZero() {
		t.Error("Last mined time should not be zero")
	}

	t.Logf("Miner stats test completed successfully. Stats: %+v", stats)
}

// TestOptimizedMinerIntegration тестирует интеграцию оптимизированного майнера
func TestOptimizedMinerIntegration(t *testing.T) {
	t.Run("Creation", TestOptimizedMinerCreation)
	t.Run("StartStop", TestOptimizedMinerStartStop)
	t.Run("Mining", TestOptimizedMinerMining)
	t.Run("Stats", TestOptimizedMinerStats)
}
