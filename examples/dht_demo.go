//go:build dht_demo
// +build dht_demo

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
)

func main() {
	fmt.Println("🌐 MiroChain DHT Demo")
	fmt.Println("=====================")

	// Создаем блокчейн
	bc := blockchain.NewBlockchain("test", []byte("genesis"), 2)

	// Создаем P2P сервер с DHT
	server := network.NewServer("localhost", 8080, bc)

	// Добавляем bootstrap узлы (в реальности это были бы известные узлы сети)
	server.AddBootstrapNode("127.0.0.1:8081")
	server.AddBootstrapNode("127.0.0.1:8082")

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	fmt.Printf("✅ P2P Server started on localhost:8080\n")
	fmt.Printf("✅ WebSocket server started on localhost:9080\n")
	fmt.Printf("✅ DHT server started on localhost:10080\n")

	// Показываем DHT статистику
	fmt.Println("\n📊 DHT Statistics:")
	stats := server.GetDHTStats()
	fmt.Printf("   - DHT Node ID: %s\n", stats["node_id"])
	fmt.Printf("   - Peer Count: %d\n", stats["peer_count"])
	fmt.Printf("   - Bootstrap Nodes: %v\n", stats["bootstrap"])

	// Ждем немного для инициализации
	time.Sleep(2 * time.Second)

	// Показываем обновленную статистику
	fmt.Println("\n📊 Updated DHT Statistics:")
	stats = server.GetDHTStats()
	fmt.Printf("   - DHT Node ID: %s\n", stats["node_id"])
	fmt.Printf("   - Peer Count: %d\n", stats["peer_count"])

	// Получаем список peer'ов из DHT
	fmt.Println("\n🔍 DHT Peers:")
	dhtPeers := server.GetDHTPeers()
	if len(dhtPeers) == 0 {
		fmt.Println("   No peers found in DHT")
	} else {
		for i, peer := range dhtPeers {
			fmt.Printf("   %d. ID: %s, Address: %s:%d\n",
				i+1, peer.ID, peer.Address, peer.Port)
		}
	}

	// Демонстрируем поиск peer'ов
	fmt.Println("\n🔍 Discovering peers...")
	if err := server.DiscoverPeers(); err != nil {
		fmt.Printf("   Discovery failed: %v\n", err)
	} else {
		fmt.Println("   Discovery completed")
	}

	// Показываем финальную статистику
	fmt.Println("\n📊 Final Statistics:")
	fmt.Printf("   - P2P Peers: %d\n", server.GetPeerCount())
	fmt.Printf("   - WebSocket Clients: %d\n", server.GetWebSocketClientCount())
	fmt.Printf("   - DHT Peers: %d\n", len(server.GetDHTPeers()))

	fmt.Println("\n✅ DHT Demo completed!")
	fmt.Println("💡 DHT provides decentralized peer discovery for scalable networks")
}
