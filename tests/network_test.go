package tests

import (
	"testing"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/wallet"
)

func TestP2PNetwork(t *testing.T) {
	// Создаем кошельки
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 4)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 8081, bc)

	// Проверяем, что сервер создан
	if server == nil {
		t.Error("Server should be created")
	}

	// Проверяем, что сервер не запущен изначально
	if server.IsRunning {
		t.Error("Server should not be running initially")
	}

	t.Logf("P2P server created successfully")
}

func TestAPIServerOld(t *testing.T) {
	// Создаем кошельки
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем API сервер с динамическим портом
	apiServer := network.NewServer("127.0.0.1", 0, bc)

	// Тестируем только создание сервера, без запуска
	if apiServer == nil {
		t.Error("API server should be created")
	}

	// Проверяем, что сервер не запущен изначально
	if apiServer.IsRunning {
		t.Error("API server should not be running initially")
	}

	t.Logf("API server creation test completed")
}

func TestMessageTypes(t *testing.T) {
	// Тестируем создание различных типов сообщений
	handshakeData := &network.HandshakeData{
		Version:     "1.0.0",
		NodeID:      "test_node",
		BestHeight:  0,
		GenesisHash: []byte("test_genesis"),
	}

	msg := network.NewMessage(network.MessageTypeHandshake, handshakeData, "from", "to")
	if msg.Type != network.MessageTypeHandshake {
		t.Errorf("Expected message type %s, got %s", network.MessageTypeHandshake, msg.Type)
	}

	if msg.From != "from" {
		t.Errorf("Expected from 'from', got %s", msg.From)
	}

	if msg.To != "to" {
		t.Errorf("Expected to 'to', got %s", msg.To)
	}

	t.Logf("Message created successfully: %s", msg.String())
}
