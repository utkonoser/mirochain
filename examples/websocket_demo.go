//go:build websocket_demo
// +build websocket_demo

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/mining"
	"mirochain/internal/network"
	"mirochain/internal/wallet"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("🚀 MiroChain WebSocket Demo")
	fmt.Println("=============================")

	// Создаем временную директорию для данных
	dataDir := "./data/websocket_demo"
	os.MkdirAll(dataDir, 0755)

	// Создаем простой блокчейн для совместимости
	simpleBC := blockchain.NewBlockchain("test", []byte("genesis"), 2)

	// Создаем P2P сервер
	server := network.NewServer("localhost", 8080, simpleBC)

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	fmt.Printf("✅ P2P Server started on localhost:8080\n")
	fmt.Printf("✅ WebSocket server started on localhost:9080\n")
	fmt.Printf("📡 WebSocket endpoint: ws://localhost:9080/ws\n")
	fmt.Printf("📊 Status endpoint: http://localhost:9080/ws/status\n\n")

	// Создаем кошелек для демонстрации
	w, err := wallet.NewWallet()
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	// Подключаемся к WebSocket
	conn, err := connectWebSocket("ws://localhost:9080/ws?user_id=demo_user")
	if err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	fmt.Println("🔌 Connected to WebSocket server")

	// Запускаем горутину для чтения уведомлений
	go readWebSocketNotifications(conn)

	// Создаем майнер для демонстрации
	genesisBlock := simpleBC.GetLastBlock()
	miner := mining.NewOptimizedProofOfWork(genesisBlock, 2)
	mempool := mining.NewMempool(100)

	// Создаем несколько транзакций для демонстрации
	fmt.Println("📝 Creating demo transactions...")
	for i := 0; i < 3; i++ {
		tx := createDemoTransaction(w, i)
		if tx != nil {
			mempool.AddTransaction(tx)
			fmt.Printf("✅ Created transaction %d: %x\n", i+1, tx.ID)

			// Отправляем транзакцию в сеть
			server.BroadcastNewTransaction(tx)
			time.Sleep(1 * time.Second)
		}
	}

	// Майним блок с транзакциями
	fmt.Println("\n⛏️  Mining block with transactions...")
	transactions := mempool.GetTransactions()
	block := mineBlock(simpleBC, transactions, miner)

	if block != nil {
		fmt.Printf("✅ Mined block at height %d with %d transactions\n",
			block.Height, len(block.Transactions))

		// Отправляем блок в сеть
		server.BroadcastNewBlock(block)
	}

	// Демонстрируем обновление баланса
	fmt.Println("\n💰 Demonstrating balance updates...")
	address := fmt.Sprintf("%x", w.GetAddress())
	server.BroadcastBalanceUpdate(address, 1000, 500)

	// Показываем статистику
	fmt.Println("\n📊 Current Statistics:")
	fmt.Printf("   - WebSocket clients: %d\n", server.GetWebSocketClientCount())
	fmt.Printf("   - P2P peers: %d\n", server.GetPeerCount())
	fmt.Printf("   - Blockchain height: %d\n", simpleBC.GetHeight())

	// Ждем сигнал завершения
	fmt.Println("\n⏳ Press Ctrl+C to stop...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	fmt.Println("\n👋 Demo completed!")
}

// connectWebSocket подключается к WebSocket серверу
func connectWebSocket(urlStr string) (*websocket.Conn, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// readWebSocketNotifications читает уведомления из WebSocket
func readWebSocketNotifications(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		var notification map[string]interface{}
		if err := json.Unmarshal(message, &notification); err != nil {
			log.Printf("Failed to parse notification: %v", err)
			continue
		}

		// Обрабатываем уведомления
		notificationType := notification["type"].(string)
		timestamp := time.Unix(int64(notification["timestamp"].(float64)), 0)

		switch notificationType {
		case "new_block":
			data := notification["data"].(map[string]interface{})
			fmt.Printf("🔔 [%s] New Block: Height %v, Hash %v, TXs %v\n",
				timestamp.Format("15:04:05"),
				data["height"], data["hash"], data["tx_count"])

		case "new_transaction":
			data := notification["data"].(map[string]interface{})
			fmt.Printf("🔔 [%s] New Transaction: %v -> %v, Amount %v\n",
				timestamp.Format("15:04:05"),
				data["from"], data["to"], data["amount"])

		case "balance_update":
			data := notification["data"].(map[string]interface{})
			fmt.Printf("🔔 [%s] Balance Update: %v, New Balance %v, Change %v\n",
				timestamp.Format("15:04:05"),
				data["address"], data["balance"], data["change"])

		case "peer_connected":
			data := notification["data"].(map[string]interface{})
			fmt.Printf("🔔 [%s] Peer Connected: %v (%v)\n",
				timestamp.Format("15:04:05"),
				data["peer_id"], data["address"])

		case "peer_disconnected":
			data := notification["data"].(map[string]interface{})
			fmt.Printf("🔔 [%s] Peer Disconnected: %v (%v)\n",
				timestamp.Format("15:04:05"),
				data["peer_id"], data["address"])

		case "network_status":
			data := notification["data"].(map[string]interface{})
			fmt.Printf("🔔 [%s] Network Status: Peers %v, Height %v\n",
				timestamp.Format("15:04:05"),
				data["peer_count"], data["block_height"])

		default:
			fmt.Printf("🔔 [%s] Unknown notification: %s\n",
				timestamp.Format("15:04:05"), notificationType)
		}
	}
}

// createDemoTransaction создает демонстрационную транзакцию
func createDemoTransaction(w *wallet.Wallet, index int) *blockchain.Transaction {
	// Создаем простую транзакцию
	output := blockchain.TransactionOutput{
		Value:     int64(100 + index*50),
		Address:   w.GetAddress(),
		PublicKey: []byte(w.GetAddress()),
	}

	tx := &blockchain.Transaction{
		ID:        []byte(fmt.Sprintf("tx_%d_%d", index, time.Now().Unix())),
		Inputs:    []*blockchain.TransactionInput{},
		Outputs:   []*blockchain.TransactionOutput{&output},
		Timestamp: time.Now().Unix(),
	}

	return tx
}

// mineBlock майнит блок с транзакциями
func mineBlock(bc *blockchain.Blockchain, transactions []*blockchain.Transaction, miner *mining.OptimizedProofOfWork) *blockchain.Block {
	lastBlock := bc.GetLastBlock()

	block := &blockchain.Block{
		Height:       lastBlock.Height + 1,
		PreviousHash: lastBlock.Hash,
		Transactions: transactions,
		Timestamp:    time.Now().Unix(),
		Nonce:        0,
	}

	// Майним блок
	nonce, hash, found := miner.Mine()
	if !found {
		log.Printf("Failed to mine block")
		return nil
	}

	block.Nonce = int(nonce)
	block.Hash = hash

	// Добавляем блок в блокчейн
	if err := bc.AddBlock(block); err != nil {
		log.Printf("Failed to add block: %v", err)
		return nil
	}

	return block
}
