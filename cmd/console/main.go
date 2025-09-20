package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/persistent"
)

// Console представляет интерактивную консоль
type Console struct {
	blockchain *persistent.CachedPersistentBlockchain
	server     *network.Server
	scanner    *bufio.Scanner
	running    bool
}

// NewConsole создает новую консоль
func NewConsole() (*Console, error) {
	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain("data/console", "console_address", []byte("console_public_key"), 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create blockchain: %v", err)
	}

	// Создаем P2P сервер
	server := network.NewServer("localhost", 8080, &blockchain.Blockchain{})

	return &Console{
		blockchain: bc,
		server:     server,
		scanner:    bufio.NewScanner(os.Stdin),
		running:    true,
	}, nil
}

// Start запускает консоль
func (c *Console) Start() error {
	fmt.Println("🚀 MiroChain Interactive Console")
	fmt.Println("=================================")
	fmt.Println("Type 'help' for available commands")
	fmt.Println()

	// Запускаем сервер в фоне
	if err := c.server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	// Основной цикл консоли
	for c.running {
		fmt.Print("mirochain> ")
		if !c.scanner.Scan() {
			break
		}

		line := strings.TrimSpace(c.scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := parts[0]
		args := parts[1:]

		if err := c.executeCommand(command, args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

// executeCommand выполняет команду
func (c *Console) executeCommand(command string, args []string) error {
	switch command {
	case "help", "h":
		c.showHelp()
	case "status", "s":
		c.showStatus()
	case "blockchain", "bc":
		c.showBlockchainInfo()
	case "peers", "p":
		c.showPeers()
	case "network", "n":
		c.showNetworkInfo()
	case "mining", "m":
		c.toggleMining(args)
	case "wallet", "w":
		c.walletCommands(args)
	case "transaction", "tx":
		c.transactionCommands(args)
	case "config", "c":
		c.configCommands(args)
	case "logs", "l":
		c.logsCommands(args)
	case "quit", "q", "exit":
		c.quit()
	default:
		fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", command)
	}
	return nil
}

// showHelp показывает справку
func (c *Console) showHelp() {
	fmt.Println("\n📚 Available Commands:")
	fmt.Println("=====================")
	fmt.Println("General:")
	fmt.Println("  help, h              - Show this help")
	fmt.Println("  status, s            - Show node status")
	fmt.Println("  quit, q, exit        - Exit console")
	fmt.Println()
	fmt.Println("Blockchain:")
	fmt.Println("  blockchain, bc       - Show blockchain info")
	fmt.Println("  peers, p             - Show connected peers")
	fmt.Println("  network, n           - Show network statistics")
	fmt.Println()
	fmt.Println("Mining:")
	fmt.Println("  mining, m [on|off]   - Toggle mining")
	fmt.Println()
	fmt.Println("Wallet:")
	fmt.Println("  wallet, w list       - List wallets")
	fmt.Println("  wallet, w create     - Create new wallet")
	fmt.Println("  wallet, w balance <addr> - Show balance")
	fmt.Println()
	fmt.Println("Transactions:")
	fmt.Println("  transaction, tx send <from> <to> <amount> - Send transaction")
	fmt.Println("  transaction, tx list - List recent transactions")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  config, c show       - Show current config")
	fmt.Println("  config, c reload     - Reload configuration")
	fmt.Println()
	fmt.Println("Logs:")
	fmt.Println("  logs, l level <level> - Set log level")
	fmt.Println("  logs, l tail         - Show recent logs")
	fmt.Println()
}

// showStatus показывает статус узла
func (c *Console) showStatus() {
	fmt.Println("\n📊 Node Status:")
	fmt.Println("===============")
	fmt.Printf("Node ID: %s\n", c.server.ID)
	fmt.Printf("Address: %s:%d\n", c.server.Address, c.server.Port)
	fmt.Printf("Running: %t\n", c.server.IsRunning)

	// Blockchain info
	height, _ := c.blockchain.GetHeight()
	fmt.Printf("Blockchain Height: %d\n", height)

	// Network info
	fmt.Printf("Connected Peers: %d\n", len(c.server.Peers))
	fmt.Printf("WebSocket Clients: %d\n", c.server.GetWebSocketClientCount())

	// DHT info
	dhtStats := c.server.GetDHTStats()
	fmt.Printf("DHT Peers: %v\n", dhtStats["peer_count"])

	// Gossip info
	gossipStats := c.server.GetGossipStats()
	fmt.Printf("Gossip Nodes: %v\n", gossipStats["total_nodes"])

	// NAT info
	natStats := c.server.GetNATStats()
	fmt.Printf("NAT Type: %s\n", natStats["nat_type"])
	fmt.Printf("Behind NAT: %v\n", natStats["is_behind_nat"])
	fmt.Println()
}

// showBlockchainInfo показывает информацию о блокчейне
func (c *Console) showBlockchainInfo() {
	fmt.Println("\n⛓️  Blockchain Information:")
	fmt.Println("==========================")

	height, _ := c.blockchain.GetHeight()
	fmt.Printf("Height: %d\n", height)

	// Получаем последний блок
	lastBlock, err := c.blockchain.GetBlockByHeight(height - 1)
	if err != nil {
		fmt.Printf("Error getting last block: %v\n", err)
		return
	}

	fmt.Printf("Last Block Hash: %x\n", lastBlock.Hash)
	fmt.Printf("Last Block Timestamp: %s\n", time.Unix(lastBlock.Timestamp, 0).Format(time.RFC3339))
	fmt.Printf("Last Block Transactions: %d\n", len(lastBlock.Transactions))
	fmt.Printf("Last Block Nonce: %d\n", lastBlock.Nonce)
	fmt.Printf("Last Block Difficulty: %d\n", lastBlock.Difficulty)
	fmt.Println()
}

// showPeers показывает подключенных peer'ов
func (c *Console) showPeers() {
	fmt.Println("\n👥 Connected Peers:")
	fmt.Println("==================")

	if len(c.server.Peers) == 0 {
		fmt.Println("No peers connected")
		return
	}

	for id, peer := range c.server.Peers {
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("  Address: %s\n", peer.Address)
		fmt.Printf("  Connected: %t\n", peer.IsConnected)
		fmt.Printf("  Last Seen: %s\n", peer.LastSeen.Format(time.RFC3339))
		fmt.Println()
	}
}

// showNetworkInfo показывает сетевую статистику
func (c *Console) showNetworkInfo() {
	fmt.Println("\n🌐 Network Statistics:")
	fmt.Println("=====================")

	// P2P Peers
	fmt.Printf("P2P Peers: %d\n", len(c.server.Peers))

	// WebSocket
	fmt.Printf("WebSocket Clients: %d\n", c.server.GetWebSocketClientCount())

	// DHT
	dhtStats := c.server.GetDHTStats()
	fmt.Printf("DHT Peers: %v\n", dhtStats["peer_count"])

	// Gossip
	gossipStats := c.server.GetGossipStats()
	fmt.Printf("Gossip Total Nodes: %v\n", gossipStats["total_nodes"])
	fmt.Printf("Gossip Active Nodes: %v\n", gossipStats["active_nodes"])
	fmt.Printf("Gossip Average Score: %.2f\n", gossipStats["average_score"])

	// Rate Limiter
	rateLimiterStats := c.server.GetRateLimiterStats()
	fmt.Printf("Rate Limiters: %d\n", len(rateLimiterStats))

	// NAT
	natStats := c.server.GetNATStats()
	fmt.Printf("NAT Type: %s\n", natStats["nat_type"])
	fmt.Printf("External IP: %s\n", natStats["external_ip"])
	fmt.Printf("Reachable Peers: %v\n", natStats["reachable_peers"])
	fmt.Println()
}

// toggleMining переключает майнинг
func (c *Console) toggleMining(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: mining [on|off]")
		return
	}

	switch args[0] {
	case "on":
		fmt.Println("Mining started (if supported)")
		// Здесь должна быть логика запуска майнинга
	case "off":
		fmt.Println("Mining stopped (if supported)")
		// Здесь должна быть логика остановки майнинга
	default:
		fmt.Println("Usage: mining [on|off]")
	}
}

// walletCommands обрабатывает команды кошелька
func (c *Console) walletCommands(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: wallet [list|create|balance <addr>]")
		return
	}

	switch args[0] {
	case "list":
		fmt.Println("Wallet list functionality not implemented yet")
	case "create":
		fmt.Println("Wallet creation functionality not implemented yet")
	case "balance":
		if len(args) < 2 {
			fmt.Println("Usage: wallet balance <address>")
			return
		}
		fmt.Printf("Balance for %s: Not implemented yet\n", args[1])
	default:
		fmt.Println("Usage: wallet [list|create|balance <addr>]")
	}
}

// transactionCommands обрабатывает команды транзакций
func (c *Console) transactionCommands(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: transaction [send <from> <to> <amount>|list]")
		return
	}

	switch args[0] {
	case "send":
		if len(args) < 4 {
			fmt.Println("Usage: transaction send <from> <to> <amount>")
			return
		}
		amount, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			fmt.Printf("Invalid amount: %v\n", err)
			return
		}
		fmt.Printf("Sending %d from %s to %s (not implemented yet)\n", amount, args[1], args[2])
	case "list":
		fmt.Println("Transaction list functionality not implemented yet")
	default:
		fmt.Println("Usage: transaction [send <from> <to> <amount>|list]")
	}
}

