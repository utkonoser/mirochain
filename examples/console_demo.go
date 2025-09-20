//go:build console_demo
// +build console_demo

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/config"
	"mirochain/internal/logging"
	"mirochain/internal/network"
	"mirochain/internal/persistent"
)

func main() {
	fmt.Println("=== MiroChain Console Demo ===")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логирование
	logger, err := logging.CreateFileLogger("logs/console.log", logging.LevelInfo, logging.FormatJSON)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Инициализируем трассировщик
	logging.InitTracer(logger)

	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain(
		cfg.Node.DataDir,
		"console_demo",
		[]byte("console_public_key"),
		cfg.Blockchain.Difficulty,
	)
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}
	defer bc.Close()

	// Создаем P2P сервер
	server := network.NewServer(cfg.Node.Address, cfg.Node.Port, &blockchain.Blockchain{})

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Демонстрируем различные функции
	fmt.Println("\n1. Testing Configuration System:")
	fmt.Printf("   Node Port: %d\n", cfg.Node.Port)
	fmt.Printf("   WebSocket Port: %d\n", cfg.Network.WebSocketPort)
	fmt.Printf("   DHT Port: %d\n", cfg.Network.DHTPort)
	fmt.Printf("   Mining Enabled: %t\n", cfg.Mining.Enabled)
	fmt.Printf("   Log Level: %s\n", cfg.Logging.Level)

	fmt.Println("\n2. Testing Logging System:")
	logger.Info("demo", "This is an info message", map[string]interface{}{
		"component": "demo",
		"test":      true,
	})
	logger.Warn("demo", "This is a warning message", map[string]interface{}{
		"warning_type": "test",
	})
	logger.Error("demo", "This is an error message", map[string]interface{}{
		"error_code": "TEST_ERROR",
	})

	fmt.Println("\n3. Testing Tracing System:")
	// Трассируем создание транзакции
	txSpan := logging.TraceTransaction("demo_tx_001", "create_transaction")
	logging.AddTag(txSpan, "amount", "100")
	logging.AddLog(txSpan, "Creating demo transaction", map[string]interface{}{
		"from": "alice",
		"to":   "bob",
	})
	time.Sleep(100 * time.Millisecond)
	logging.FinishSpan(txSpan, nil)

	// Трассируем майнинг блока
	miningSpan := logging.TraceMining(1, "mine_block")
	logging.AddTag(miningSpan, "difficulty", "4")
	logging.AddLog(miningSpan, "Starting block mining", map[string]interface{}{
		"block_height": 1,
	})
	time.Sleep(200 * time.Millisecond)
	logging.FinishSpan(miningSpan, nil)

	// Трассируем сетевую операцию
	networkSpan := logging.TraceNetwork("peer_001", "send_message")
	logging.AddTag(networkSpan, "message_type", "block")
	logging.AddLog(networkSpan, "Sending block to peer", map[string]interface{}{
		"peer_address": "localhost:8081",
	})
	time.Sleep(50 * time.Millisecond)
	logging.FinishSpan(networkSpan, nil)

	fmt.Println("\n4. Testing Network Statistics:")
	fmt.Printf("   Connected Peers: %d\n", len(server.Peers))
	fmt.Printf("   WebSocket Clients: %d\n", server.GetWebSocketClientCount())

	// DHT статистика
	dhtStats := server.GetDHTStats()
	fmt.Printf("   DHT Peers: %v\n", dhtStats["peer_count"])

	// Gossip статистика
	gossipStats := server.GetGossipStats()
	fmt.Printf("   Gossip Nodes: %v\n", gossipStats["total_nodes"])

	// Rate Limiter статистика
	rateLimiterStats := server.GetRateLimiterStats()
	fmt.Printf("   Rate Limiters: %d\n", len(rateLimiterStats))

	// NAT статистика
	natStats := server.GetNATStats()
	fmt.Printf("   NAT Type: %s\n", natStats["nat_type"])

	fmt.Println("\n5. Testing Blockchain Operations:")
	height, _ := bc.GetHeight()
	fmt.Printf("   Blockchain Height: %d\n", height)

	// Создаем тестовую транзакцию
	tx := &blockchain.Transaction{
		ID: []byte("demo_tx_002"),
		Inputs: []*blockchain.TransactionInput{
			{
				TransactionID: []byte("genesis"),
				OutputIndex:   0,
				Signature:     []byte("demo_signature"),
				PublicKey:     []byte("demo_public_key"),
			},
		},
		Outputs: []*blockchain.TransactionOutput{
			{
				Address:   "demo_address",
				Value:     50,
				PublicKey: []byte("demo_public_key"),
			},
		},
	}

	// Распространяем транзакцию через Gossip
	if err := server.BroadcastTransactionGossip(tx); err != nil {
		fmt.Printf("   Failed to broadcast transaction: %v\n", err)
	} else {
		fmt.Println("   Transaction broadcasted successfully")
	}

	fmt.Println("\n6. Testing Rate Limiting:")
	// Тестируем rate limiting
	allowed := 0
	blocked := 0
	for i := 0; i < 10; i++ {
		if server.CheckRateLimit("api", "demo_client") {
			allowed++
		} else {
			blocked++
		}
	}
	fmt.Printf("   API Rate Limiter: Allowed=%d, Blocked=%d\n", allowed, blocked)

	fmt.Println("\n7. Exporting Trace Data:")
	// Экспортируем данные трассировки
	traceData, err := logging.GetTracer().ExportSpans()
	if err != nil {
		fmt.Printf("   Failed to export traces: %v\n", err)
	} else {
		fmt.Printf("   Exported %d bytes of trace data\n", len(traceData))
	}

	fmt.Println("\n8. Testing Configuration Validation:")
	if err := cfg.Validate(); err != nil {
		fmt.Printf("   Configuration validation failed: %v\n", err)
	} else {
		fmt.Println("   Configuration is valid")
	}

	fmt.Println("\n9. Testing Configuration Save:")
	if err := cfg.Save("config_demo.yaml"); err != nil {
		fmt.Printf("   Failed to save config: %v\n", err)
	} else {
		fmt.Println("   Configuration saved to config_demo.yaml")
	}

	fmt.Println("\nConsole Demo completed!")
	fmt.Println("Check logs/console.log for detailed logs")
	fmt.Println("Check config_demo.yaml for configuration example")
}
