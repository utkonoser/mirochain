//go:build websocket_simple_test
// +build websocket_simple_test

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
)

func main() {
	fmt.Println("🧪 WebSocket Simple Test")
	fmt.Println("========================")

	// Создаем простой блокчейн
	bc := blockchain.NewBlockchain("test", []byte("genesis"), 2)

	// Создаем P2P сервер
	server := network.NewServer("localhost", 8080, bc)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	fmt.Printf("✅ P2P Server started on localhost:8080\n")
	fmt.Printf("✅ WebSocket server started on localhost:9080\n")
	fmt.Printf("📡 WebSocket endpoint: ws://localhost:9080/ws\n")
	fmt.Printf("📊 Status endpoint: http://localhost:9080/ws/status\n\n")

	// Ждем немного для запуска сервера
	time.Sleep(1 * time.Second)

	// Проверяем статус
	fmt.Printf("📊 Current Statistics:\n")
	fmt.Printf("   - WebSocket clients: %d\n", server.GetWebSocketClientCount())
	fmt.Printf("   - P2P peers: %d\n", server.GetPeerCount())
	fmt.Printf("   - Blockchain height: %d\n", bc.GetHeight())

	// Создаем тестовый блок
	block := &blockchain.Block{
		Height:       1,
		PreviousHash: []byte("genesis"),
		Hash:         []byte("test_block"),
		Timestamp:    time.Now().Unix(),
		Nonce:        12345,
		Transactions: []*blockchain.Transaction{},
	}

	// Отправляем уведомление о новом блоке
	fmt.Println("\n📤 Broadcasting new block...")
	server.BroadcastNewBlock(block)

	// Отправляем уведомление об обновлении баланса
	fmt.Println("📤 Broadcasting balance update...")
	server.BroadcastBalanceUpdate("test_address", 1000, 500)

	fmt.Println("\n✅ Test completed successfully!")
	fmt.Println("💡 Connect to ws://localhost:9080/ws to see real-time notifications")
}
