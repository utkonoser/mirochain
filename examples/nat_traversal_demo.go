//go:build nat_traversal_demo
// +build nat_traversal_demo

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
	fmt.Println("=== NAT Traversal Demo ===")

	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain("data/nat_traversal_demo", "test_address", []byte("test_public_key"), 1)
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

	// Ждем инициализации
	time.Sleep(3 * time.Second)

	// Получаем информацию о NAT
	fmt.Println("\nNAT Information:")
	natStats := server.GetNATStats()

	for key, value := range natStats {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Добавляем тестовых peer'ов с разными типами NAT
	fmt.Println("\nAdding test peers with different NAT types:")

	peers := []struct {
		ID           string
		InternalAddr string
		ExternalAddr string
		NATType      network.NATType
	}{
		{"peer_001", "192.168.1.100:8081", "203.0.113.1:8081", network.NATFullCone},
		{"peer_002", "192.168.1.101:8082", "203.0.113.2:8082", network.NATRestrictedCone},
		{"peer_003", "192.168.1.102:8083", "203.0.113.3:8083", network.NATPortRestrictedCone},
		{"peer_004", "192.168.1.103:8084", "203.0.113.4:8084", network.NATSymmetric},
		{"peer_005", "10.0.0.100:8085", "10.0.0.100:8085", network.NATNone},
	}

	for _, peer := range peers {
		server.AddNATPeer(peer.ID, peer.InternalAddr, peer.ExternalAddr, peer.NATType)
		fmt.Printf("Added peer %s: %s -> %s (NAT: %s)\n",
			peer.ID, peer.InternalAddr, peer.ExternalAddr, peer.NATType)
	}

	// Ждем немного
	time.Sleep(2 * time.Second)

	// Пытаемся установить соединения
	fmt.Println("\nAttempting to establish connections:")

	for _, peer := range peers {
		fmt.Printf("Connecting to peer %s...\n", peer.ID)

		if err := server.EstablishNATConnection(peer.ID); err != nil {
			fmt.Printf("  Failed to connect to %s: %v\n", peer.ID, err)
		} else {
			fmt.Printf("  Successfully connected to %s\n", peer.ID)
		}

		// Небольшая задержка между попытками
		time.Sleep(1 * time.Second)
	}

	// Получаем статистику NAT Traversal
	fmt.Println("\nNAT Traversal Statistics:")
	stats := server.GetNATStats()

	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Симулируем работу с NAT Traversal
	fmt.Println("\nSimulating NAT Traversal activity:")

	// Добавляем еще несколько peer'ов
	additionalPeers := []struct {
		ID           string
		InternalAddr string
		ExternalAddr string
		NATType      network.NATType
	}{
		{"peer_006", "172.16.0.100:8086", "198.51.100.1:8086", network.NATFullCone},
		{"peer_007", "172.16.0.101:8087", "198.51.100.2:8087", network.NATSymmetric},
		{"peer_008", "10.0.1.100:8088", "10.0.1.100:8088", network.NATNone},
	}

	for _, peer := range additionalPeers {
		server.AddNATPeer(peer.ID, peer.InternalAddr, peer.ExternalAddr, peer.NATType)
		fmt.Printf("Added additional peer %s\n", peer.ID)
	}

	// Пытаемся подключиться к новым peer'ам
	fmt.Println("\nConnecting to additional peers:")
	for _, peer := range additionalPeers {
		fmt.Printf("Connecting to peer %s...\n", peer.ID)

		if err := server.EstablishNATConnection(peer.ID); err != nil {
			fmt.Printf("  Failed to connect to %s: %v\n", peer.ID, err)
		} else {
			fmt.Printf("  Successfully connected to %s\n", peer.ID)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Симулируем keep-alive
	fmt.Println("\nSimulating keep-alive activity:")
	for i := 0; i < 5; i++ {
		fmt.Printf("Keep-alive cycle %d...\n", i+1)
		time.Sleep(2 * time.Second)
	}

	// Финальная статистика
	fmt.Println("\nFinal NAT Traversal Statistics:")
	finalStats := server.GetNATStats()

	for key, value := range finalStats {
		fmt.Printf("%s: %v\n", key, value)
	}

	// Тестируем различные сценарии NAT
	fmt.Println("\nTesting different NAT scenarios:")

	scenarios := []struct {
		Name      string
		LocalNAT  network.NATType
		RemoteNAT network.NATType
		Expected  string
	}{
		{"Direct connection", network.NATNone, network.NATNone, "Direct"},
		{"Full cone to full cone", network.NATFullCone, network.NATFullCone, "Hole punching"},
		{"Restricted to restricted", network.NATRestrictedCone, network.NATRestrictedCone, "Hole punching"},
		{"Symmetric to symmetric", network.NATSymmetric, network.NATSymmetric, "Relay"},
		{"Mixed NAT types", network.NATFullCone, network.NATSymmetric, "Hole punching"},
	}

	for _, scenario := range scenarios {
		fmt.Printf("Scenario: %s (Local: %s, Remote: %s) -> %s\n",
			scenario.Name, scenario.LocalNAT, scenario.RemoteNAT, scenario.Expected)
	}

	fmt.Println("\nNAT Traversal Demo completed!")
}
