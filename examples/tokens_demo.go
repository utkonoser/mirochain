//go:build tokens_demo

package main

import (
	"fmt"
	"log"
	"math/big"

	"mirochain/internal/tokens"
)

func main() {
	fmt.Println("=== ERC-20 Tokens Demo ===")
	fmt.Println()

	// Создаем менеджер токенов
	manager := tokens.NewERC20Manager()

	// 1. Создание токенов
	fmt.Println("1. Creating Tokens:")
	demonstrateTokenCreation(manager)
	fmt.Println()

	// 2. Переводы токенов
	fmt.Println("2. Token Transfers:")
	demonstrateTokenTransfers(manager)
	fmt.Println()

	// 3. Система разрешений
	fmt.Println("3. Token Approvals:")
	demonstrateTokenApprovals(manager)
	fmt.Println()

	// 4. Создание и сжигание токенов
	fmt.Println("4. Minting and Burning:")
	demonstrateMintingAndBurning(manager)
	fmt.Println()

	// 5. Статистика и поиск
	fmt.Println("5. Token Statistics and Search:")
	demonstrateTokenStats(manager)
	fmt.Println()

	// 6. Экспорт и импорт
	fmt.Println("6. Token Export/Import:")
	demonstrateTokenExportImport(manager)
	fmt.Println()

	fmt.Println("ERC-20 Tokens demo completed!")
}

