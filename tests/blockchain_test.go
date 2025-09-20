package tests

import (
	"testing"

	"mirochain/internal/blockchain"
	"mirochain/internal/wallet"
)

func TestBlockchain(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 4)

	// Проверяем, что блокчейн создан
	if bc == nil {
		t.Fatal("Blockchain is nil")
	}

	// Проверяем genesis блок
	if bc.GetHeight() != 0 {
		t.Errorf("Expected height 0, got %d", bc.GetHeight())
	}

	// Проверяем валидность блокчейна
	if !bc.IsValid() {
		t.Error("Blockchain is not valid")
	}

	// Проверяем баланс
	balance := bc.GetBalance(nodeWallet.GetAddress())
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}

	t.Logf("Blockchain created successfully. Height: %d, Balance: %d", bc.GetHeight(), balance)
}

func TestWallet(t *testing.T) {
	// Создаем кошелек
	walletManager := wallet.NewWalletManager()
	wallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Проверяем, что кошелек создан
	if wallet == nil {
		t.Fatal("Wallet is nil")
	}

	// Проверяем адрес
	address := wallet.GetAddress()
	if address == "" {
		t.Error("Wallet address is empty")
	}

	// Проверяем публичный ключ
	publicKey := wallet.GetPublicKeyBytes()
	if len(publicKey) == 0 {
		t.Error("Wallet public key is empty")
	}

	// Проверяем приватный ключ
	privateKey := wallet.GetPrivateKeyBytes()
	if len(privateKey) == 0 {
		t.Error("Wallet private key is empty")
	}

	t.Logf("Wallet created successfully. Address: %s", address)
}

func TestTransaction(t *testing.T) {
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

	// Проверяем баланс первого кошелька
	balance := bc.GetBalance(wallet1.GetAddress())
	if balance == 0 {
		t.Error("Wallet1 should have balance from genesis block")
	}

	// Создаем транзакцию с меньшей суммой
	tx, err := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 100, wallet1.GetPrivateKeyBytes())
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Проверяем транзакцию
	if tx == nil {
		t.Fatal("Transaction is nil")
	}

	// Проверяем валидность транзакции
	if !tx.IsValid() {
		t.Error("Transaction is not valid")
	}

	t.Logf("Transaction created successfully. ID: %x", tx.ID)
}
