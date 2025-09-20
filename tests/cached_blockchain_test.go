package tests

import (
	"testing"

	"mirochain/internal/blockchain"
	"mirochain/internal/persistent"
	"mirochain/internal/wallet"
)

// TestCachedBlockchainCreation тестирует создание кэшированного блокчейна
func TestCachedBlockchainCreation(t *testing.T) {
	dataDir := t.TempDir()

	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	cpbc, err := persistent.NewCachedPersistentBlockchain(
		dataDir,
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create cached persistent blockchain: %v", err)
	}
	defer cpbc.Close()

	// Проверяем, что блокчейн создан
	height, err := cpbc.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}

	if height < 0 {
		t.Errorf("Expected height >= 0, got %d", height)
	}

	// Проверяем genesis блок
	genesisHash := cpbc.GetGenesisHash()
	if genesisHash == nil {
		t.Error("Genesis hash should not be nil")
	}

	t.Logf("Cached persistent blockchain created successfully, height: %d", height)
}

// TestCachedBlockchainPersistence тестирует персистентность кэшированного блокчейна
func TestCachedBlockchainPersistence(t *testing.T) {
	dataDir := t.TempDir()

	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем первый блокчейн
	cpbc1, err := persistent.NewCachedPersistentBlockchain(
		dataDir,
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create first blockchain: %v", err)
	}

	// Добавляем несколько блоков
	var previousHash []byte = cpbc1.GetGenesisHash()
	for i := 0; i < 3; i++ {
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
		block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, previousHash, int64(i+1), 0)

		err = cpbc1.AddBlock(block)
		if err != nil {
			t.Fatalf("Failed to add block %d: %v", i+1, err)
		}

		// Обновляем previous hash для следующего блока
		previousHash = block.Hash
	}

	// Получаем статистику первого блокчейна
	stats1 := cpbc1.GetStats()
	height1, _ := cpbc1.GetHeight()
	balance1 := cpbc1.GetBalance(nodeWallet.GetAddress())

	cpbc1.Close()

	// Создаем второй блокчейн с теми же данными
	cpbc2, err := persistent.NewCachedPersistentBlockchain(
		dataDir,
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create second blockchain: %v", err)
	}
	defer cpbc2.Close()

	// Проверяем, что данные сохранились
	height2, err := cpbc2.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height from second blockchain: %v", err)
	}

	if height2 != height1 {
		t.Errorf("Expected height %d, got %d", height1, height2)
	}

	balance2 := cpbc2.GetBalance(nodeWallet.GetAddress())
	if balance2 != balance1 {
		t.Errorf("Expected balance %d, got %d", balance1, balance2)
	}

	// Проверяем статистику
	stats2 := cpbc2.GetStats()
	if stats2["height"] != stats1["height"] {
		t.Errorf("Expected height %v, got %v", stats1["height"], stats2["height"])
	}

	t.Logf("Cached blockchain persistence test completed. Height: %d, Balance: %d", height2, balance2)
}

// TestCachedBlockchainCaching тестирует работу кэширования
func TestCachedBlockchainCaching(t *testing.T) {
	dataDir := t.TempDir()

	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	cpbc, err := persistent.NewCachedPersistentBlockchain(
		dataDir,
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create cached blockchain: %v", err)
	}
	defer cpbc.Close()

	// Получаем genesis блок несколько раз
	genesisHash := cpbc.GetGenesisHash()
	block1, err := cpbc.GetBlock(genesisHash)
	if err != nil {
		t.Fatalf("Failed to get genesis block: %v", err)
	}

	block2, err := cpbc.GetBlock(genesisHash)
	if err != nil {
		t.Fatalf("Failed to get genesis block second time: %v", err)
	}

	if block1.Height != block2.Height {
		t.Error("Blocks should be identical")
	}

	// Получаем блок по высоте
	blockByHeight, err := cpbc.GetBlockByHeight(0)
	if err != nil {
		t.Fatalf("Failed to get block by height: %v", err)
	}

	if blockByHeight.Height != block1.Height {
		t.Error("Blocks should be identical")
	}

	// Проверяем кэширование баланса
	balance1 := cpbc.GetBalance(nodeWallet.GetAddress())
	balance2 := cpbc.GetBalance(nodeWallet.GetAddress())

	if balance1 != balance2 {
		t.Error("Balances should be identical")
	}

	// Проверяем статистику кэша
	stats := cpbc.GetStats()
	cacheHits, ok := stats["cache_hits"].(int64)
	if !ok {
		t.Error("Cache hits should be available in stats")
	}

	if cacheHits < 1 {
		t.Error("Should have at least one cache hit")
	}

	t.Logf("Caching test completed. Cache hits: %d", cacheHits)
}