func demonstrateTokenCreation(manager *tokens.ERC20Manager) {
	// Создаем несколько токенов
	tokens := []struct {
		name        string
		symbol      string
		decimals    uint8
		totalSupply *big.Int
		owner       string
	}{
		{
			name:        "MiroCoin",
			symbol:      "MIRO",
			decimals:    18,
			totalSupply: big.NewInt(1000000), // 1M токенов
			owner:       "alice",
		},
		{
			name:        "TestToken",
			symbol:      "TEST",
			decimals:    6,
			totalSupply: big.NewInt(500000), // 500K токенов
			owner:       "bob",
		},
		{
			name:        "UtilityToken",
			symbol:      "UTIL",
			decimals:    8,
			totalSupply: big.NewInt(10000000), // 10M токенов
			owner:       "charlie",
		},
	}

	for i, tokenData := range tokens {
		token, err := manager.CreateToken(
			tokenData.name,
			tokenData.symbol,
			tokenData.decimals,
			tokenData.totalSupply,
			tokenData.owner,
		)
		if err != nil {
			log.Printf("Error creating token %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Token %d created:\n", i+1)
		fmt.Printf("  Name: %s\n", token.Name)
		fmt.Printf("  Symbol: %s\n", token.Symbol)
		fmt.Printf("  Decimals: %d\n", token.Decimals)
		fmt.Printf("  Total Supply: %s\n", token.TotalSupply.String())
		fmt.Printf("  Owner: %s\n", token.Owner)
		fmt.Printf("  Address: %s\n", token.Address)
		fmt.Printf("  Owner Balance: %s\n", token.Balances[token.Owner].String())
		fmt.Println()
	}
}

func demonstrateTokenTransfers(manager *tokens.ERC20Manager) {
	// Получаем первый токен
	tokenList := manager.ListTokens()
	if len(tokenList) == 0 {
		fmt.Println("No tokens available for transfer demo")
		return
	}

	token := tokenList[0]
	fmt.Printf("Using token: %s (%s)\n", token.Name, token.Symbol)

	// Выполняем несколько переводов
	transfers := []struct {
		from   string
		to     string
		amount *big.Int
	}{
		{token.Owner, "alice", big.NewInt(1000)},
		{token.Owner, "bob", big.NewInt(2000)},
		{"alice", "charlie", big.NewInt(500)},
		{"bob", "dave", big.NewInt(1000)},
	}

	for i, transfer := range transfers {
		event, err := manager.Transfer(token.Address, transfer.from, transfer.to, transfer.amount)
		if err != nil {
			log.Printf("Transfer %d failed: %v", i+1, err)
			continue
		}

		fmt.Printf("Transfer %d: %s -> %s (%s tokens)\n",
			i+1, event.From, event.To, event.Value.String())
	}

	// Показываем финальные балансы
	fmt.Println("\nFinal balances:")
	addresses := []string{token.Owner, "alice", "bob", "charlie", "dave"}
	for _, addr := range addresses {
		balance, _ := manager.BalanceOf(token.Address, addr)
		fmt.Printf("  %s: %s\n", addr, balance.String())
	}
}

func demonstrateTokenApprovals(manager *tokens.ERC20Manager) {
	// Получаем первый токен
	tokenList := manager.ListTokens()
	if len(tokenList) == 0 {
		fmt.Println("No tokens available for approval demo")
		return
	}

	token := tokenList[0]
	fmt.Printf("Using token: %s (%s)\n", token.Name, token.Symbol)

	// alice одобряет bob тратить 1000 токенов
	approvalEvent, err := manager.Approve(token.Address, "alice", "bob", big.NewInt(1000))
	if err != nil {
		log.Printf("Approval failed: %v", err)
		return
	}

	fmt.Printf("Approval: %s approved %s to spend %s tokens\n",
		approvalEvent.Owner, approvalEvent.Spender, approvalEvent.Value.String())

	// Проверяем разрешение
	allowance, err := manager.Allowance(token.Address, "alice", "bob")
	if err != nil {
		log.Printf("Failed to get allowance: %v", err)
		return
	}

	fmt.Printf("Current allowance: %s\n", allowance.String())

	// bob переводит токены от имени alice
	transferEvent, err := manager.TransferFrom(token.Address, "bob", "alice", "charlie", big.NewInt(300))
	if err != nil {
		log.Printf("TransferFrom failed: %v", err)
		return
	}

	fmt.Printf("TransferFrom: %s transferred %s tokens from %s to %s\n",
		"bob", transferEvent.Value.String(), transferEvent.From, transferEvent.To)

	// Проверяем оставшееся разрешение
	allowance, _ = manager.Allowance(token.Address, "alice", "bob")
	fmt.Printf("Remaining allowance: %s\n", allowance.String())
}

func demonstrateMintingAndBurning(manager *tokens.ERC20Manager) {
	// Получаем первый токен
	tokenList := manager.ListTokens()
	if len(tokenList) == 0 {
		fmt.Println("No tokens available for minting/burning demo")
		return
	}

	token := tokenList[0]
	fmt.Printf("Using token: %s (%s)\n", token.Name, token.Symbol)

	// Показываем начальное состояние
	initialSupply, _ := manager.GetTotalSupply(token.Address)
	ownerBalance, _ := manager.BalanceOf(token.Address, token.Owner)
	fmt.Printf("Initial total supply: %s\n", initialSupply.String())
	fmt.Printf("Owner balance: %s\n", ownerBalance.String())

	// Создаем новые токены (только владелец может)
	err := manager.Mint(token.Address, token.Owner, big.NewInt(10000))
	if err != nil {
		log.Printf("Minting failed: %v", err)
	} else {
		fmt.Println("Minted 10,000 new tokens")
	}

	// Показываем состояние после создания
	supply, _ := manager.GetTotalSupply(token.Address)
	balance, _ := manager.BalanceOf(token.Address, token.Owner)
	fmt.Printf("Total supply after minting: %s\n", supply.String())
	fmt.Printf("Owner balance after minting: %s\n", balance.String())

	// Сжигаем токены
	err = manager.Burn(token.Address, token.Owner, big.NewInt(5000))
	if err != nil {
		log.Printf("Burning failed: %v", err)
	} else {
		fmt.Println("Burned 5,000 tokens")
	}

	// Показываем финальное состояние
	supply, _ = manager.GetTotalSupply(token.Address)
	balance, _ = manager.BalanceOf(token.Address, token.Owner)
	fmt.Printf("Final total supply: %s\n", supply.String())
	fmt.Printf("Final owner balance: %s\n", balance.String())
}

func demonstrateTokenStats(manager *tokens.ERC20Manager) {
	tokenList := manager.ListTokens()
	if len(tokenList) == 0 {
		fmt.Println("No tokens available for stats demo")
		return
	}

	// Показываем информацию о всех токенах
	for i, token := range tokenList {
		fmt.Printf("Token %d: %s (%s)\n", i+1, token.Name, token.Symbol)

		info, err := manager.GetTokenInfo(token.Address)
		if err != nil {
			log.Printf("Failed to get token info: %v", err)
			continue
		}

		fmt.Printf("  Address: %s\n", info["address"])
		fmt.Printf("  Total Supply: %s\n", info["total_supply"])
		fmt.Printf("  Holders: %d\n", info["holder_count"])
		fmt.Printf("  Created: %s\n", info["created_at"])
		fmt.Println()

		// Показываем статистику
		stats, err := manager.GetTokenStats(token.Address)
		if err != nil {
			log.Printf("Failed to get token stats: %v", err)
			continue
		}

		fmt.Printf("  Statistics:\n")
		fmt.Printf("    Total Holders: %d\n", stats["total_holders"])
		fmt.Printf("    Circulation: %s\n", stats["circulation"])
		fmt.Printf("    Max Balance: %s\n", stats["max_balance"])
		fmt.Printf("    Active Allowances: %d\n", stats["active_allowances"])
		fmt.Printf("    Burned: %s\n", stats["burned"])
		fmt.Println()
	}

	// Демонстрируем поиск
	fmt.Println("Searching for tokens with symbol 'MIRO':")
	criteria := map[string]interface{}{
		"symbol": "MIRO",
	}
	results := manager.SearchTokens(criteria)
	fmt.Printf("Found %d tokens\n", len(results))
	for _, token := range results {
		fmt.Printf("  - %s (%s)\n", token.Name, token.Symbol)
	}
}

func demonstrateTokenExportImport(manager *tokens.ERC20Manager) {
	tokenList := manager.ListTokens()
	if len(tokenList) == 0 {
		fmt.Println("No tokens available for export/import demo")
		return
	}

	token := tokenList[0]
	fmt.Printf("Exporting token: %s (%s)\n", token.Name, token.Symbol)

	// Экспортируем токен
	exportData, err := manager.ExportToken(token.Address)
	if err != nil {
		log.Printf("Export failed: %v", err)
		return
	}

	fmt.Printf("Token exported: %d bytes\n", len(exportData))

	// Создаем новый менеджер для импорта
	newManager := tokens.NewERC20Manager()

	// Импортируем токен
	importedToken, err := newManager.ImportToken(exportData)
	if err != nil {
		log.Printf("Import failed: %v", err)
		return
	}

	fmt.Printf("Token imported successfully:\n")
	fmt.Printf("  Name: %s\n", importedToken.Name)
	fmt.Printf("  Symbol: %s\n", importedToken.Symbol)
	fmt.Printf("  Address: %s\n", importedToken.Address)
	fmt.Printf("  Total Supply: %s\n", importedToken.TotalSupply.String())
	fmt.Printf("  Owner: %s\n", importedToken.Owner)

	// Проверяем, что токен работает в новом менеджере
	balance, err := newManager.BalanceOf(importedToken.Address, importedToken.Owner)
	if err != nil {
		log.Printf("Failed to get balance: %v", err)
		return
	}

	fmt.Printf("  Owner Balance: %s\n", balance.String())
}
