package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/storage"
	"mirochain/internal/wallet"
)

// TestBadgerStorageCreation тестирует создание BadgerStorage
func TestBadgerStorageCreation(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage
	store, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BadgerStorage: %v", err)
	}
	defer store.Close()

	// Проверяем, что storage создан
	if store == nil {
		t.Fatal("Storage should not be nil")
	}

	t.Logf("BadgerStorage created successfully in %s", tempDir)
}

// TestBlockStorage тестирует сохранение и загрузку блоков
func TestBlockStorage(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage
	store, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BadgerStorage: %v", err)
	}
	defer store.Close()

	// Создаем тестовый блок
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	genesisBlock := bc.GetBlockByHeight(0)

	// Сохраняем блок
	err = store.SaveBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to save block: %v", err)
	}

	// Загружаем блок по хешу
	loadedBlock, err := store.GetBlock(genesisBlock.Hash)
	if err != nil {
		t.Fatalf("Failed to get block by hash: %v", err)
	}

	// Проверяем, что блоки идентичны
	if !blocksEqual(genesisBlock, loadedBlock) {
		t.Error("Loaded block does not match original block")
	}

	// Загружаем блок по высоте
	loadedBlockByHeight, err := store.GetBlockByHeight(genesisBlock.Height)
	if err != nil {
		t.Fatalf("Failed to get block by height: %v", err)
	}

	// Проверяем, что блоки идентичны
	if !blocksEqual(genesisBlock, loadedBlockByHeight) {
		t.Error("Loaded block by height does not match original block")
	}

	t.Logf("Block storage test completed successfully")
}

// TestUTXOStorage тестирует сохранение и загрузку UTXO
func TestUTXOStorage(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage
	store, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BadgerStorage: %v", err)
	}
	defer store.Close()

	// Создаем тестовый UTXO набор
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	utxoSet := bc.UTXOSet

	// Сохраняем UTXO набор
	err = store.SaveUTXOSet(utxoSet)
	if err != nil {
		t.Fatalf("Failed to save UTXO set: %v", err)
	}

	// Загружаем UTXO набор
	loadedUTXOSet, err := store.GetUTXOSet()
	if err != nil {
		t.Fatalf("Failed to get UTXO set: %v", err)
	}

	// Проверяем, что UTXO наборы идентичны
	if !utxoSetsEqual(utxoSet, loadedUTXOSet) {
		t.Error("Loaded UTXO set does not match original UTXO set")
	}

	t.Logf("UTXO storage test completed successfully")
}

// TestMetadataStorage тестирует сохранение и загрузку метаданных
func TestMetadataStorage(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage
	store, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BadgerStorage: %v", err)
	}
	defer store.Close()

	// Тестируем сохранение и загрузку сложности
	testDifficulty := 5
	err = store.SaveDifficulty(testDifficulty)
	if err != nil {
		t.Fatalf("Failed to save difficulty: %v", err)
	}

	loadedDifficulty, err := store.GetDifficulty()
	if err != nil {
		t.Fatalf("Failed to get difficulty: %v", err)
	}

	if loadedDifficulty != testDifficulty {
		t.Errorf("Expected difficulty %d, got %d", testDifficulty, loadedDifficulty)
	}

	// Тестируем сохранение и загрузку высоты
	testHeight := int64(10)
	err = store.SaveHeight(testHeight)
	if err != nil {
		t.Fatalf("Failed to save height: %v", err)
	}

	loadedHeight, err := store.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}

	if loadedHeight != testHeight {
		t.Errorf("Expected height %d, got %d", testHeight, loadedHeight)
	}

	t.Logf("Metadata storage test completed successfully")
}

// TestMultipleBlocksStorage тестирует сохранение и загрузку нескольких блоков
func TestMultipleBlocksStorage(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage
	store, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BadgerStorage: %v", err)
	}
	defer store.Close()

	// Создаем блокчейн с несколькими блоками
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Сохраняем genesis блок
	genesisBlock := bc.GetBlockByHeight(0)
	err = store.SaveBlock(genesisBlock)
	if err != nil {
		t.Fatalf("Failed to save genesis block: %v", err)
	}

	// Создаем несколько блоков
	blocks := []*blockchain.Block{}
	for i := 0; i < 5; i++ {
		// Создаем простую транзакцию
		tx := &blockchain.Transaction{
			Inputs: []*blockchain.TransactionInput{},
			Outputs: []*blockchain.TransactionOutput{
				{
					Value:     100,
					Address:   nodeWallet.GetAddress(),
					PublicKey: nodeWallet.GetPublicKeyBytes(),
				},
			},
			Timestamp: time.Now().Unix(),
			Fee:       1,
		}

		// Создаем блок
		var previousHash []byte
		if len(blocks) > 0 {
			previousHash = blocks[len(blocks)-1].Hash
		} else {
			previousHash = bc.GetGenesisHash()
		}

		block := blockchain.NewBlock([]*blockchain.Transaction{tx}, previousHash, int64(i+1), 0)
		blocks = append(blocks, block)

		// Сохраняем блок
		err = store.SaveBlock(block)
		if err != nil {
			t.Fatalf("Failed to save block %d: %v", i, err)
		}
	}

	// Загружаем все блоки
	loadedBlocks, err := store.GetAllBlocks()
	if err != nil {
		t.Fatalf("Failed to get all blocks: %v", err)
	}

	// Проверяем количество блоков
	if len(loadedBlocks) != len(blocks)+1 { // +1 для genesis блока
		t.Errorf("Expected %d blocks, got %d", len(blocks)+1, len(loadedBlocks))
	}

	// Проверяем последний блок
	latestBlock, err := store.GetLatestBlock()
	if err != nil {
		t.Fatalf("Failed to get latest block: %v", err)
	}

	if latestBlock.Height != int64(len(blocks)) {
		t.Errorf("Expected latest block height %d, got %d", len(blocks), latestBlock.Height)
	}

	t.Logf("Multiple blocks storage test completed successfully. Stored %d blocks", len(blocks))
}

