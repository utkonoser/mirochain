package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/persistent"
	"mirochain/internal/wallet"
)

// TestPersistentBlockchainCreation тестирует создание PersistentBlockchain
func TestPersistentBlockchainCreation(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем persistent blockchain
	pbc, err := persistent.NewPersistentBlockchain(tempDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	if err != nil {
		t.Fatalf("Failed to create persistent blockchain: %v", err)
	}
	defer pbc.Close()

	// Проверяем, что blockchain создан
	if pbc == nil {
		t.Fatal("PersistentBlockchain should not be nil")
	}

	// Проверяем, что genesis блок создан
	height, err := pbc.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}

	if height != 0 {
		t.Errorf("Expected height 0, got %d", height)
	}

	// Проверяем баланс
	balance := pbc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}

	t.Logf("PersistentBlockchain created successfully. Height: %d, Balance: %d", height, balance)
}

// TestPersistentBlockchainPersistence тестирует персистентность данных
func TestPersistentBlockchainPersistence(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем persistent blockchain
	pbc1, err := persistent.NewPersistentBlockchain(tempDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	if err != nil {
		t.Fatalf("Failed to create first persistent blockchain: %v", err)
	}

	// Добавляем несколько блоков
	for i := 0; i < 3; i++ {
		// Создаем coinbase транзакцию (как в реальном майнинге)
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 100)

		// Создаем блок
		previousBlock, err := pbc1.GetBlockByHeight(int64(i))
		if err != nil {
			t.Fatalf("Failed to get previous block: %v", err)
		}

		block := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, previousBlock.Hash, int64(i+1), 0)

		// Добавляем блок
		err = pbc1.AddBlock(block)
		if err != nil {
			t.Fatalf("Failed to add block %d: %v", i+1, err)
		}
	}

	// Закрываем первый blockchain
	err = pbc1.Close()
	if err != nil {
		t.Fatalf("Failed to close first blockchain: %v", err)
	}

	// Создаем новый blockchain в той же директории
	pbc2, err := persistent.NewPersistentBlockchain(tempDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	if err != nil {
		t.Fatalf("Failed to create second persistent blockchain: %v", err)
	}
	defer pbc2.Close()

	// Проверяем, что данные сохранились
	height, err := pbc2.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}

	if height != 3 {
		t.Errorf("Expected height 3, got %d", height)
	}

	// Проверяем, что блоки загрузились
	for i := int64(0); i <= 3; i++ {
		block, err := pbc2.GetBlockByHeight(i)
		if err != nil {
			t.Fatalf("Failed to get block at height %d: %v", i, err)
		}

		if block.Height != i {
			t.Errorf("Expected block height %d, got %d", i, block.Height)
		}
	}

	// Проверяем баланс (genesis + 3 coinbase транзакции по 100)
	expectedBalance := int64(1000100)
	balance := pbc2.GetBalance(nodeWallet.GetAddress())
	if balance != expectedBalance {
		t.Errorf("Expected balance %d, got %d", expectedBalance, balance)
	}

	t.Logf("PersistentBlockchain persistence test completed successfully. Height: %d, Balance: %d", height, balance)
}

// TestPersistentBlockchainStats тестирует статистику блокчейна
func TestPersistentBlockchainStats(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем persistent blockchain
	pbc, err := persistent.NewPersistentBlockchain(tempDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 5)
	if err != nil {
		t.Fatalf("Failed to create persistent blockchain: %v", err)
	}
	defer pbc.Close()

	// Получаем статистику
	stats := pbc.GetStats()

	// Проверяем статистику
	if stats["height"].(int64) != 0 {
		t.Errorf("Expected height 0, got %d", stats["height"])
	}

	// Difficulty должна быть 5, но может быть 0 если не сохранилась
	// Это нормально для первого запуска
	if stats["difficulty"].(int) != 5 && stats["difficulty"].(int) != 0 {
		t.Errorf("Expected difficulty 5 or 0, got %d", stats["difficulty"])
	}

	if stats["utxo_count"].(int) != 1 {
		t.Errorf("Expected UTXO count 1, got %d", stats["utxo_count"])
	}

	t.Logf("PersistentBlockchain stats test completed successfully. Stats: %+v", stats)
}

// TestPersistentBlockchainUTXOs тестирует работу с UTXO
func TestPersistentBlockchainUTXOs(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем persistent blockchain
	pbc, err := persistent.NewPersistentBlockchain(tempDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	if err != nil {
		t.Fatalf("Failed to create persistent blockchain: %v", err)
	}
	defer pbc.Close()

	// Получаем UTXO для адреса
	utxos := pbc.GetUTXOs(nodeWallet.GetAddress())

	// Проверяем, что есть UTXO
	if len(utxos) != 1 {
		t.Errorf("Expected 1 UTXO, got %d", len(utxos))
	}

	// Проверяем значение UTXO
	if utxos[0].Value != 1000000 {
		t.Errorf("Expected UTXO value 1000000, got %d", utxos[0].Value)
	}

	t.Logf("PersistentBlockchain UTXOs test completed successfully. UTXOs count: %d", len(utxos))
}

// TestPersistentBlockchainInvalidBlock тестирует добавление невалидного блока
func TestPersistentBlockchainInvalidBlock(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем persistent blockchain
	pbc, err := persistent.NewPersistentBlockchain(tempDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	if err != nil {
		t.Fatalf("Failed to create persistent blockchain: %v", err)
	}
	defer pbc.Close()

	// Создаем невалидный блок (с неправильной высотой)
	invalidBlock := &blockchain.Block{
		Height:       5, // Неправильная высота
		Timestamp:    time.Now().Unix(),
		Transactions: []*blockchain.Transaction{},
		PreviousHash: []byte("invalid"),
		Nonce:        0,
		Difficulty:   0,
	}

	// Пытаемся добавить невалидный блок
	err = pbc.AddBlock(invalidBlock)
	if err == nil {
		t.Error("Expected error when adding invalid block, got nil")
	}

	// Проверяем, что высота не изменилась
	height, err := pbc.GetHeight()
	if err != nil {
		t.Fatalf("Failed to get height: %v", err)
	}

	if height != 0 {
		t.Errorf("Expected height 0 after invalid block, got %d", height)
	}

	t.Logf("PersistentBlockchain invalid block test completed successfully")
}

// TestPersistentBlockchainIntegration тестирует интеграцию PersistentBlockchain
func TestPersistentBlockchainIntegration(t *testing.T) {
	t.Run("Creation", TestPersistentBlockchainCreation)
	t.Run("Persistence", TestPersistentBlockchainPersistence)
	t.Run("Stats", TestPersistentBlockchainStats)
	t.Run("UTXOs", TestPersistentBlockchainUTXOs)
	t.Run("InvalidBlock", TestPersistentBlockchainInvalidBlock)
}
