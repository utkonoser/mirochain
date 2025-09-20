//go:build advanced_p2p_demo
// +build advanced_p2p_demo

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
	fmt.Println("=== Advanced P2P Network Demo ===")
	fmt.Println("Demonstrating: WebSocket, DHT, Gossip, Rate Limiting, NAT Traversal")

	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain("data/advanced_p2p_demo", "test_address", []byte("test_public_key"), 1)
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

	// Ждем инициализации всех компонентов
	fmt.Println("Initializing all P2P components...")
	time.Sleep(5 * time.Second)

	// 1. WebSocket уведомления
	fmt.Println("\n1. WebSocket Notifications:")
	fmt.Printf("WebSocket clients connected: %d\n", server.GetWebSocketClientCount())

	// Симулируем WebSocket уведомления
	server.BroadcastBalanceUpdate("test_address", 1000, 100)
	fmt.Println("Sent balance update notification")

	// 2. DHT (Distributed Hash Table)
	fmt.Println("\n2. DHT (Distributed Hash Table):")

	// Добавляем bootstrap узлы
	server.AddBootstrapNode("localhost:8081")
	server.AddBootstrapNode("localhost:8082")
	server.AddBootstrapNode("localhost:8083")
	fmt.Println("Added bootstrap nodes")

	// Получаем статистику DHT
	dhtStats := server.GetDHTStats()
	fmt.Printf("DHT Node ID: %s\n", dhtStats["node_id"])
	fmt.Printf("DHT Peer count: %v\n", dhtStats["peer_count"])

	// 3. Gossip протокол
	fmt.Println("\n3. Gossip Protocol:")

	// Добавляем узлы в Gossip сеть
	server.AddGossipNode("gossip_node_1", "localhost:8081")
	server.AddGossipNode("gossip_node_2", "localhost:8082")
	server.AddGossipNode("gossip_node_3", "localhost:8083")
	server.AddGossipNode("gossip_node_4", "localhost:8084")
	fmt.Println("Added 4 nodes to Gossip network")

	// Создаем и распространяем транзакцию через Gossip
	tx := &blockchain.Transaction{
		ID: "gossip_tx_001",
		Inputs: []blockchain.TransactionInput{
			{
				PreviousTxID: "genesis",
				OutputIndex:  0,
				Signature:    "test_signature",
			},
		},
		Outputs: []blockchain.TransactionOutput{
			{
				Address: "test_address",
				Amount:  500,
			},
		},
	}

	if err := server.BroadcastTransactionGossip(tx); err != nil {
		log.Printf("Failed to broadcast transaction via Gossip: %v", err)
	} else {
		fmt.Println("Broadcasted transaction via Gossip")
	}

	// Получаем статистику Gossip
	gossipStats := server.GetGossipStats()
	fmt.Printf("Gossip total nodes: %v\n", gossipStats["total_nodes"])
	fmt.Printf("Gossip active nodes: %v\n", gossipStats["active_nodes"])

	// 4. Rate Limiting
	fmt.Println("\n4. Rate Limiting:")

	// Тестируем API rate limiter
	clientID := "demo_client"
	allowed := 0
	blocked := 0

	for i := 0; i < 25; i++ {
		if server.CheckRateLimit("api", clientID) {
			allowed++
		} else {
			blocked++
		}
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("API Rate Limiter: Allowed=%d, Blocked=%d\n", allowed, blocked)

	// Тестируем P2P rate limiter
	allowed = 0
	blocked = 0

	for i := 0; i < 30; i++ {
		if server.CheckRateLimit("p2p", clientID) {
			allowed++
		} else {
			blocked++
		}
		time.Sleep(30 * time.Millisecond)
	}

	fmt.Printf("P2P Rate Limiter: Allowed=%d, Blocked=%d\n", allowed, blocked)

	// 5. NAT Traversal
	fmt.Println("\n5. NAT Traversal:")

	// Получаем информацию о NAT
	natStats := server.GetNATStats()
	fmt.Printf("NAT Type: %s\n", natStats["nat_type"])
	fmt.Printf("External IP: %s\n", natStats["external_ip"])
	fmt.Printf("Is behind NAT: %v\n", natStats["is_behind_nat"])

	// Добавляем peer'ов для NAT Traversal
	peers := []struct {
		ID           string
		InternalAddr string
		ExternalAddr string
		NATType      network.NATType
	}{
		{"nat_peer_1", "192.168.1.100:8081", "203.0.113.1:8081", network.NATFullCone},
		{"nat_peer_2", "192.168.1.101:8082", "203.0.113.2:8082", network.NATRestrictedCone},
		{"nat_peer_3", "10.0.0.100:8083", "10.0.0.100:8083", network.NATNone},
	}

	for _, peer := range peers {
		server.AddNATPeer(peer.ID, peer.InternalAddr, peer.ExternalAddr, peer.NATType)
		fmt.Printf("Added NAT peer %s (NAT: %s)\n", peer.ID, peer.NATType)
	}

	// Пытаемся установить соединения
	for _, peer := range peers {
		if err := server.EstablishNATConnection(peer.ID); err != nil {
			fmt.Printf("Failed to connect to %s: %v\n", peer.ID, err)
		} else {
			fmt.Printf("Successfully connected to %s\n", peer.ID)
		}
	}

	// 6. Комплексная симуляция
	fmt.Println("\n6. Complex Simulation:")
	fmt.Println("Simulating real-world P2P network activity...")

	// Создаем несколько клиентов с разной активностью
	clients := []string{"client_001", "client_002", "client_003", "client_004", "client_005"}

	// Симулируем активность клиентов
	for i, client := range clients {
		go func(c string, index int) {
			for j := 0; j < 10; j++ {
				// Проверяем rate limit
				if server.CheckRateLimit("api", c) {
					// Создаем транзакцию
					tx := &blockchain.Transaction{
						ID: fmt.Sprintf("sim_tx_%d_%d", index, j),
						Inputs: []blockchain.TransactionInput{
							{
								PreviousTxID: "genesis",
								OutputIndex:  0,
								Signature:    "sim_signature",
							},
						},
						Outputs: []blockchain.TransactionOutput{
							{
								Address: "sim_address",
								Amount:  int64(100 + j*10),
							},
						},
					}

					// Распространяем через Gossip
					server.BroadcastTransactionGossip(tx)
				}

				time.Sleep(time.Duration(200+index*50) * time.Millisecond)
			}
		}(client, i)
	}

	// Ждем завершения симуляции
	time.Sleep(15 * time.Second)

	// 7. Финальная статистика
	fmt.Println("\n7. Final Statistics:")

	// WebSocket статистика
	fmt.Printf("WebSocket clients: %d\n", server.GetWebSocketClientCount())

	// DHT статистика
	dhtStats = server.GetDHTStats()
	fmt.Printf("DHT peers: %v\n", dhtStats["peer_count"])

	// Gossip статистика
	gossipStats = server.GetGossipStats()
	fmt.Printf("Gossip active nodes: %v\n", gossipStats["active_nodes"])

	// Rate Limiter статистика
	rateLimiterStats := server.GetRateLimiterStats()
	fmt.Printf("Rate limiters: %d\n", len(rateLimiterStats))

	// NAT статистика
	natStats = server.GetNATStats()
	fmt.Printf("NAT peers: %v\n", natStats["total_peers"])
	fmt.Printf("Reachable peers: %v\n", natStats["reachable_peers"])

	// 8. Демонстрация интеграции
	fmt.Println("\n8. Integration Demo:")
	fmt.Println("Demonstrating how all components work together...")

	// Создаем блок с транзакциями
	block := &blockchain.Block{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		PreviousHash: "genesis_hash",
		Hash:         "demo_block_hash",
		Transactions: []*blockchain.Transaction{tx},
		Nonce:        54321,
	}

	// Распространяем блок через Gossip
	if err := server.BroadcastBlockGossip(block); err != nil {
		log.Printf("Failed to broadcast block via Gossip: %v", err)
	} else {
		fmt.Println("Broadcasted block via Gossip")
	}

	// Отправляем WebSocket уведомление
	server.BroadcastBalanceUpdate("demo_address", 2000, 500)
	fmt.Println("Sent WebSocket notification")

	// Проверяем rate limit для нового клиента
	newClient := "integration_client"
	if server.CheckRateLimit("api", newClient) {
		fmt.Println("New client allowed by rate limiter")
	} else {
		fmt.Println("New client blocked by rate limiter")
	}

	fmt.Println("\nAdvanced P2P Network Demo completed!")
	fmt.Println("All components are working together seamlessly!")
}
