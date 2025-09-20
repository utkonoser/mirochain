package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/network"
	"mirochain/internal/wallet"
)

// TestBlockchainErrorHandling тестирует обработку ошибок в блокчейне
func TestBlockchainErrorHandling(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Тестируем добавление невалидного блока
	invalidBlock := &blockchain.Block{
		Height:       1,
		Timestamp:    time.Now().Unix(),
		Transactions: []*blockchain.Transaction{},
		PreviousHash: []byte("invalid_previous_hash"),
		Nonce:        0,
		Difficulty:   1,
		Hash:         []byte("invalid_hash"),
	}

	// Попытка добавить невалидный блок должна вернуть ошибку
	err = bc.AddBlock(invalidBlock)
	if err == nil {
		t.Error("Expected error when adding invalid block")
	}

	// Тестируем получение блока по невалидной высоте
	block := bc.GetBlockByHeight(-1)
	if block != nil {
		t.Error("Expected nil block for invalid height")
	}

	block = bc.GetBlockByHeight(999)
	if block != nil {
		t.Error("Expected nil block for non-existent height")
	}

	t.Logf("Blockchain error handling test completed")
}

// TestMiningErrorHandling тестирует обработку ошибок в майнинге
func TestMiningErrorHandling(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)
	mempool := mining.NewMempool(100)

	// Создаем майнер
	miner := mining.NewMiner(
		nodeWallet.GetAddress(),
		nodeWallet.GetPublicKeyBytes(),
		bc,
		mempool,
		nil,
		nodeWallet,
	)

	// Тестируем двойной запуск майнинга
	err = miner.StartMining()
	if err != nil {
		t.Fatalf("Failed to start mining: %v", err)
	}

	// Попытка запустить майнинг повторно должна вернуть ошибку
	err = miner.StartMining()
	if err == nil {
		t.Error("Expected error when starting mining twice")
	}

	// Останавливаем майнинг
	err = miner.StopMining()
	if err != nil {
		t.Fatalf("Failed to stop mining: %v", err)
	}

	// Попытка остановить майнинг повторно должна вернуть ошибку
	err = miner.StopMining()
	if err == nil {
		t.Error("Expected error when stopping mining twice")
	}

	t.Logf("Mining error handling test completed")
}

// TestNetworkErrorHandling тестирует обработку ошибок в сети
func TestNetworkErrorHandling(t *testing.T) {
	// Тестируем создание сервера с невалидным адресом
	server := network.NewServer("invalid_address", 8080, nil)
	if server == nil {
		t.Error("Server should be created even with invalid address")
	}

	// Тестируем создание клиента с невалидным сервером
	client := network.NewClient(nil)
	if client == nil {
		t.Error("Client should be created even with nil server")
	}

	t.Logf("Network error handling test completed")
}

// TestWalletErrorHandling тестирует обработку ошибок в кошельках
func TestWalletErrorHandling(t *testing.T) {
	walletManager := wallet.NewWalletManager()

	// Тестируем создание кошелька с невалидными параметрами
	wallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Тестируем подпись с невалидными данными
	invalidData := []byte("")
	_, err = wallet.Sign(invalidData)
	if err != nil {
		t.Logf("Expected error when signing invalid data: %v", err)
	}

	// Тестируем проверку подписи с невалидными данными
	validData := []byte("valid data")
	validSignature, err := wallet.Sign(validData)
	if err != nil {
		t.Fatalf("Failed to sign valid data: %v", err)
	}

	// Проверяем подпись с невалидными данными
	isValid := wallet.VerifySignature(invalidData, validSignature)
	if isValid {
		t.Error("Expected invalid signature for invalid data")
	}

	// Проверяем подпись с невалидной подписью
	isValid = wallet.VerifySignature(validData, []byte("invalid_signature"))
	if isValid {
		t.Error("Expected invalid signature for invalid signature")
	}

	t.Logf("Wallet error handling test completed")
}

// TestTransactionErrorHandling тестирует обработку ошибок в транзакциях
func TestTransactionErrorHandling(t *testing.T) {
	// Тестируем создание транзакции с невалидными данными
	invalidTx := &blockchain.Transaction{
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Value:     -100, // Отрицательная сумма
				Address:   "",
				PublicKey: []byte(""),
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}

	// Проверяем, что транзакция невалидна
	if invalidTx.IsValid() {
		t.Error("Expected invalid transaction to be invalid")
	}

	// Тестируем создание транзакции с пустыми входами и выходами
	emptyTx := &blockchain.Transaction{
		Inputs:    []*blockchain.TransactionInput{},
		Outputs:   []*blockchain.TransactionOutput{},
		Timestamp: time.Now().Unix(),
		Fee:       0,
	}

	// Проверяем, что пустая транзакция невалидна
	if emptyTx.IsValid() {
		t.Error("Expected empty transaction to be invalid")
	}

	t.Logf("Transaction error handling test completed")
}

// TestMempoolErrorHandling тестирует обработку ошибок в mempool
func TestMempoolErrorHandling(t *testing.T) {
	// Создаем mempool с нулевым размером
	mempool := mining.NewMempool(0)

	// Тестируем добавление транзакции в mempool с нулевым размером
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

	// Попытка добавить транзакцию в mempool с нулевым размером должна вернуть ошибку
	err := mempool.AddTransaction(tx)
	if err == nil {
		t.Error("Expected error when adding transaction to zero-size mempool")
	}

	// Тестируем добавление nil транзакции
	err = mempool.AddTransaction(nil)
	if err == nil {
		t.Error("Expected error when adding nil transaction")
	}

	t.Logf("Mempool error handling test completed")
}

// TestErrorHandling тестирует обработку ошибок для всех компонентов
func TestErrorHandling(t *testing.T) {
	t.Run("Blockchain", TestBlockchainErrorHandling)
	t.Run("Mining", TestMiningErrorHandling)
	t.Run("Network", TestNetworkErrorHandling)
	t.Run("Wallet", TestWalletErrorHandling)
	t.Run("Transaction", TestTransactionErrorHandling)
	t.Run("Mempool", TestMempoolErrorHandling)
}
