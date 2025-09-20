//go:build gossip_demo
// +build gossip_demo

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/persistent"
)

func main() {
	fmt.Println("=== Gossip Protocol Demo ===")

	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain("data/gossip_demo", "test_address", []byte("test_public_key"), 1)
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}
	defer bc.Close()

	// Создаем P2P сервер
	server := network.NewServer("localhost", 8080, &blockchain.Blockchain{})

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Добавляем узлы в Gossip сеть
	server.AddGossipNode("node1", "localhost:8081")
	server.AddGossipNode("node2", "localhost:8082")
	server.AddGossipNode("node3", "localhost:8083")
	server.AddGossipNode("node4", "localhost:8084")

	fmt.Println("Added 4 nodes to Gossip network")

	// Ждем инициализации
	time.Sleep(2 * time.Second)

	// Создаем тестовую транзакцию
	tx := &blockchain.Transaction{
		ID: []byte("test_tx_001"),
		Inputs: []*blockchain.TransactionInput{
			{
				TransactionID: []byte("genesis"),
				OutputIndex:   0,
				Signature:     []byte("test_signature"),
				PublicKey:     []byte("test_public_key"),
			},
		},
		Outputs: []*blockchain.TransactionOutput{
			{
				Address:   "test_address",
				Value:     100,
				PublicKey: []byte("test_public_key"),
			},
		},
	}

	// Распространяем транзакцию через Gossip
	fmt.Println("Broadcasting transaction through Gossip...")
	if err := server.BroadcastTransactionGossip(tx); err != nil {
		log.Printf("Failed to broadcast transaction: %v", err)
	}

	// Создаем тестовый блок
	block := &blockchain.Block{
		Height:       1,
		Timestamp:    time.Now().Unix(),
		PreviousHash: []byte("genesis_hash"),
		Hash:         []byte("test_block_hash"),
		Transactions: []*blockchain.Transaction{tx},
		Nonce:        12345,
		Difficulty:   1,
	}

	// Распространяем блок через Gossip
	fmt.Println("Broadcasting block through Gossip...")
	if err := server.BroadcastBlockGossip(block); err != nil {
		log.Printf("Failed to broadcast block: %v", err)
	}

	// Ждем распространения
	time.Sleep(3 * time.Second)

	// Получаем статистику Gossip
	stats := server.GetGossipStats()
	fmt.Printf("\nGossip Statistics:\n")
	fmt.Printf("Total nodes: %v\n", stats["total_nodes"])
	fmt.Printf("Active nodes: %v\n", stats["active_nodes"])
	fmt.Printf("Average score: %v\n", stats["average_score"])
	fmt.Printf("Fanout: %v\n", stats["fanout"])
	fmt.Printf("Max TTL: %v\n", stats["max_ttl"])

	// Симулируем работу в течение некоторого времени
	fmt.Println("\nSimulating Gossip network activity...")
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)

		// Создаем новую транзакцию
		newTx := &blockchain.Transaction{
			ID: []byte(fmt.Sprintf("test_tx_%03d", i+2)),
			Inputs: []*blockchain.TransactionInput{
				{
					TransactionID: []byte("genesis"),
					OutputIndex:   0,
					Signature:     []byte("test_signature"),
					PublicKey:     []byte("test_public_key"),
				},
			},
			Outputs: []*blockchain.TransactionOutput{
				{
					Address:   "test_address",
					Value:     50 + int64(i*10),
					PublicKey: []byte("test_public_key"),
				},
			},
		}

		// Распространяем через Gossip
		if err := server.BroadcastTransactionGossip(newTx); err != nil {
			log.Printf("Failed to broadcast transaction %d: %v", i+2, err)
		}

		fmt.Printf("Broadcasted transaction %d\n", i+2)
	}

	// Финальная статистика
	fmt.Println("\nFinal Gossip Statistics:")
	finalStats := server.GetGossipStats()
	for key, value := range finalStats {
		fmt.Printf("%s: %v\n", key, value)
	}

	fmt.Println("\nGossip Demo completed!")
}
