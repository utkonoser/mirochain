//go:build nft_api_demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"mirochain/internal/nft"
)

func main() {
	fmt.Println("=== NFT API Demo ===")
	fmt.Println()

	// Создаем менеджер NFT и API
	manager := nft.NewERC721Manager()
	nftAPI := nft.NewNFTAPI(manager)

	// Создаем HTTP сервер для тестирования
	mux := http.NewServeMux()
	nftAPI.RegisterRoutes(mux)

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Starting NFT API server on port 8890...")
		err := http.ListenAndServe(":8890", mux)
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(2 * time.Second)

	// Тестируем endpoints
	testCreateContract()
	testMintNFT()
	testTransferNFT()
	testApproveNFT()
	testGetToken()
	testListContracts()
	testContractStats()
	testSearchNFTs()
	testBurnNFT()
}

func testCreateContract() {
	fmt.Println("1. Testing /api/nft/create-contract")
	
	createRequest := map[string]interface{}{
		"name":       "Test NFT Collection",
		"symbol":     "TNC",
		"owner":      "alice",
		"base_uri":   "https://api.testnft.com/metadata/",
		"max_supply": "1000",
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/create-contract", "application/json", bytes.NewBuffer(jsonData))
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

func testMintNFT() {
	fmt.Println("2. Testing /api/nft/mint")
	
	// Сначала создаем контракт
	createContract()
	
	mintRequest := map[string]interface{}{
		"contract_address": "nft_contract_placeholder", // Будет заменен реальным адресом
		"to":              "alice",
		"token_id":        "1",
		"metadata": map[string]interface{}{
			"name":        "Test NFT",
			"description": "A test NFT for demonstration",
			"image":       "https://api.testnft.com/images/1.jpg",
			"external_url": "https://testnft.com/1",
		},
		"attributes": map[string]interface{}{
			"artist": "Test Artist",
			"style":  "Digital",
			"rarity": "common",
		},
	}

	jsonData, err := json.Marshal(mintRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/mint", "application/json", bytes.NewBuffer(jsonData))
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

func testTransferNFT() {
	fmt.Println("3. Testing /api/nft/transfer")
	
	transferRequest := map[string]interface{}{
		"contract_address": "nft_contract_placeholder",
		"from":            "alice",
		"to":              "bob",
		"token_id":        "1",
	}

	jsonData, err := json.Marshal(transferRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/transfer", "application/json", bytes.NewBuffer(jsonData))
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

func testApproveNFT() {
	fmt.Println("4. Testing /api/nft/approve")
	
	approveRequest := map[string]interface{}{
		"contract_address": "nft_contract_placeholder",
		"owner":           "alice",
		"approved":        "bob",
		"token_id":        "1",
	}

	jsonData, err := json.Marshal(approveRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/approve", "application/json", bytes.NewBuffer(jsonData))
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

func testGetToken() {
	fmt.Println("5. Testing /api/nft/get-token")
	
	resp, err := http.Get("http://localhost:8890/api/nft/get-token?contract_address=nft_contract_placeholder&token_id=1")
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

func testListContracts() {
	fmt.Println("6. Testing /api/nft/list-contracts")
	
	resp, err := http.Get("http://localhost:8890/api/nft/list-contracts")
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

func testContractStats() {
	fmt.Println("7. Testing /api/nft/contract-stats")
	
	resp, err := http.Get("http://localhost:8890/api/nft/contract-stats?address=nft_contract_placeholder")
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

func testSearchNFTs() {
	fmt.Println("8. Testing /api/nft/search")
	
	searchRequest := map[string]interface{}{
		"attributes": map[string]interface{}{
			"rarity": "common",
		},
	}

	jsonData, err := json.Marshal(searchRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/search", "application/json", bytes.NewBuffer(jsonData))
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

func testBurnNFT() {
	fmt.Println("9. Testing /api/nft/burn")
	
	burnRequest := map[string]interface{}{
		"contract_address": "nft_contract_placeholder",
		"owner":           "alice",
		"token_id":        "1",
	}

	jsonData, err := json.Marshal(burnRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/burn", "application/json", bytes.NewBuffer(jsonData))
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

func createContract() {
	createRequest := map[string]interface{}{
		"name":       "Test Contract",
		"symbol":     "TEST",
		"owner":      "alice",
		"base_uri":   "https://api.test.com/metadata/",
		"max_supply": "1000",
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		return
	}

	resp, err := http.Post("http://localhost:8890/api/nft/create-contract", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
