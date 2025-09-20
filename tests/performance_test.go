package tests

import (
	"testing"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/wallet"
)

func BenchmarkWalletCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := wallet.NewWallet()
		if err != nil {
			b.Fatalf("Failed to create wallet: %v", err)
		}
	}
}

func BenchmarkWalletSigning(b *testing.B) {
	w, err := wallet.NewWallet()
	if err != nil {
		b.Fatalf("Failed to create wallet: %v", err)
	}

	message := []byte("Performance test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := w.SignMessage(message)
		if err != nil {
			b.Fatalf("Failed to sign message: %v", err)
		}
	}
}

func BenchmarkWalletVerification(b *testing.B) {
	w, err := wallet.NewWallet()
	if err != nil {
		b.Fatalf("Failed to create wallet: %v", err)
	}

	message := []byte("Performance test message")
	signature, err := w.SignMessage(message)
	if err != nil {
		b.Fatalf("Failed to sign message: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.VerifySignature(message, signature)
	}
}

func BenchmarkBlockchainCreation(b *testing.B) {
	wm := wallet.NewWalletManager()
	wallet1, _ := wm.CreateWallet()

	for i := 0; i < b.N; i++ {
		bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)
		if bc == nil {
			b.Fatal("Blockchain is nil")
		}
	}
}

func BenchmarkTransactionCreation(b *testing.B) {
	wm := wallet.NewWalletManager()
	wallet1, _ := wm.CreateWallet()
	wallet2, _ := wm.CreateWallet()
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 100, wallet1.GetPrivateKeyBytes())
		if err != nil {
			b.Fatalf("Failed to create transaction: %v", err)
		}
	}
}

func BenchmarkMempoolOperations(b *testing.B) {
	mempool := mining.NewMempool(1000)
	wm := wallet.NewWalletManager()
	wallet1, _ := wm.CreateWallet()
	wallet2, _ := wm.CreateWallet()
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)

	// Создаем транзакции заранее
	transactions := make([]*blockchain.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		tx, _ := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 100, wallet1.GetPrivateKeyBytes())
		transactions[i] = tx
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mempool.AddTransaction(transactions[i])
	}
}

func BenchmarkMining(b *testing.B) {
	wm := wallet.NewWalletManager()
	wallet1, _ := wm.CreateWallet()
	// Используем сложность 0 для быстрого майнинга
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 0)
	mempool := mining.NewMempool(100)

	miner := mining.NewMiner("miner_001", wallet1.GetPublicKeyBytes(), bc, mempool, nil, wallet1)

	// Запускаем майнинг
	err := miner.StartMining()
	if err != nil {
		b.Fatalf("Failed to start mining: %v", err)
	}
	defer miner.StopMining()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Проверяем статистику майнера
		stats := miner.GetStats()
		if stats == nil {
			b.Fatal("Miner stats is nil")
		}
	}
}

func TestPerformanceMetrics(t *testing.T) {
	// Тест создания кошельков
	start := time.Now()
	wm := wallet.NewWalletManager()
	for i := 0; i < 100; i++ {
		_, err := wm.CreateWallet()
		if err != nil {
			t.Fatalf("Failed to create wallet %d: %v", i, err)
		}
	}
	walletCreationTime := time.Since(start)
	t.Logf("Created 100 wallets in %v (avg: %v per wallet)", walletCreationTime, walletCreationTime/100)

	// Тест подписи сообщений
	wallet, _ := wm.CreateWallet()
	message := []byte("Performance test message")

	start = time.Now()
	for i := 0; i < 1000; i++ {
		_, err := wallet.SignMessage(message)
		if err != nil {
			t.Fatalf("Failed to sign message %d: %v", i, err)
		}
	}
	signingTime := time.Since(start)
	t.Logf("Signed 1000 messages in %v (avg: %v per signature)", signingTime, signingTime/1000)

	// Тест создания транзакций
	wallet1, _ := wm.CreateWallet()
	wallet2, _ := wm.CreateWallet()
	bc := blockchain.NewBlockchain(wallet1.GetAddress(), wallet1.GetPublicKeyBytes(), 1)

	start = time.Now()
	for i := 0; i < 100; i++ {
		_, err := bc.CreateTransaction(wallet1.GetAddress(), wallet2.GetAddress(), 100, wallet1.GetPrivateKeyBytes())
		if err != nil {
			t.Fatalf("Failed to create transaction %d: %v", i, err)
		}
	}
	transactionCreationTime := time.Since(start)
	t.Logf("Created 100 transactions in %v (avg: %v per transaction)", transactionCreationTime, transactionCreationTime/100)

	// Тест майнинга (упрощенный - только создание майнера без реального майнинга)
	mempool := mining.NewMempool(100)
	miner := mining.NewMiner("miner_001", wallet1.GetPublicKeyBytes(), bc, mempool, nil, wallet1)

	start = time.Now()

	// Проверяем, что майнер создан
	if miner == nil {
		t.Fatalf("Failed to create miner")
	}

	// Проверяем статистику майнера
	stats := miner.GetStats()
	if stats == nil {
		t.Fatalf("Miner should have stats")
	}

	miningTime := time.Since(start)
	t.Logf("Miner created and configured in %v", miningTime)
}
