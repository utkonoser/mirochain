package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/wallet"
)

// TestEmptyBlockchain тестирует пустой блокчейн
func TestEmptyBlockchain(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Проверяем, что блокчейн имеет только genesis блок
	stats := bc.GetStats()
	if stats["height"].(int64) != 0 {
		t.Errorf("Expected height 0, got %d", stats["height"])
	}

	// Проверяем, что genesis блок валиден
	genesisBlock := bc.GetBlockByHeight(0)
	if genesisBlock == nil {
		t.Fatal("Genesis block should not be nil")
	}

	if !genesisBlock.IsValid(nil) {
		t.Error("Genesis block should be valid")
	}
}

// TestInvalidBlockHash тестирует блок с неверным хешем
func TestInvalidBlockHash(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию с правильной структурой
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

	// Создаем блок с неверным хешем
	block := &blockchain.Block{
		Height:       1,
		Timestamp:    time.Now().Unix(),
		Transactions: []*blockchain.Transaction{tx},
		PreviousHash: bc.GetGenesisHash(),
		Nonce:        0,
		Difficulty:   1,
		Hash:         []byte("invalid_hash"),
	}

	// Проверяем, что блок невалиден
	if block.IsValid(bc.GetBlockByHeight(0)) {
		t.Error("Block with invalid hash should be invalid")
	}
}

// TestInvalidPreviousHash тестирует блок с неверным previous hash
func TestInvalidPreviousHash(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию
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

	// Создаем блок с неверным previous hash
	block := &blockchain.Block{
		Height:       1,
		Timestamp:    time.Now().Unix(),
		Transactions: []*blockchain.Transaction{tx},
		PreviousHash: []byte("invalid_previous_hash"),
		Nonce:        0,
		Difficulty:   1,
	}

	// Проверяем, что блок невалиден
	if block.IsValid(bc.GetBlockByHeight(0)) {
		t.Error("Block with invalid previous hash should be invalid")
	}
}

// TestInvalidBlockHeight тестирует блок с неверной высотой
func TestInvalidBlockHeight(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию
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

	// Создаем блок с неверной высотой
	block := &blockchain.Block{
		Height:       5, // Неправильная высота
		Timestamp:    time.Now().Unix(),
		Transactions: []*blockchain.Transaction{tx},
		PreviousHash: bc.GetGenesisHash(),
		Nonce:        0,
		Difficulty:   1,
	}

	// Проверяем, что блок невалиден
	if block.IsValid(bc.GetBlockByHeight(0)) {
		t.Error("Block with invalid height should be invalid")
	}
}

// TestEmptyTransaction тестирует пустую транзакцию
func TestEmptyTransaction(t *testing.T) {
	// Создаем пустую транзакцию
	tx := &blockchain.Transaction{
		Inputs:    []*blockchain.TransactionInput{},
		Outputs:   []*blockchain.TransactionOutput{},
		Timestamp: time.Now().Unix(),
		Fee:       0,
	}

	// Проверяем, что транзакция невалидна
	if tx.IsValid() {
		t.Error("Empty transaction should be invalid")
	}
}

// TestInvalidTransactionSignature тестирует транзакцию с неверной подписью
func TestInvalidTransactionSignature(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем транзакцию с неверной подписью
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{
			{
				TransactionID: []byte("invalid_tx_id"),
				OutputIndex:   0,
				Signature:     []byte("invalid_signature"),
				PublicKey:     nodeWallet.GetPublicKeyBytes(),
			},
		},
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

	// Проверяем, что транзакция невалидна
	if tx.IsValid() {
		t.Error("Transaction with invalid signature should be invalid")
	}
}

// TestNegativeTransactionAmount тестирует транзакцию с отрицательной суммой
func TestNegativeTransactionAmount(t *testing.T) {
	// Создаем транзакцию с отрицательной суммой
	tx := &blockchain.Transaction{
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

	// Проверяем, что транзакция невалидна
	if tx.IsValid() {
		t.Error("Transaction with negative amount should be invalid")
	}
}

// TestZeroTransactionAmount тестирует транзакцию с нулевой суммой
func TestZeroTransactionAmount(t *testing.T) {
	// Создаем транзакцию с нулевой суммой
	tx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     0, // Нулевая сумма
				Address:   "recipient_address",
				PublicKey: []byte("recipient_public_key"),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Проверяем, что транзакция невалидна
	if tx.IsValid() {
		t.Error("Transaction with zero amount should be invalid")
	}
}

// TestInvalidMerkleRoot тестирует блок с неверным Merkle root
func TestInvalidMerkleRoot(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем транзакцию
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

	// Создаем блок с неверным Merkle root
	block := &blockchain.Block{
		Height:       1,
		Timestamp:    time.Now().Unix(),
		Transactions: []*blockchain.Transaction{tx},
		PreviousHash: bc.GetGenesisHash(),
		Nonce:        0,
		Difficulty:   1,
		MerkleRoot:   []byte("invalid_merkle_root"),
	}

	// Проверяем, что блок невалиден
	if block.IsValid(bc.GetBlockByHeight(0)) {
		t.Error("Block with invalid Merkle root should be invalid")
	}
}

// TestBlockchainEdgeCases тестирует граничные случаи блокчейна
func TestBlockchainEdgeCases(t *testing.T) {
	t.Run("EmptyBlockchain", TestEmptyBlockchain)
	t.Run("InvalidBlockHash", TestInvalidBlockHash)
	t.Run("InvalidPreviousHash", TestInvalidPreviousHash)
	t.Run("InvalidBlockHeight", TestInvalidBlockHeight)
	t.Run("EmptyTransaction", TestEmptyTransaction)
	t.Run("InvalidTransactionSignature", TestInvalidTransactionSignature)
	t.Run("NegativeTransactionAmount", TestNegativeTransactionAmount)
	t.Run("ZeroTransactionAmount", TestZeroTransactionAmount)
	t.Run("InvalidMerkleRoot", TestInvalidMerkleRoot)
}
