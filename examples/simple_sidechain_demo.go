//go:build simple_sidechain_demo

package main

import (
	"fmt"
	"log"
	"math/big"

	"mirochain/internal/sidechain"
)

func main() {
	fmt.Println("=== Simple Sidechain Demo ===")
	fmt.Println()

	// Создаем менеджер sidechains
	manager := sidechain.NewSidechainManager()

	// 1. Создание sidechain
	fmt.Println("1. Creating Sidechain:")
	config := sidechain.SidechainConfig{
		ConsensusAlgorithm: "PoS",
		BlockTime:         2,
		Difficulty:        3,
		MaxBlockSize:      1024 * 1024, // 1MB
		GasLimit:          1000000,
		ValidatorCount:    5,
		BridgeEnabled:     true,
		CrossChainEnabled: true,
	}

	sidechain, err := manager.CreateSidechain("Test Chain", "A test sidechain", "alice", "main", config)
	if err != nil {
		log.Fatalf("Error creating sidechain: %v", err)
	}

	fmt.Printf("Sidechain created:\n")
	fmt.Printf("  ID: %s\n", sidechain.ID)
	fmt.Printf("  Name: %s\n", sidechain.Name)
	fmt.Printf("  Creator: %s\n", sidechain.Creator)
	fmt.Printf("  Consensus: %s\n", sidechain.Config.ConsensusAlgorithm)
	fmt.Printf("  Block Time: %d seconds\n", sidechain.Config.BlockTime)
	fmt.Printf("  Validators: %d\n", sidechain.Config.ValidatorCount)
	fmt.Println()

	// 2. Создание актива
	fmt.Println("2. Creating Asset:")
	asset, err := manager.CreateAsset(sidechain.ID, "Test Token", "TEST", 18, big.NewInt(1000000), "alice", sidechain.AssetTypeToken)
	if err != nil {
		log.Fatalf("Error creating asset: %v", err)
	}

	fmt.Printf("Asset created:\n")
	fmt.Printf("  ID: %s\n", asset.ID)
	fmt.Printf("  Name: %s\n", asset.Name)
	fmt.Printf("  Symbol: %s\n", asset.Symbol)
	fmt.Printf("  Total Supply: %s\n", asset.TotalSupply.String())
	fmt.Printf("  Type: %s\n", asset.Type)
	fmt.Println()

	// 3. Создание мостовой транзакции
	fmt.Println("3. Creating Bridge Transaction:")
	bridgeTx, err := manager.CreateBridgeTransaction(sidechain.ID, "main", "native", big.NewInt(1000), "alice", "bob")
	if err != nil {
		log.Fatalf("Error creating bridge transaction: %v", err)
	}

	fmt.Printf("Bridge Transaction created:\n")
	fmt.Printf("  ID: %s\n", bridgeTx.ID)
	fmt.Printf("  Source Chain: %s\n", bridgeTx.SourceChain)
	fmt.Printf("  Target Chain: %s\n", bridgeTx.TargetChain)
	fmt.Printf("  Asset: %s\n", bridgeTx.Asset)
	fmt.Printf("  Amount: %s\n", bridgeTx.Amount.String())
	fmt.Printf("  Status: %s\n", bridgeTx.Status)
	fmt.Println()

	// 4. Отправка кросс-чейн сообщения
	fmt.Println("4. Sending Cross-Chain Message:")
	message, err := manager.SendCrossChainMessage(sidechain.ID, "main", "asset_transfer", map[string]interface{}{
		"asset":  "native",
		"amount": "1000",
		"from":   "alice",
		"to":     "bob",
	})
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

	fmt.Printf("Cross-Chain Message sent:\n")
	fmt.Printf("  ID: %s\n", message.ID)
	fmt.Printf("  Source Chain: %s\n", message.SourceChain)
	fmt.Printf("  Target Chain: %s\n", message.TargetChain)
	fmt.Printf("  Type: %s\n", message.Type)
	fmt.Printf("  Status: %s\n", message.Status)
	fmt.Println()

	// 5. Получение статистики
	fmt.Println("5. Getting Sidechain Stats:")
	stats, err := manager.GetSidechainStats(sidechain.ID)
	if err != nil {
		log.Fatalf("Error getting stats: %v", err)
	}

	fmt.Printf("Sidechain Stats:\n")
	fmt.Printf("  ID: %s\n", stats["id"])
	fmt.Printf("  Name: %s\n", stats["name"])
	fmt.Printf("  Status: %s\n", stats["status"])
	fmt.Printf("  Height: %d\n", stats["height"])
	fmt.Printf("  Total Blocks: %d\n", stats["total_blocks"])
	fmt.Printf("  Total Transactions: %d\n", stats["total_transactions"])
	fmt.Printf("  Total Assets: %d\n", stats["total_assets"])
	fmt.Printf("  Active Validators: %d\n", stats["active_validators"])
	fmt.Printf("  Bridge Transactions: %d\n", stats["bridge_transactions"])
	fmt.Println()

	fmt.Println("Simple sidechain demo completed!")
}
