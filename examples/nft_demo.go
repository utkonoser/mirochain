//go:build nft_demo

package main

import (
	"fmt"
	"log"
	"math/big"

	"mirochain/internal/nft"
)

func main() {
	fmt.Println("=== NFT (ERC-721) Demo ===")
	fmt.Println()

	// Создаем менеджер NFT
	manager := nft.NewERC721Manager()

	// 1. Создание NFT контрактов
	fmt.Println("1. Creating NFT Contracts:")
	demonstrateContractCreation(manager)
	fmt.Println()

	// 2. Создание NFT токенов
	fmt.Println("2. Minting NFT Tokens:")
	demonstrateTokenMinting(manager)
	fmt.Println()

	// 3. Переводы NFT
	fmt.Println("3. NFT Transfers:")
	demonstrateNFTTransfers(manager)
	fmt.Println()

	// 4. Система одобрений
	fmt.Println("4. NFT Approvals:")
	demonstrateNFTApprovals(manager)
	fmt.Println()

	// 5. Поиск и статистика
	fmt.Println("5. NFT Search and Statistics:")
	demonstrateNFTSearch(manager)
	fmt.Println()

	// 6. Сжигание NFT
	fmt.Println("6. NFT Burning:")
	demonstrateNFTBurning(manager)
	fmt.Println()

	fmt.Println("NFT demo completed!")
}

