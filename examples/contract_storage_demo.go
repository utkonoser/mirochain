//go:build contract_storage_demo

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Contract Storage Demo ===")

	// Ждем запуска узла
	fmt.Println("Waiting for node to start...")
	time.Sleep(3 * time.Second)

	// Тестируем создание контракта
	fmt.Println("\n1. Creating a contract...")
	contractAddress := createContract()
	if contractAddress == "" {
		log.Fatal("Failed to create contract")
	}
	fmt.Printf("Contract created with address: %s\n", contractAddress)

	// Тестируем установку значений в хранилище
	fmt.Println("\n2. Setting storage values...")
	setStorageValue(contractAddress, "counter", "100")
	setStorageValue(contractAddress, "owner", "alice")
	setStorageValue(contractAddress, "balance", "1000")

	// Тестируем получение значений из хранилища
	fmt.Println("\n3. Getting storage values...")
	getStorageValue(contractAddress, "counter")
	getStorageValue(contractAddress, "owner")
	getStorageValue(contractAddress, "balance")

	// Тестируем получение всего хранилища
	fmt.Println("\n4. Getting contract storage...")
	getContractStorage(contractAddress)

	// Тестируем статистику
	fmt.Println("\n5. Getting contract stats...")
	getContractStats()

	fmt.Println("\n=== Contract Storage Demo Complete ===")
}

func createContract() string {
	// Создаем простой контракт-счетчик
	code := `PUSH 0
SSTORE counter
PUSH 1
SLOAD counter
ADD
SSTORE counter
SLOAD counter
RETURN`

	req := map[string]interface{}{
		"code":            code,
		"owner":           "alice",
		"initial_balance": "0",
	}

	resp, err := http.Post("http://localhost:12901/api/contracts/deploy", 
		"application/json", 
		jsonEncode(req))
	if err != nil {
		log.Printf("Error creating contract: %v", err)
		return ""
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return ""
	}

	if success, ok := result["success"].(bool); !ok || !success {
		log.Printf("Contract creation failed: %v", result["error"])
		return ""
	}

	contractAddress, ok := result["contract_address"].(string)
	if !ok {
		log.Printf("Invalid contract address in response")
		return ""
	}
	return contractAddress
}

func setStorageValue(address, key, value string) {
	req := map[string]interface{}{
		"address": address,
		"key":     key,
		"value":   value,
	}

	resp, err := http.Post("http://localhost:12901/api/contracts/storage/set", 
		"application/json", 
		jsonEncode(req))
	if err != nil {
		log.Printf("Error setting storage value: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	if success, ok := result["success"].(bool); !ok || !success {
		log.Printf("Failed to set storage value: %v", result["error"])
		return
	}

	fmt.Printf("Set %s = %s for contract %s\n", key, value, address)
}

func getStorageValue(address, key string) {
	url := fmt.Sprintf("http://localhost:12901/api/contracts/storage/get?address=%s&key=%s", 
		address, key)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error getting storage value: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	value := result["value"].(string)
	fmt.Printf("Get %s = %s from contract %s\n", key, value, address)
}

func getContractStorage(address string) {
	url := fmt.Sprintf("http://localhost:12901/api/contracts/storage/%s", address)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error getting contract storage: %v", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	storage := result["storage"].(map[string]interface{})
	fmt.Printf("Contract %s storage:\n", address)
	for key, value := range storage {
		fmt.Printf("  %s = %v\n", key, value)
	}
}

func getContractStats() {
	resp, err := http.Get("http://localhost:12901/api/contracts/stats")
	if err != nil {
		log.Printf("Error getting contract stats: %v", err)
		return
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	fmt.Printf("Contract Statistics:\n")
	fmt.Printf("  Total contracts: %v\n", stats["total_contracts"])
	fmt.Printf("  Total storage keys: %v\n", stats["total_storage_keys"])
	fmt.Printf("  Average keys per contract: %v\n", stats["average_storage_keys_per_contract"])
}

func jsonEncode(v interface{}) *strings.Reader {
	data, _ := json.Marshal(v)
	return strings.NewReader(string(data))
}
