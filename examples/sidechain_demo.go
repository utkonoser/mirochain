//go:build sidechain_demo

package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"mirochain/internal/sidechain"
)

func main() {
	fmt.Println("=== Sidechain Demo ===")
	fmt.Println()

	// Создаем менеджер sidechains
	manager := sidechain.NewSidechainManager()

	// 1. Создание sidechains
	fmt.Println("1. Creating Sidechains:")
	demonstrateSidechainCreation(manager)
	fmt.Println()

	// 2. Создание активов
	fmt.Println("2. Creating Assets:")
	demonstrateAssetCreation(manager)
	fmt.Println()

	// 3. Добавление блоков и транзакций
	fmt.Println("3. Adding Blocks and Transactions:")
	demonstrateBlocksAndTransactions(manager)
	fmt.Println()

	// 4. Мостовые транзакции
	fmt.Println("4. Bridge Transactions:")
	demonstrateBridgeTransactions(manager)
	fmt.Println()

	// 5. Кросс-чейн сообщения
	fmt.Println("5. Cross-Chain Messages:")
	demonstrateCrossChainMessages(manager)
	fmt.Println()

	// 6. Статистика и управление
	fmt.Println("6. Statistics and Management:")
	demonstrateStatistics(manager)
	fmt.Println()

	fmt.Println("Sidechain demo completed!")
}