func demonstrateContractCreation(manager *nft.ERC721Manager) {
	// Создаем несколько NFT контрактов
	contracts := []struct {
		name      string
		symbol    string
		owner     string
		baseURI   string
		maxSupply *big.Int
	}{
		{
			name:      "Digital Art Collection",
			symbol:    "DAC",
			owner:     "alice",
			baseURI:   "https://api.digitalart.com/metadata/",
			maxSupply: big.NewInt(1000),
		},
		{
			name:      "Game Items",
			symbol:    "GAME",
			owner:     "bob",
			baseURI:   "https://api.gameitems.com/metadata/",
			maxSupply: big.NewInt(10000),
		},
		{
			name:      "Domain Names",
			symbol:    "DOMAIN",
			owner:     "charlie",
			baseURI:   "https://api.domains.com/metadata/",
			maxSupply: nil, // Без ограничений
		},
	}

	for i, contractData := range contracts {
		contract, err := manager.CreateContract(
			contractData.name,
			contractData.symbol,
			contractData.owner,
			contractData.baseURI,
			contractData.maxSupply,
		)
		if err != nil {
			log.Printf("Error creating contract %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Contract %d created:\n", i+1)
		fmt.Printf("  Name: %s\n", contract.Name)
		fmt.Printf("  Symbol: %s\n", contract.Symbol)
		fmt.Printf("  Owner: %s\n", contract.Owner)
		fmt.Printf("  Address: %s\n", contract.Address)
		fmt.Printf("  Base URI: %s\n", contract.BaseURI)
		if contract.MaxSupply != nil {
			fmt.Printf("  Max Supply: %s\n", contract.MaxSupply.String())
		} else {
			fmt.Printf("  Max Supply: Unlimited\n")
		}
		fmt.Printf("  Total Supply: %s\n", contract.TotalSupply.String())
		fmt.Println()
	}
}

func demonstrateTokenMinting(manager *nft.ERC721Manager) {
	// Получаем первый контракт
	contractList := manager.ListContracts()
	if len(contractList) == 0 {
		fmt.Println("No contracts available for minting demo")
		return
	}

	contract := contractList[0]
	fmt.Printf("Using contract: %s (%s)\n", contract.Name, contract.Symbol)

	// Создаем несколько NFT с разными метаданными
	nfts := []struct {
		tokenID    *big.Int
		to         string
		metadata   *nft.TokenMetadata
		attributes map[string]interface{}
	}{
		{
			tokenID: big.NewInt(1),
			to:      "alice",
			metadata: &nft.TokenMetadata{
				Name:        "Sunset Over Mountains",
				Description: "A beautiful digital painting of a sunset over mountain peaks",
				Image:       "https://api.digitalart.com/images/sunset_mountains.jpg",
				ExternalURL: "https://digitalart.com/artwork/1",
			},
			attributes: map[string]interface{}{
				"artist":    "Alice Artist",
				"style":     "Digital Painting",
				"colors":    []string{"orange", "purple", "blue"},
				"rarity":    "common",
				"year":      2024,
			},
		},
		{
			tokenID: big.NewInt(2),
			to:      "bob",
			metadata: &nft.TokenMetadata{
				Name:        "Abstract Geometry",
				Description: "An abstract geometric composition with vibrant colors",
				Image:       "https://api.digitalart.com/images/abstract_geometry.jpg",
				ExternalURL: "https://digitalart.com/artwork/2",
			},
			attributes: map[string]interface{}{
				"artist":    "Bob Creator",
				"style":     "Abstract",
				"colors":    []string{"red", "yellow", "green"},
				"rarity":    "rare",
				"year":      2024,
			},
		},
		{
			tokenID: big.NewInt(3),
			to:      "charlie",
			metadata: &nft.TokenMetadata{
				Name:        "Cyberpunk City",
				Description: "A futuristic cyberpunk cityscape at night",
				Image:       "https://api.digitalart.com/images/cyberpunk_city.jpg",
				ExternalURL: "https://digitalart.com/artwork/3",
			},
			attributes: map[string]interface{}{
				"artist":    "Charlie Designer",
				"style":     "Cyberpunk",
				"colors":    []string{"neon", "purple", "pink"},
				"rarity":    "legendary",
				"year":      2024,
			},
		},
	}

	for i, nftData := range nfts {
		token, err := manager.Mint(contract.Address, nftData.to, nftData.tokenID, nftData.metadata, nftData.attributes)
		if err != nil {
			log.Printf("Error minting NFT %d: %v", i+1, err)
			continue
		}

		fmt.Printf("NFT %d minted:\n", i+1)
		fmt.Printf("  Token ID: %s\n", token.TokenID.String())
		fmt.Printf("  Owner: %s\n", token.Owner)
		fmt.Printf("  Name: %s\n", token.Metadata.Name)
		fmt.Printf("  Description: %s\n", token.Metadata.Description)
		fmt.Printf("  Image: %s\n", token.Metadata.Image)
		fmt.Printf("  Artist: %s\n", token.Attributes["artist"])
		fmt.Printf("  Style: %s\n", token.Attributes["style"])
		fmt.Printf("  Rarity: %s\n", token.Attributes["rarity"])
		fmt.Println()
	}
}

func demonstrateNFTTransfers(manager *nft.ERC721Manager) {
	// Получаем первый контракт
	contractList := manager.ListContracts()
	if len(contractList) == 0 {
		fmt.Println("No contracts available for transfer demo")
		return
	}

	contract := contractList[0]
	fmt.Printf("Using contract: %s (%s)\n", contract.Name, contract.Symbol)

	// Выполняем несколько переводов
	transfers := []struct {
		from    string
		to      string
		tokenID *big.Int
	}{
		{"alice", "bob", big.NewInt(1)},
		{"bob", "charlie", big.NewInt(2)},
		{"charlie", "dave", big.NewInt(3)},
	}

	for i, transfer := range transfers {
		event, err := manager.Transfer(contract.Address, transfer.from, transfer.to, transfer.tokenID)
		if err != nil {
			log.Printf("Transfer %d failed: %v", i+1, err)
			continue
		}

		fmt.Printf("Transfer %d: Token %s from %s to %s\n",
			i+1, event.TokenID.String(), event.From, event.To)
	}

	// Показываем владельцев после переводов
	fmt.Println("\nOwners after transfers:")
	for _, transfer := range transfers {
		owner, _ := manager.OwnerOf(contract.Address, transfer.tokenID)
		fmt.Printf("  Token %s: %s\n", transfer.tokenID.String(), owner)
	}
}

func demonstrateNFTApprovals(manager *nft.ERC721Manager) {
	// Получаем первый контракт
	contractList := manager.ListContracts()
	if len(contractList) == 0 {
		fmt.Println("No contracts available for approval demo")
		return
	}

	contract := contractList[0]
	fmt.Printf("Using contract: %s (%s)\n", contract.Name, contract.Symbol)

	// alice одобряет bob управлять токеном 1
	approvalEvent, err := manager.Approve(contract.Address, "alice", "bob", big.NewInt(1))
	if err != nil {
		log.Printf("Approval failed: %v", err)
		return
	}

	fmt.Printf("Approval: %s approved %s to manage token %s\n",
		approvalEvent.Owner, approvalEvent.Approved, approvalEvent.TokenID.String())

	// Проверяем одобрение
	approved, err := manager.GetApproved(contract.Address, big.NewInt(1))
	if err != nil {
		log.Printf("Failed to get approval: %v", err)
		return
	}

	fmt.Printf("Current approval for token 1: %s\n", approved)

	// alice одобряет charlie управлять всеми своими токенами
	approvalForAllEvent, err := manager.SetApprovalForAll(contract.Address, "alice", "charlie", true)
	if err != nil {
		log.Printf("SetApprovalForAll failed: %v", err)
		return
	}

	fmt.Printf("ApprovalForAll: %s approved %s for all tokens: %t\n",
		approvalForAllEvent.Owner, approvalForAllEvent.Operator, approvalForAllEvent.Approved)

	// Проверяем одобрение для всех токенов
	isApproved, err := manager.IsApprovedForAll(contract.Address, "alice", "charlie")
	if err != nil {
		log.Printf("Failed to check approval for all: %v", err)
		return
	}

	fmt.Printf("Is charlie approved for all alice's tokens: %t\n", isApproved)
}

func demonstrateNFTSearch(manager *nft.ERC721Manager) {
	contractList := manager.ListContracts()
	if len(contractList) == 0 {
		fmt.Println("No contracts available for search demo")
		return
	}

	// Показываем информацию о всех контрактах
	for i, contract := range contractList {
		fmt.Printf("Contract %d: %s (%s)\n", i+1, contract.Name, contract.Symbol)
		
		info, err := manager.GetContractInfo(contract.Address)
		if err != nil {
			log.Printf("Failed to get contract info: %v", err)
			continue
		}

		fmt.Printf("  Address: %s\n", info["address"])
		fmt.Printf("  Total Supply: %s\n", info["total_supply"])
		fmt.Printf("  Unique Owners: %d\n", info["unique_owners"])
		fmt.Printf("  Token Count: %d\n", info["token_count"])
		fmt.Println()

		// Показываем статистику
		stats, err := manager.GetContractStats(contract.Address)
		if err != nil {
			log.Printf("Failed to get contract stats: %v", err)
			continue
		}

		fmt.Printf("  Statistics:\n")
		fmt.Printf("    Total Owners: %d\n", stats["total_owners"])
		fmt.Printf("    Total Transfers: %d\n", stats["total_transfers"])
		fmt.Printf("    Max Tokens Owner: %s\n", stats["max_tokens_owner"])
		fmt.Printf("    Max Tokens Count: %d\n", stats["max_tokens_count"])
		fmt.Printf("    Average Tokens: %.2f\n", stats["average_tokens"])
		fmt.Println()
	}

	// Демонстрируем поиск
	fmt.Println("Searching for NFTs with rarity 'rare':")
	criteria := map[string]interface{}{
		"attributes": map[string]interface{}{
			"rarity": "rare",
		},
	}
	results, err := manager.SearchTokens(criteria)
	if err != nil {
		log.Printf("Search failed: %v", err)
		return
	}
	
	fmt.Printf("Found %d NFTs\n", len(results))
	for _, token := range results {
		fmt.Printf("  - %s (Token ID: %s, Owner: %s)\n", 
			token.Metadata.Name, token.TokenID.String(), token.Owner)
	}
}

func demonstrateNFTBurning(manager *nft.ERC721Manager) {
	// Получаем первый контракт
	contractList := manager.ListContracts()
	if len(contractList) == 0 {
		fmt.Println("No contracts available for burning demo")
		return
	}

	contract := contractList[0]
	fmt.Printf("Using contract: %s (%s)\n", contract.Name, contract.Symbol)

	// Показываем начальное состояние
	initialSupply, _ := manager.GetContractInfo(contract.Address)
	fmt.Printf("Initial total supply: %s\n", initialSupply["total_supply"])

	// Сжигаем токен 1
	err := manager.Burn(contract.Address, "alice", big.NewInt(1))
	if err != nil {
		log.Printf("Burning failed: %v", err)
		return
	}

	fmt.Println("Burned token 1")

	// Показываем состояние после сжигания
	finalSupply, _ := manager.GetContractInfo(contract.Address)
	fmt.Printf("Final total supply: %s\n", finalSupply["total_supply"])

	// Проверяем, что токен больше не существует
	_, err = manager.GetToken(contract.Address, big.NewInt(1))
	if err != nil {
		fmt.Println("Token 1 no longer exists (successfully burned)")
	} else {
		fmt.Println("Token 1 still exists (burning failed)")
	}
}
