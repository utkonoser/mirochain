package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/wallet"
)

// TestWalletWithEmptyBalance тестирует кошелек с нулевым балансом
func TestWalletWithEmptyBalance(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Проверяем, что баланс равен 1000000 (genesis блок создает начальный баланс)
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}

	t.Logf("Wallet with empty balance test completed")
}

// TestWalletWithInvalidAddress тестирует кошелек с невалидным адресом
func TestWalletWithInvalidAddress(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Проверяем баланс для невалидного адреса
	invalidAddress := "invalid_address"
	balance := bc.GetBalance(invalidAddress)
	if balance != 0 {
		t.Errorf("Expected balance 0 for invalid address, got %d", balance)
	}

	t.Logf("Wallet with invalid address test completed")
}

// TestWalletWithNegativeBalance тестирует кошелек с отрицательным балансом
func TestWalletWithNegativeBalance(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с отрицательным балансом
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     -100, // Отрицательная сумма
				Address:   nodeWallet.GetAddress(),
				PublicKey: nodeWallet.GetPublicKeyBytes(),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Добавляем транзакцию в блокчейн
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 0)
	bc.AddBlock(block)

	// Проверяем, что баланс остался положительным (genesis блок создает начальный баланс)
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance < 0 {
		t.Errorf("Balance should not be negative, got %d", balance)
	}

	t.Logf("Wallet with negative balance test completed")
}

// TestWalletWithZeroBalance тестирует кошелек с нулевым балансом
func TestWalletWithZeroBalance(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с нулевым балансом
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     0, // Нулевая сумма
				Address:   nodeWallet.GetAddress(),
				PublicKey: nodeWallet.GetPublicKeyBytes(),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Добавляем транзакцию в блокчейн
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 0)
	bc.AddBlock(block)

	// Проверяем, что баланс остался 1000000 (genesis блок создает начальный баланс)
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}

	t.Logf("Wallet with zero balance test completed")
}

// TestWalletWithLargeBalance тестирует кошелек с большим балансом
func TestWalletWithLargeBalance(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с большим балансом
	largeAmount := int64(1000000000) // 1 миллиард
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     largeAmount,
				Address:   nodeWallet.GetAddress(),
				PublicKey: nodeWallet.GetPublicKeyBytes(),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Добавляем транзакцию в блокчейн
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 0)
	bc.AddBlock(block)

	// Проверяем, что баланс остался genesis балансом (невалидная транзакция не добавилась)
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000 (genesis only), got %d", balance)
	}

	t.Logf("Wallet with large balance test completed. Balance: %d", balance)
}

// TestWalletWithInvalidPublicKey тестирует кошелек с невалидным публичным ключом
func TestWalletWithInvalidPublicKey(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с невалидным публичным ключом
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     100,
				Address:   nodeWallet.GetAddress(),
				PublicKey: []byte("invalid_public_key"), // Невалидный ключ
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Добавляем транзакцию в блокчейн
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 0)
	bc.AddBlock(block)

	// Проверяем, что баланс остался genesis балансом
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000 with invalid public key, got %d", balance)
	}

	t.Logf("Wallet with invalid public key test completed")
}

// TestWalletWithInvalidSignature тестирует кошелек с невалидной подписью
func TestWalletWithInvalidSignature(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с невалидной подписью
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{
			{
				TransactionID: []byte("invalid_tx_id"),
				OutputIndex:   0,
				Signature:     []byte("invalid_signature"), // Невалидная подпись
				PublicKey:     nodeWallet.GetPublicKeyBytes(),
			},
		},
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

	// Добавляем транзакцию в блокчейн
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 0)
	bc.AddBlock(block)

	// Проверяем, что баланс остался genesis балансом
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000 with invalid signature, got %d", balance)
	}

	t.Logf("Wallet with invalid signature test completed")
}

// TestWalletWithDuplicateAddress тестирует кошелек с дублирующимся адресом
func TestWalletWithDuplicateAddress(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с дублирующимся адресом
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     100,
				Address:   nodeWallet.GetAddress(),
				PublicKey: nodeWallet.GetPublicKeyBytes(),
			},
			{
				Value:     200,
				Address:   nodeWallet.GetAddress(), // Дублирующийся адрес
				PublicKey: nodeWallet.GetPublicKeyBytes(),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Добавляем транзакцию в блокчейн
	block := blockchain.NewBlock([]*blockchain.Transaction{tx}, bc.GetGenesisHash(), 1, 0)
	bc.AddBlock(block)

	// Проверяем, что баланс остался genesis балансом (невалидная транзакция не добавилась)
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000 (genesis only), got %d", balance)
	}

	t.Logf("Wallet with duplicate address test completed. Balance: %d", balance)
}

// TestWalletEdgeCases тестирует граничные случаи кошельков
func TestWalletEdgeCases(t *testing.T) {
	t.Run("EmptyBalance", TestWalletWithEmptyBalance)
	t.Run("InvalidAddress", TestWalletWithInvalidAddress)
	t.Run("NegativeBalance", TestWalletWithNegativeBalance)
	t.Run("ZeroBalance", TestWalletWithZeroBalance)
	t.Run("LargeBalance", TestWalletWithLargeBalance)
	t.Run("InvalidPublicKey", TestWalletWithInvalidPublicKey)
	t.Run("InvalidSignature", TestWalletWithInvalidSignature)
	t.Run("DuplicateAddress", TestWalletWithDuplicateAddress)
}
