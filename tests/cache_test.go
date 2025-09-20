package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/cache"
	"mirochain/internal/wallet"
)

// TestLRUCacheCreation тестирует создание LRU кэша
func TestLRUCacheCreation(t *testing.T) {
	c := cache.NewLRUCache(100, 50, 50, 5*time.Minute)

	if c == nil {
		t.Fatal("Cache should not be nil")
	}

	stats := c.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected initial size 0, got %d", stats.Size)
	}

	t.Logf("LRU cache created successfully")
}

// TestBlockCaching тестирует кэширование блоков
func TestBlockCaching(t *testing.T) {
	c := cache.NewLRUCache(10, 10, 10, 5*time.Minute)

	// Создаем тестовый блок
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	genesisBlock := bc.GetBlockByHeight(0)

	// Тестируем кэширование по хешу
	hash := genesisBlock.Hash
	block, found := c.GetBlock(hash)
	if found {
		t.Error("Block should not be found in empty cache")
	}

	c.SetBlock(hash, genesisBlock)
	block, found = c.GetBlock(hash)
	if !found {
		t.Error("Block should be found after caching")
	}

	if block.Height != genesisBlock.Height {
		t.Errorf("Expected height %d, got %d", genesisBlock.Height, block.Height)
	}

	// Тестируем кэширование по высоте
	height := genesisBlock.Height
	block, found = c.GetBlockByHeight(height)
	if found {
		t.Error("Block should not be found in empty cache by height")
	}

	c.SetBlockByHeight(height, genesisBlock)
	block, found = c.GetBlockByHeight(height)
	if !found {
		t.Error("Block should be found after caching by height")
	}

	if block.Height != genesisBlock.Height {
		t.Errorf("Expected height %d, got %d", genesisBlock.Height, block.Height)
	}

	t.Logf("Block caching test completed successfully")
}

// TestUTXOCaching тестирует кэширование UTXO
func TestUTXOCaching(t *testing.T) {
	c := cache.NewLRUCache(10, 10, 10, 5*time.Minute)

	// Создаем тестовые UTXO
	utxos := []*blockchain.UTXO{
		{
			TransactionID: []byte("tx1"),
			OutputIndex:   0,
			Value:         100,
			Address:       "address1",
			PublicKey:     []byte("pubkey1"),
		},
		{
			TransactionID: []byte("tx2"),
			OutputIndex:   1,
			Value:         200,
			Address:       "address1",
			PublicKey:     []byte("pubkey1"),
		},
	}

	address := "address1"

	// Тестируем кэширование UTXO
	cachedUTXOs, found := c.GetUTXOs(address)
	if found {
		t.Error("UTXOs should not be found in empty cache")
	}

	c.SetUTXOs(address, utxos)
	cachedUTXOs, found = c.GetUTXOs(address)
	if !found {
		t.Error("UTXOs should be found after caching")
	}

	if len(cachedUTXOs) != len(utxos) {
		t.Errorf("Expected %d UTXOs, got %d", len(utxos), len(cachedUTXOs))
	}

	if cachedUTXOs[0].Value != utxos[0].Value {
		t.Errorf("Expected value %d, got %d", utxos[0].Value, cachedUTXOs[0].Value)
	}

	t.Logf("UTXO caching test completed successfully")
}

// TestBalanceCaching тестирует кэширование балансов
func TestBalanceCaching(t *testing.T) {
	c := cache.NewLRUCache(10, 10, 10, 5*time.Minute)

	address := "test_address"
	balance := int64(1000)

	// Тестируем кэширование баланса
	cachedBalance, found := c.GetBalance(address)
	if found {
		t.Error("Balance should not be found in empty cache")
	}

	c.SetBalance(address, balance)
	cachedBalance, found = c.GetBalance(address)
	if !found {
		t.Error("Balance should be found after caching")
	}

	if cachedBalance != balance {
		t.Errorf("Expected balance %d, got %d", balance, cachedBalance)
	}

	t.Logf("Balance caching test completed successfully")
}

// TestMetadataCaching тестирует кэширование метаданных
func TestMetadataCaching(t *testing.T) {
	c := cache.NewLRUCache(10, 10, 10, 5*time.Minute)

	height := int64(100)
	difficulty := 5

	// Тестируем кэширование высоты
	cachedHeight, found := c.GetHeight()
	if !found {
		t.Error("Height should always be found (default value)")
	}

	c.SetHeight(height)
	cachedHeight, found = c.GetHeight()
	if !found {
		t.Error("Height should be found after caching")
	}

	if cachedHeight != height {
		t.Errorf("Expected height %d, got %d", height, cachedHeight)
	}

	// Тестируем кэширование сложности
	cachedDifficulty, found := c.GetDifficulty()
	if !found {
		t.Error("Difficulty should always be found (default value)")
	}

	c.SetDifficulty(difficulty)
	cachedDifficulty, found = c.GetDifficulty()
	if !found {
		t.Error("Difficulty should be found after caching")
	}

	if cachedDifficulty != difficulty {
		t.Errorf("Expected difficulty %d, got %d", difficulty, cachedDifficulty)
	}

	t.Logf("Metadata caching test completed successfully")
}

