//go:build token_api_demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"mirochain/internal/tokens"
)

func main() {
	fmt.Println("=== Token API Demo ===")
	fmt.Println()

	// Создаем менеджер токенов и API
	manager := tokens.NewERC20Manager()
	tokenAPI := tokens.NewTokenAPI(manager)

	// Создаем HTTP сервер для тестирования
	mux := http.NewServeMux()
	tokenAPI.RegisterRoutes(mux)

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Starting Token API server on port 8889...")
		err := http.ListenAndServe(":8889", mux)
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(2 * time.Second)

	// Тестируем endpoints
	testCreateToken()
	testTransferTokens()
	testApproveTokens()
	testGetBalance()
	testListTokens()
	testTokenStats()
	testSearchTokens()
	testMintAndBurn()
}

func testCreateToken() {
	fmt.Println("1. Testing /api/tokens/create")

	createRequest := map[string]interface{}{
		"name":         "DemoToken",
		"symbol":       "DEMO",
		"decimals":     18,
		"total_supply": "1000000",
		"owner":        "alice",
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8889/api/tokens/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testTransferTokens() {
	fmt.Println("2. Testing /api/tokens/transfer")

	// Сначала создаем токен
	createToken()

	transferRequest := map[string]interface{}{
		"token_address": "token_placeholder", // Будет заменен реальным адресом
		"from":          "alice",
		"to":            "bob",
		"amount":        "1000",
	}

	jsonData, err := json.Marshal(transferRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8889/api/tokens/transfer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testApproveTokens() {
	fmt.Println("3. Testing /api/tokens/approve")

	approveRequest := map[string]interface{}{
		"token_address": "token_placeholder",
		"owner":         "alice",
		"spender":       "bob",
		"amount":        "500",
	}

	jsonData, err := json.Marshal(approveRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8889/api/tokens/approve", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testGetBalance() {
	fmt.Println("4. Testing /api/tokens/balance")

	resp, err := http.Get("http://localhost:8889/api/tokens/balance?token_address=token_placeholder&address=alice")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testListTokens() {
	fmt.Println("5. Testing /api/tokens/list")

	resp, err := http.Get("http://localhost:8889/api/tokens/list")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testTokenStats() {
	fmt.Println("6. Testing /api/tokens/stats")

	resp, err := http.Get("http://localhost:8889/api/tokens/stats?address=token_placeholder")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testSearchTokens() {
	fmt.Println("7. Testing /api/tokens/search")

	searchRequest := map[string]interface{}{
		"symbol": "DEMO",
	}

	jsonData, err := json.Marshal(searchRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8889/api/tokens/search", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println()
}

func testMintAndBurn() {
	fmt.Println("8. Testing /api/tokens/mint and /api/tokens/burn")

	// Тест создания токенов
	mintRequest := map[string]interface{}{
		"token_address": "token_placeholder",
		"to":            "alice",
		"amount":        "10000",
	}

	jsonData, err := json.Marshal(mintRequest)
	if err != nil {
		log.Printf("Error marshaling mint request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8889/api/tokens/mint", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Mint error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading mint response: %v", err)
		return
	}

	fmt.Printf("Mint Status: %d\n", resp.StatusCode)
	fmt.Printf("Mint Response: %s\n", string(body))

	// Тест сжигания токенов
	burnRequest := map[string]interface{}{
		"token_address": "token_placeholder",
		"from":          "alice",
		"amount":        "5000",
	}

	jsonData, err = json.Marshal(burnRequest)
	if err != nil {
		log.Printf("Error marshaling burn request: %v", err)
		return
	}

	resp, err = http.Post("http://localhost:8889/api/tokens/burn", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Burn error: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading burn response: %v", err)
		return
	}

	fmt.Printf("Burn Status: %d\n", resp.StatusCode)
	fmt.Printf("Burn Response: %s\n", string(body))
	fmt.Println()
}

func createToken() {
	createRequest := map[string]interface{}{
		"name":         "TestToken",
		"symbol":       "TEST",
		"decimals":     18,
		"total_supply": "1000000",
		"owner":        "alice",
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		return
	}

	resp, err := http.Post("http://localhost:8889/api/tokens/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