// TestCachedBlockchainStats тестирует статистику кэшированного блокчейна
func TestCachedBlockchainStats(t *testing.T) {
	dataDir := t.TempDir()

	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	cpbc, err := persistent.NewCachedPersistentBlockchain(
		dataDir,
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		5,
	)
	if err != nil {
		t.Fatalf("Failed to create cached blockchain: %v", err)
	}
	defer cpbc.Close()

	stats := cpbc.GetStats()

	// Проверяем наличие всех полей
	requiredFields := []string{"height", "difficulty", "utxo_count", "cache_hits", "cache_misses", "cache_size"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Stats should contain field: %s", field)
		}
	}

	// Проверяем значения
	height, ok := stats["height"].(int64)
	if !ok {
		t.Error("Height should be int64")
	}
	if height < 0 {
		t.Errorf("Expected height >= 0, got %d", height)
	}

	difficulty, ok := stats["difficulty"].(int)
	if !ok {
		t.Error("Difficulty should be int")
	}
	if difficulty != 5 && difficulty != 0 {
		t.Errorf("Expected difficulty 5 or 0, got %d", difficulty)
	}

	utxoCount, ok := stats["utxo_count"].(int)
	if !ok {
		t.Error("UTXO count should be int")
	}
	if utxoCount < 0 {
		t.Errorf("Expected UTXO count >= 0, got %d", utxoCount)
	}

	cacheHits, ok := stats["cache_hits"].(int64)
	if !ok {
		t.Error("Cache hits should be int64")
	}
	if cacheHits < 0 {
		t.Errorf("Expected cache hits >= 0, got %d", cacheHits)
	}

	t.Logf("Stats test completed. Height: %d, Difficulty: %d, UTXO count: %d, Cache hits: %d",
		height, difficulty, utxoCount, cacheHits)
}

// TestCachedBlockchainInvalidBlock тестирует обработку невалидных блоков
func TestCachedBlockchainInvalidBlock(t *testing.T) {
	dataDir := t.TempDir()

	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	cpbc, err := persistent.NewCachedPersistentBlockchain(
		dataDir,
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		0,
	)
	if err != nil {
		t.Fatalf("Failed to create cached blockchain: %v", err)
	}
	defer cpbc.Close()

	// Создаем невалидный блок (с неправильной высотой)
	coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)
	invalidBlock := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, cpbc.GetGenesisHash(), 999, 0) // Неправильная высота

	err = cpbc.AddBlock(invalidBlock)
	if err == nil {
		t.Error("Should fail to add invalid block")
	}

	t.Logf("Invalid block test completed successfully")
}

// TestCachedBlockchainIntegration тестирует интеграцию кэшированного блокчейна
func TestCachedBlockchainIntegration(t *testing.T) {
	t.Run("Creation", TestCachedBlockchainCreation)
	t.Run("Persistence", TestCachedBlockchainPersistence)
	t.Run("Caching", TestCachedBlockchainCaching)
	t.Run("Stats", TestCachedBlockchainStats)
	t.Run("InvalidBlock", TestCachedBlockchainInvalidBlock)
}
