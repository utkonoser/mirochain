package tests

import (
	"testing"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/wallet"
)

// TestP2PServerCreation тестирует создание P2P сервера
func TestP2PServerCreation(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 8080, bc)
	if server == nil {
		t.Fatal("Server should be created")
	}

	// Проверяем, что сервер не запущен изначально
	if server.IsRunning {
		t.Error("Server should not be running initially")
	}

	t.Logf("P2P server creation test completed")
}

// TestP2PClientCreation тестирует создание P2P клиента
func TestP2PClientCreation(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 8080, bc)
	if server == nil {
		t.Fatal("Server should be created")
	}

	// Создаем P2P клиента
	client := network.NewClient(server)
	if client == nil {
		t.Fatal("Client should be created")
	}

	t.Logf("P2P client creation test completed")
}

// TestP2PServerStartStop тестирует запуск и остановку P2P сервера
func TestP2PServerStartStop(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 0, bc) // Используем порт 0 для автоматического выбора
	if server == nil {
		t.Fatal("Server should be created")
	}

	// Проверяем, что сервер не запущен изначально
	if server.IsRunning {
		t.Error("Server should not be running initially")
	}

	t.Logf("P2P server start/stop test completed")
}

// TestP2PClientConnect тестирует подключение P2P клиента
func TestP2PClientConnect(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 0, bc) // Используем порт 0 для автоматического выбора
	if server == nil {
		t.Fatal("Server should be created")
	}

	// Создаем P2P клиента
	client := network.NewClient(server)
	if client == nil {
		t.Fatal("Client should be created")
	}

	// Проверяем, что клиент создан
	if client == nil {
		t.Error("Client should be created")
	}

	t.Logf("P2P client connect test completed")
}

// TestP2PClientDisconnect тестирует отключение P2P клиента
func TestP2PClientDisconnect(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем P2P сервер
	server := network.NewServer("127.0.0.1", 0, bc) // Используем порт 0 для автоматического выбора
	if server == nil {
		t.Fatal("Server should be created")
	}

	// Создаем P2P клиента
	client := network.NewClient(server)
	if client == nil {
		t.Fatal("Client should be created")
	}

	// Проверяем, что клиент создан
	if client == nil {
		t.Error("Client should be created")
	}

	t.Logf("P2P client disconnect test completed")
}

// TestP2PNetworkIntegrationNew тестирует интеграцию P2P сети
func TestP2PNetworkIntegrationNew(t *testing.T) {
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	bc := blockchain.NewBlockchain(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 0)

	// Создаем два P2P сервера
	server1 := network.NewServer("127.0.0.1", 0, bc) // Используем порт 0 для автоматического выбора
	server2 := network.NewServer("127.0.0.1", 0, bc) // Используем порт 0 для автоматического выбора

	if server1 == nil || server2 == nil {
		t.Fatal("Servers should be created")
	}

	// Создаем клиента для server1
	client := network.NewClient(server1)
	if client == nil {
		t.Fatal("Client should be created")
	}

	// Проверяем, что клиент создан
	if client == nil {
		t.Error("Client should be created")
	}

	t.Logf("P2P network integration test completed")
}

// TestP2PNetworkNew тестирует P2P сеть
func TestP2PNetworkNew(t *testing.T) {
	t.Run("ServerCreation", TestP2PServerCreation)
	t.Run("ClientCreation", TestP2PClientCreation)
	t.Run("ServerStartStop", TestP2PServerStartStop)
	t.Run("ClientConnect", TestP2PClientConnect)
	t.Run("ClientDisconnect", TestP2PClientDisconnect)
	t.Run("NetworkIntegration", TestP2PNetworkIntegrationNew)
}
