//go:build sidechain_api_demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"mirochain/internal/sidechain"
)

func main() {
	fmt.Println("=== Sidechain API Demo ===")
	fmt.Println()

	// Создаем менеджер sidechains и API
	manager := sidechain.NewSidechainManager()
	sidechainAPI := sidechain.NewSidechainAPI(manager)

	// Создаем HTTP сервер для тестирования
	mux := http.NewServeMux()
	sidechainAPI.RegisterRoutes(mux)

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Starting Sidechain API server on port 8891...")
		err := http.ListenAndServe(":8891", mux)
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(2 * time.Second)

	// Тестируем endpoints
	testCreateSidechain()
	testListSidechains()
	testCreateAsset()
	testAddBlock()
	testAddTransaction()
	testCreateBridgeTransaction()
	testSendCrossChainMessage()
	testGetSidechainStats()
	testUpdateSidechainStatus()
	testAddValidator()
}

func testCreateSidechain() {
	fmt.Println("1. Testing /api/sidechain/create")
	
	createRequest := map[string]interface{}{
		"name":         "Test Sidechain",
		"description":  "A test sidechain for demonstration",
		"creator":      "alice",
		"parent_chain": "main",
		"config": map[string]interface{}{
			"consensus_algorithm": "PoS",
			"block_time":         2,
			"difficulty":         3,
			"max_block_size":     1048576,
			"gas_limit":          1000000,
			"validator_count":    5,
			"bridge_enabled":     true,
			"cross_chain_enabled": true,
		},
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/create", "application/json", bytes.NewBuffer(jsonData))
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

func testListSidechains() {
	fmt.Println("2. Testing /api/sidechain/list")
	
	resp, err := http.Get("http://localhost:8891/api/sidechain/list")
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

func testCreateAsset() {
	fmt.Println("3. Testing /api/sidechain/create-asset")
	
	// Сначала создаем sidechain
	createSidechain()
	
	createAssetRequest := map[string]interface{}{
		"sidechain_id": "sidechain_placeholder", // Будет заменен реальным ID
		"name":         "Test Token",
		"symbol":       "TEST",
		"decimals":     18,
		"total_supply": "1000000",
		"creator":      "alice",
		"type":         "token",
	}

	jsonData, err := json.Marshal(createAssetRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/create-asset", "application/json", bytes.NewBuffer(jsonData))
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

func testAddBlock() {
	fmt.Println("4. Testing /api/sidechain/add-block")
	
	addBlockRequest := map[string]interface{}{
		"sidechain_id": "sidechain_placeholder",
		"block": map[string]interface{}{
			"index":         1,
			"timestamp":     time.Now(),
			"previous_hash": "genesis",
			"hash":          "block_hash_1",
			"merkle_root":   "merkle_root_1",
			"nonce":         1000,
			"difficulty":    3,
			"transactions":  []interface{}{},
			"validator":     "alice",
			"signature":     "signature_1",
		},
	}

	jsonData, err := json.Marshal(addBlockRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/add-block", "application/json", bytes.NewBuffer(jsonData))
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

func testAddTransaction() {
	fmt.Println("5. Testing /api/sidechain/add-transaction")
	
	addTransactionRequest := map[string]interface{}{
		"sidechain_id": "sidechain_placeholder",
		"transaction": map[string]interface{}{
			"id":        "tx_1",
			"type":      "transfer",
			"from":      "alice",
			"to":        "bob",
			"amount":    "1000",
			"asset":     "native",
			"gas_limit": 21000,
			"gas_price": "20",
			"nonce":     1,
			"timestamp": time.Now(),
		},
	}

	jsonData, err := json.Marshal(addTransactionRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/add-transaction", "application/json", bytes.NewBuffer(jsonData))
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

func testCreateBridgeTransaction() {
	fmt.Println("6. Testing /api/sidechain/create-bridge-tx")
	
	createBridgeTxRequest := map[string]interface{}{
		"source_chain": "sidechain_1",
		"target_chain": "sidechain_2",
		"asset":        "native",
		"amount":       "1000",
		"from":         "alice",
		"to":           "bob",
	}

	jsonData, err := json.Marshal(createBridgeTxRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/create-bridge-tx", "application/json", bytes.NewBuffer(jsonData))
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

func testSendCrossChainMessage() {
	fmt.Println("7. Testing /api/sidechain/send-message")
	
	sendMessageRequest := map[string]interface{}{
		"source_chain": "sidechain_1",
		"target_chain": "sidechain_2",
		"type":         "asset_transfer",
		"data": map[string]interface{}{
			"asset":  "native",
			"amount": "1000",
			"from":   "alice",
			"to":     "bob",
		},
	}

	jsonData, err := json.Marshal(sendMessageRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/send-message", "application/json", bytes.NewBuffer(jsonData))
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

func testGetSidechainStats() {
	fmt.Println("8. Testing /api/sidechain/stats")
	
	resp, err := http.Get("http://localhost:8891/api/sidechain/stats?id=sidechain_placeholder")
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

func testUpdateSidechainStatus() {
	fmt.Println("9. Testing /api/sidechain/update-status")
	
	updateStatusRequest := map[string]interface{}{
		"id":     "sidechain_placeholder",
		"status": "paused",
	}

	jsonData, err := json.Marshal(updateStatusRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/update-status", "application/json", bytes.NewBuffer(jsonData))
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

func testAddValidator() {
	fmt.Println("10. Testing /api/sidechain/add-validator")
	
	addValidatorRequest := map[string]interface{}{
		"id":        "sidechain_placeholder",
		"validator": "new_validator",
	}

	jsonData, err := json.Marshal(addValidatorRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/add-validator", "application/json", bytes.NewBuffer(jsonData))
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

func createSidechain() {
	createRequest := map[string]interface{}{
		"name":         "Test Sidechain",
		"description":  "A test sidechain",
		"creator":      "alice",
		"parent_chain": "main",
		"config": map[string]interface{}{
			"consensus_algorithm": "PoS",
			"block_time":         2,
			"difficulty":         3,
			"max_block_size":     1048576,
			"gas_limit":          1000000,
			"validator_count":    5,
			"bridge_enabled":     true,
			"cross_chain_enabled": true,
		},
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		return
	}

	resp, err := http.Post("http://localhost:8891/api/sidechain/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