// configCommands обрабатывает команды конфигурации
func (c *Console) configCommands(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: config [show|reload]")
		return
	}

	switch args[0] {
	case "show":
		fmt.Println("Current Configuration:")
		fmt.Printf("  Node ID: %s\n", c.server.ID)
		fmt.Printf("  Address: %s:%d\n", c.server.Address, c.server.Port)
		fmt.Printf("  WebSocket Port: %d\n", c.server.Port+1000)
		fmt.Printf("  DHT Port: %d\n", c.server.Port+2000)
	case "reload":
		fmt.Println("Configuration reload functionality not implemented yet")
	default:
		fmt.Println("Usage: config [show|reload]")
	}
}

// logsCommands обрабатывает команды логов
func (c *Console) logsCommands(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: logs [level <level>|tail]")
		return
	}

	switch args[0] {
	case "level":
		if len(args) < 2 {
			fmt.Println("Usage: logs level <level>")
			return
		}
		fmt.Printf("Log level set to: %s (not implemented yet)\n", args[1])
	case "tail":
		fmt.Println("Recent logs (not implemented yet)")
	default:
		fmt.Println("Usage: logs [level <level>|tail]")
	}
}

// quit выходит из консоли
func (c *Console) quit() {
	fmt.Println("Shutting down...")
	c.running = false
	c.server.Stop()
	c.blockchain.Close()
}

func main() {
	console, err := NewConsole()
	if err != nil {
		log.Fatalf("Failed to create console: %v", err)
	}
	defer console.blockchain.Close()

	if err := console.Start(); err != nil {
		log.Fatalf("Console error: %v", err)
	}
}
