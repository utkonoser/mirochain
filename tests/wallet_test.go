package tests

import (
	"testing"

	"mirochain/internal/wallet"
)

func TestWalletCreation(t *testing.T) {
	// Создаем новый кошелек
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Проверяем, что кошелек создан
	if w == nil {
		t.Fatal("Wallet is nil")
	}

	// Проверяем, что адрес не пустой
	if w.GetAddress() == "" {
		t.Fatal("Wallet address is empty")
	}

	// Проверяем, что ключи не пустые
	if len(w.GetPublicKeyBytes()) == 0 {
		t.Fatal("Public key is empty")
	}

	if len(w.GetPrivateKeyBytes()) == 0 {
		t.Fatal("Private key is empty")
	}

	t.Logf("Wallet created successfully. Address: %s", w.GetAddress())
}

func TestWalletManager(t *testing.T) {
	// Создаем менеджер кошельков
	wm := wallet.NewWalletManager()

	// Проверяем, что менеджер создан
	if wm == nil {
		t.Fatal("Wallet manager is nil")
	}

	// Проверяем начальное количество кошельков
	if wm.GetWalletCount() != 0 {
		t.Fatalf("Expected 0 wallets, got %d", wm.GetWalletCount())
	}

	// Создаем кошелек
	wallet1, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Проверяем, что кошелек добавлен
	if wm.GetWalletCount() != 1 {
		t.Fatalf("Expected 1 wallet, got %d", wm.GetWalletCount())
	}

	// Создаем второй кошелек
	wallet2, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create second wallet: %v", err)
	}

	// Проверяем, что адреса разные
	if wallet1.GetAddress() == wallet2.GetAddress() {
		t.Fatal("Wallet addresses should be different")
	}

	// Проверяем получение кошелька по адресу
	retrievedWallet, exists := wm.GetWallet(wallet1.GetAddress())
	if !exists {
		t.Fatal("Wallet should exist")
	}

	if retrievedWallet.GetAddress() != wallet1.GetAddress() {
		t.Fatal("Retrieved wallet address mismatch")
	}

	// Проверяем список адресов
	addresses := wm.ListAddresses()
	if len(addresses) != 2 {
		t.Fatalf("Expected 2 addresses, got %d", len(addresses))
	}

	t.Logf("Wallet manager test completed. Count: %d", wm.GetWalletCount())
}

func TestWalletSigning(t *testing.T) {
	// Создаем кошелек
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Тестовое сообщение
	message := []byte("Hello, MiroChain!")

	// Подписываем сообщение
	signature, err := w.SignMessage(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// Проверяем, что подпись не пустая
	if len(signature) == 0 {
		t.Fatal("Signature is empty")
	}

	// Проверяем подпись
	isValid := w.VerifySignature(message, signature)
	if !isValid {
		t.Fatal("Signature verification failed")
	}

	// Проверяем с неправильным сообщением
	wrongMessage := []byte("Wrong message")
	isValid = w.VerifySignature(wrongMessage, signature)
	if isValid {
		t.Fatal("Signature should be invalid for wrong message")
	}

	t.Logf("Wallet signing test completed. Signature length: %d", len(signature))
}

func TestWalletPersistence(t *testing.T) {
	// Создаем менеджер кошельков с временной директорией
	wm := wallet.NewWalletManagerWithDataDir("./test_data/wallets")

	// Создаем кошелек
	wallet1, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Сохраняем кошелек
	err = wm.SaveWallet(wallet1)
	if err != nil {
		t.Fatalf("Failed to save wallet: %v", err)
	}

	// Сохраняем общий список
	err = wm.SaveWallets()
	if err != nil {
		t.Fatalf("Failed to save wallets: %v", err)
	}

	// Создаем новый менеджер и загружаем кошельки
	wm2 := wallet.NewWalletManagerWithDataDir("./test_data/wallets")
	err = wm2.LoadWallets()
	if err != nil {
		t.Fatalf("Failed to load wallets: %v", err)
	}

	// Проверяем, что кошелек загружен
	if wm2.GetWalletCount() != 1 {
		t.Fatalf("Expected 1 wallet after loading, got %d", wm2.GetWalletCount())
	}

	// Проверяем, что адрес совпадает
	loadedWallet, exists := wm2.GetWallet(wallet1.GetAddress())
	if !exists {
		t.Fatal("Loaded wallet should exist")
	}

	if loadedWallet.GetAddress() != wallet1.GetAddress() {
		t.Fatal("Loaded wallet address mismatch")
	}

	t.Logf("Wallet persistence test completed. Address: %s", loadedWallet.GetAddress())
}

func TestWalletBalance(t *testing.T) {
	// Создаем кошелек
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Получаем баланс (заглушка)
	balance, err := w.GetBalance(nil)
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	// Проверяем, что баланс не отрицательный
	if balance < 0 {
		t.Fatal("Balance should not be negative")
	}

	t.Logf("Wallet balance test completed. Balance: %d", balance)
}