func demonstrateSidechainCreation(manager *sidechain.SidechainManager) {
	// Создаем несколько sidechains с разными конфигурациями
	sidechains := []struct {
		name        string
		description string
		creator     string
		parentChain string
		config      sidechain.SidechainConfig
	}{
		{
			name:        "Game Chain",
			description: "A sidechain for gaming applications",
			creator:     "alice",
			parentChain: "main",
			config: sidechain.SidechainConfig{
				ConsensusAlgorithm: "PoS",
				BlockTime:         2,
				Difficulty:        3,
				MaxBlockSize:      1024 * 1024, // 1MB
				GasLimit:          1000000,
				ValidatorCount:    5,
				BridgeEnabled:     true,
				CrossChainEnabled: true,
			},
		},
		{
			name:        "DeFi Chain",
			description: "A sidechain for DeFi applications",
			creator:     "bob",
			parentChain: "main",
			config: sidechain.SidechainConfig{
				ConsensusAlgorithm: "DPoS",
				BlockTime:         1,
				Difficulty:        2,
				MaxBlockSize:      2 * 1024 * 1024, // 2MB
				GasLimit:          2000000,
				ValidatorCount:    10,
				BridgeEnabled:     true,
				CrossChainEnabled: true,
			},
		},
		{
			name:        "NFT Chain",
			description: "A sidechain for NFT marketplace",
			creator:     "charlie",
			parentChain: "main",
			config: sidechain.SidechainConfig{
				ConsensusAlgorithm: "PoW",
				BlockTime:         5,
				Difficulty:        4,
				MaxBlockSize:      512 * 1024, // 512KB
				GasLimit:          500000,
				ValidatorCount:    3,
				BridgeEnabled:     true,
				CrossChainEnabled: false,
			},
		},
	}

	for i, scData := range sidechains {
		sc, err := manager.CreateSidechain(scData.name, scData.description, scData.creator, scData.parentChain, scData.config)
		if err != nil {
			log.Printf("Error creating sidechain %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Sidechain %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", sc.ID)
		fmt.Printf("  Name: %s\n", sc.Name)
		fmt.Printf("  Creator: %s\n", sc.Creator)
		fmt.Printf("  Consensus: %s\n", sc.Config.ConsensusAlgorithm)
		fmt.Printf("  Block Time: %d seconds\n", sc.Config.BlockTime)
		fmt.Printf("  Validators: %d\n", sc.Config.ValidatorCount)
		fmt.Printf("  Bridge Enabled: %t\n", sc.Config.BridgeEnabled)
		fmt.Printf("  Cross-Chain Enabled: %t\n", sc.Config.CrossChainEnabled)
		fmt.Println()
	}
}

func demonstrateAssetCreation(manager *sidechain.SidechainManager) {
	// Получаем первый sidechain
	sidechains := manager.ListSidechains()
	if len(sidechains) == 0 {
		fmt.Println("No sidechains available for asset creation demo")
		return
	}

	sidechain := sidechains[0]
	fmt.Printf("Using sidechain: %s (%s)\n", sidechain.Name, sidechain.ID)

	// Создаем различные типы активов
	assets := []struct {
		name        string
		symbol      string
		decimals    int
		totalSupply *big.Int
		creator     string
		assetType   sidechain.AssetType
	}{
		{
			name:        "Game Token",
			symbol:      "GAME",
			decimals:    18,
			totalSupply: big.NewInt(1000000),
			creator:     "alice",
			assetType:   sidechain.AssetTypeToken,
		},
		{
			name:        "DeFi Token",
			symbol:      "DEFI",
			decimals:    18,
			totalSupply: big.NewInt(5000000),
			creator:     "bob",
			assetType:   sidechain.AssetTypeToken,
		},
		{
			name:        "NFT Collection",
			symbol:      "NFT",
			decimals:    0,
			totalSupply: big.NewInt(10000),
			creator:     "charlie",
			assetType:   sidechain.AssetTypeNFT,
		},
		{
			name:        "Bridged Bitcoin",
			symbol:      "BTC",
			decimals:    8,
			totalSupply: big.NewInt(0),
			creator:     "bridge",
			assetType:   sidechain.AssetTypeBridged,
		},
	}

	for i, assetData := range assets {
		asset, err := manager.CreateAsset(sidechain.ID, assetData.name, assetData.symbol, assetData.decimals, assetData.totalSupply, assetData.creator, assetData.assetType)
		if err != nil {
			log.Printf("Error creating asset %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Asset %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", asset.ID)
		fmt.Printf("  Name: %s\n", asset.Name)
		fmt.Printf("  Symbol: %s\n", asset.Symbol)
		fmt.Printf("  Decimals: %d\n", asset.Decimals)
		fmt.Printf("  Total Supply: %s\n", asset.TotalSupply.String())
		fmt.Printf("  Creator: %s\n", asset.Creator)
		fmt.Printf("  Type: %s\n", asset.Type)
		fmt.Println()
	}
}

func demonstrateBlocksAndTransactions(manager *sidechain.SidechainManager) {
	// Получаем первый sidechain
	sidechains := manager.ListSidechains()
	if len(sidechains) == 0 {
		fmt.Println("No sidechains available for blocks demo")
		return
	}

	sidechain := sidechains[0]
	fmt.Printf("Using sidechain: %s (%s)\n", sidechain.Name, sidechain.ID)

	// Создаем несколько блоков с транзакциями
	for i := 1; i <= 3; i++ {
		// Создаем транзакции для блока
		transactions := []*sidechain.SidechainTransaction{
			{
				ID:        fmt.Sprintf("tx_%d_1", i),
				Type:      sidechain.TxTypeTransfer,
				From:      "alice",
				To:        "bob",
				Amount:    big.NewInt(1000),
				Asset:     "native",
				GasLimit:  21000,
				GasPrice:  big.NewInt(20),
				Nonce:     int64(i),
				Timestamp: time.Now(),
			},
			{
				ID:        fmt.Sprintf("tx_%d_2", i),
				Type:      sidechain.TxTypeAssetMint,
				From:      "alice",
				To:        "alice",
				Amount:    big.NewInt(500),
				Asset:     "asset_1",
				GasLimit:  50000,
				GasPrice:  big.NewInt(20),
				Nonce:     int64(i + 1),
				Timestamp: time.Now(),
			},
		}

		// Создаем блок
		block := &sidechain.SidechainBlock{
			Index:        int64(i),
			Timestamp:    time.Now(),
			PreviousHash: sidechain.Hash,
			Hash:         fmt.Sprintf("block_hash_%d", i),
			MerkleRoot:   fmt.Sprintf("merkle_root_%d", i),
			Nonce:        int64(i * 1000),
			Difficulty:   sidechain.Config.Difficulty,
			Transactions: transactions,
			Validator:    "alice",
			Signature:    fmt.Sprintf("signature_%d", i),
		}

		// Добавляем блок
		err := manager.AddBlock(sidechain.ID, block)
		if err != nil {
			log.Printf("Error adding block %d: %v", i, err)
			continue
		}

		fmt.Printf("Block %d added:\n", i)
		fmt.Printf("  Index: %d\n", block.Index)
		fmt.Printf("  Hash: %s\n", block.Hash)
		fmt.Printf("  Previous Hash: %s\n", block.PreviousHash)
		fmt.Printf("  Transactions: %d\n", len(block.Transactions))
		fmt.Printf("  Validator: %s\n", block.Validator)
		fmt.Printf("  Nonce: %d\n", block.Nonce)
		fmt.Println()
	}
}

func demonstrateBridgeTransactions(manager *sidechain.SidechainManager) {
	// Получаем sidechains
	sidechains := manager.ListSidechains()
	if len(sidechains) < 2 {
		fmt.Println("Need at least 2 sidechains for bridge demo")
		return
	}

	sourceChain := sidechains[0]
	targetChain := sidechains[1]

	fmt.Printf("Creating bridge transaction from %s to %s\n", sourceChain.Name, targetChain.Name)

	// Создаем мостовые транзакции
	bridgeTxs := []struct {
		asset  string
		amount *big.Int
		from   string
		to     string
	}{
		{
			asset:  "native",
			amount: big.NewInt(1000),
			from:   "alice",
			to:     "bob",
		},
		{
			asset:  "asset_1",
			amount: big.NewInt(500),
			from:   "bob",
			to:     "charlie",
		},
	}

	for i, txData := range bridgeTxs {
		bridgeTx, err := manager.CreateBridgeTransaction(sourceChain.ID, targetChain.ID, txData.asset, txData.amount, txData.from, txData.to)
		if err != nil {
			log.Printf("Error creating bridge transaction %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Bridge Transaction %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", bridgeTx.ID)
		fmt.Printf("  Source Chain: %s\n", bridgeTx.SourceChain)
		fmt.Printf("  Target Chain: %s\n", bridgeTx.TargetChain)
		fmt.Printf("  Asset: %s\n", bridgeTx.Asset)
		fmt.Printf("  Amount: %s\n", bridgeTx.Amount.String())
		fmt.Printf("  From: %s\n", bridgeTx.From)
		fmt.Printf("  To: %s\n", bridgeTx.To)
		fmt.Printf("  Status: %s\n", bridgeTx.Status)
		fmt.Println()

		// Обрабатываем транзакцию
		targetTxID := fmt.Sprintf("target_tx_%d", i+1)
		validatorProof := fmt.Sprintf("proof_%d", i+1)
		
		err = manager.ProcessBridgeTransaction(bridgeTx.ID, targetTxID, validatorProof)
		if err != nil {
			log.Printf("Error processing bridge transaction %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Bridge Transaction %d processed successfully\n", i+1)
		fmt.Printf("  Target Tx ID: %s\n", targetTxID)
		fmt.Printf("  Validator Proof: %s\n", validatorProof)
		fmt.Println()
	}
}

func demonstrateCrossChainMessages(manager *sidechain.SidechainManager) {
	// Получаем sidechains
	sidechains := manager.ListSidechains()
	if len(sidechains) < 2 {
		fmt.Println("Need at least 2 sidechains for cross-chain demo")
		return
	}

	sourceChain := sidechains[0]
	targetChain := sidechains[1]

	fmt.Printf("Sending cross-chain messages from %s to %s\n", sourceChain.Name, targetChain.Name)

	// Отправляем различные типы сообщений
	messages := []struct {
		msgType string
		data    map[string]interface{}
	}{
		{
			msgType: "asset_transfer",
			data: map[string]interface{}{
				"asset":   "native",
				"amount":  "1000",
				"from":    "alice",
				"to":      "bob",
			},
		},
		{
			msgType: "validator_update",
			data: map[string]interface{}{
				"validator": "new_validator",
				"action":    "add",
			},
		},
		{
			msgType: "consensus_update",
			data: map[string]interface{}{
				"new_consensus": "PoS",
				"block_time":    3,
			},
		},
	}

	for i, msgData := range messages {
		message, err := manager.SendCrossChainMessage(sourceChain.ID, targetChain.ID, msgData.msgType, msgData.data)
		if err != nil {
			log.Printf("Error sending message %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Cross-Chain Message %d sent:\n", i+1)
		fmt.Printf("  ID: %s\n", message.ID)
		fmt.Printf("  Source Chain: %s\n", message.SourceChain)
		fmt.Printf("  Target Chain: %s\n", message.TargetChain)
		fmt.Printf("  Type: %s\n", message.Type)
		fmt.Printf("  Status: %s\n", message.Status)
		fmt.Printf("  Data: %v\n", message.Data)
		fmt.Println()

		// Обрабатываем сообщение
		err = manager.ProcessCrossChainMessage(message.ID)
		if err != nil {
			log.Printf("Error processing message %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Cross-Chain Message %d processed successfully\n", i+1)
		fmt.Println()
	}
}

func demonstrateStatistics(manager *sidechain.SidechainManager) {
	sidechains := manager.ListSidechains()
	if len(sidechains) == 0 {
		fmt.Println("No sidechains available for statistics demo")
		return
	}

	// Показываем статистику для каждого sidechain
	for i, sc := range sidechains {
		fmt.Printf("Sidechain %d: %s (%s)\n", i+1, sc.Name, sc.ID)
		
		stats, err := manager.GetSidechainStats(sc.ID)
		if err != nil {
			log.Printf("Failed to get stats for sidechain %d: %v", i+1, err)
			continue
		}

		fmt.Printf("  Status: %s\n", stats["status"])
		fmt.Printf("  Height: %d\n", stats["height"])
		fmt.Printf("  Total Blocks: %d\n", stats["total_blocks"])
		fmt.Printf("  Total Transactions: %d\n", stats["total_transactions"])
		fmt.Printf("  Total Assets: %d\n", stats["total_assets"])
		fmt.Printf("  Active Validators: %d\n", stats["active_validators"])
		fmt.Printf("  Bridge Transactions: %d\n", stats["bridge_transactions"])
		fmt.Printf("  Consensus: %s\n", stats["consensus"])
		fmt.Printf("  Block Time: %d seconds\n", stats["block_time"])
		fmt.Println()

		// Показываем активы
		assets, err := manager.ListAssets(sc.ID)
		if err != nil {
			log.Printf("Failed to get assets for sidechain %d: %v", i+1, err)
			continue
		}

		fmt.Printf("  Assets (%d):\n", len(assets))
		for _, asset := range assets {
			fmt.Printf("    - %s (%s): %s %s\n", asset.Name, asset.Symbol, asset.TotalSupply.String(), asset.Type)
		}
		fmt.Println()
	}

	// Показываем мостовые транзакции
	bridgeTxs := manager.ListBridgeTransactions()
	fmt.Printf("Total Bridge Transactions: %d\n", len(bridgeTxs))
	for i, tx := range bridgeTxs {
		fmt.Printf("  %d. %s: %s -> %s (%s %s)\n", i+1, tx.ID, tx.SourceChain, tx.TargetChain, tx.Amount.String(), tx.Asset)
	}
}