// TestLRUEviction тестирует вытеснение по алгоритму LRU
func TestLRUEviction(t *testing.T) {
	c := cache.NewLRUCache(2, 2, 2, 5*time.Minute) // Маленький кэш для тестирования

	// Добавляем блоки до превышения лимита
	block1 := &blockchain.Block{Height: 1, Hash: []byte("hash1")}
	block2 := &blockchain.Block{Height: 2, Hash: []byte("hash2")}
	block3 := &blockchain.Block{Height: 3, Hash: []byte("hash3")}

	c.SetBlock([]byte("hash1"), block1)
	c.SetBlock([]byte("hash2"), block2)
	c.SetBlock([]byte("hash3"), block3) // Должен вытеснить block1

	// Проверяем, что block1 вытеснен
	_, found := c.GetBlock([]byte("hash1"))
	if found {
		t.Error("Block1 should be evicted")
	}

	// Проверяем, что block2 и block3 остались
	_, found = c.GetBlock([]byte("hash2"))
	if !found {
		t.Error("Block2 should not be evicted")
	}

	_, found = c.GetBlock([]byte("hash3"))
	if !found {
		t.Error("Block3 should not be evicted")
	}

	t.Logf("LRU eviction test completed successfully")
}

// TestCacheStats тестирует статистику кэша
func TestCacheStats(t *testing.T) {
	c := cache.NewLRUCache(10, 10, 10, 5*time.Minute)

	// Проверяем начальную статистику
	stats := c.Stats()
	if stats.Hits != 0 {
		t.Errorf("Expected initial hits 0, got %d", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Expected initial misses 0, got %d", stats.Misses)
	}
	if stats.Size != 0 {
		t.Errorf("Expected initial size 0, got %d", stats.Size)
	}

	// Добавляем данные
	c.SetHeight(100)
	c.SetDifficulty(5)
	c.SetBalance("address1", 1000)

	// Проверяем статистику после добавления
	stats = c.Stats()
	if stats.Size < 1 {
		t.Errorf("Expected size at least 1, got %d", stats.Size)
	}

	// Проверяем hits
	c.GetHeight()
	c.GetDifficulty()
	c.GetBalance("address1")

	stats = c.Stats()
	if stats.Hits < 3 {
		t.Errorf("Expected at least 3 hits, got %d", stats.Hits)
	}

	t.Logf("Cache stats test completed successfully. Hits: %d, Misses: %d, Size: %d",
		stats.Hits, stats.Misses, stats.Size)
}

// TestCacheClear тестирует очистку кэша
func TestCacheClear(t *testing.T) {
	c := cache.NewLRUCache(10, 10, 10, 5*time.Minute)

	// Добавляем данные
	c.SetHeight(100)
	c.SetDifficulty(5)
	c.SetBalance("address1", 1000)
	c.SetUTXOs("address1", []*blockchain.UTXO{})

	// Проверяем, что данные есть
	stats := c.Stats()
	if stats.Size == 0 {
		t.Error("Cache should have data before clearing")
	}

	// Очищаем кэш
	c.Clear()

	// Проверяем, что кэш пуст
	stats = c.Stats()
	if stats.Size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", stats.Size)
	}

	// Проверяем, что данные недоступны
	// GetHeight и GetDifficulty всегда возвращают значения по умолчанию
	// Проверяем только баланс
	_, found := c.GetBalance("address1")
	if found {
		t.Error("Balance should not be found after clear")
	}

	t.Logf("Cache clear test completed successfully")
}

// TestCacheIntegration тестирует интеграцию кэша
func TestCacheIntegration(t *testing.T) {
	t.Run("Creation", TestLRUCacheCreation)
	t.Run("BlockCaching", TestBlockCaching)
	t.Run("UTXOCaching", TestUTXOCaching)
	t.Run("BalanceCaching", TestBalanceCaching)
	t.Run("MetadataCaching", TestMetadataCaching)
	t.Run("LRUEviction", TestLRUEviction)
	t.Run("Stats", TestCacheStats)
	t.Run("Clear", TestCacheClear)
}