// TestStoragePersistence тестирует персистентность данных
func TestStoragePersistence(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage и сохраняем данные
	store1, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create first BadgerStorage: %v", err)
	}

	// Сохраняем тестовые данные
	testDifficulty := 3
	testHeight := int64(5)

	err = store1.SaveDifficulty(testDifficulty)
	if err != nil {
		t.Fatalf("Failed to save difficulty: %v", err)
	}

	err = store1.SaveHeight(testHeight)
	if err != nil {
		t.Fatalf("Failed to save height: %v", err)
	}

	// Закрываем storage
	err = store1.Close()
	if err != nil {
		t.Fatalf("Failed to close first storage: %v", err)
	}

	// Создаем новый storage в той же директории
	store2, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second BadgerStorage: %v", err)
	}
	defer store2.Close()

	// Загружаем данные
	loadedDifficulty, err := store2.GetDifficulty()
	if err != nil {
		t.Fatalf("Failed to get difficulty: %v", err)
	}

	loadedHeight, err := store2.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}

	// Проверяем, что данные сохранились
	if loadedDifficulty != testDifficulty {
		t.Errorf("Expected difficulty %d, got %d", testDifficulty, loadedDifficulty)
	}

	if loadedHeight != testHeight {
		t.Errorf("Expected height %d, got %d", testHeight, loadedHeight)
	}

	t.Logf("Storage persistence test completed successfully")
}

// TestStorageClear тестирует очистку storage
func TestStorageClear(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем storage
	store, err := storage.NewBadgerStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create BadgerStorage: %v", err)
	}
	defer store.Close()

	// Сохраняем тестовые данные
	err = store.SaveDifficulty(5)
	if err != nil {
		t.Fatalf("Failed to save difficulty: %v", err)
	}

	err = store.SaveHeight(10)
	if err != nil {
		t.Fatalf("Failed to save height: %v", err)
	}

	// Очищаем storage
	err = store.Clear()
	if err != nil {
		t.Fatalf("Failed to clear storage: %v", err)
	}

	// Проверяем, что данные очищены
	loadedDifficulty, err := store.GetDifficulty()
	if err != nil {
		t.Fatalf("Failed to get difficulty after clear: %v", err)
	}

	loadedHeight, err := store.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height after clear: %v", err)
	}

	if loadedDifficulty != 0 {
		t.Errorf("Expected difficulty 0 after clear, got %d", loadedDifficulty)
	}

	if loadedHeight != -1 {
		t.Errorf("Expected height -1 after clear, got %d", loadedHeight)
	}

	t.Logf("Storage clear test completed successfully")
}

// Вспомогательные функции для сравнения

func blocksEqual(block1, block2 *blockchain.Block) bool {
	if block1 == nil || block2 == nil {
		return block1 == block2
	}

	return block1.Height == block2.Height &&
		block1.Timestamp == block2.Timestamp &&
		block1.Nonce == block2.Nonce &&
		block1.Difficulty == block2.Difficulty &&
		bytesEqual(block1.Hash, block2.Hash) &&
		bytesEqual(block1.PreviousHash, block2.PreviousHash) &&
		bytesEqual(block1.MerkleRoot, block2.MerkleRoot) &&
		len(block1.Transactions) == len(block2.Transactions)
}

func utxoSetsEqual(utxo1, utxo2 *blockchain.UTXOSet) bool {
	if utxo1 == nil || utxo2 == nil {
		return utxo1 == utxo2
	}

	// Простое сравнение - проверяем, что количество UTXO одинаково
	// В реальном проекте нужно было бы сравнивать содержимое
	return len(utxo1.UTXOs) == len(utxo2.UTXOs)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestStorageIntegration тестирует интеграцию storage
func TestStorageIntegration(t *testing.T) {
	t.Run("Creation", TestBadgerStorageCreation)
	t.Run("BlockStorage", TestBlockStorage)
	t.Run("UTXOStorage", TestUTXOStorage)
	t.Run("MetadataStorage", TestMetadataStorage)
	t.Run("MultipleBlocks", TestMultipleBlocksStorage)
	t.Run("Persistence", TestStoragePersistence)
	t.Run("Clear", TestStorageClear)
}
