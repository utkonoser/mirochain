//go:build contract_api_demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"mirochain/internal/vm"
)

func main() {
	fmt.Println("=== Contract API Test ===")
	fmt.Println()

	// Создаем VM и API
	vmInstance := vm.NewVM(1000000)
	contractAPI := vm.NewContractAPI(vmInstance)

	// Создаем HTTP сервер для тестирования
	mux := http.NewServeMux()
	contractAPI.RegisterRoutes(mux)

	// Запускаем сервер в горутине
	go func() {
		fmt.Println("Starting test server on port 8888...")
		err := http.ListenAndServe(":8888", mux)
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Ждем запуска сервера
	time.Sleep(2 * time.Second)

	// Тестируем endpoints
	testTemplates()
	testDeploy()
	testList()
}

func testTemplates() {
	fmt.Println("1. Testing /api/contracts/templates")
	
	resp, err := http.Get("http://localhost:8888/api/contracts/templates")
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

func testDeploy() {
	fmt.Println("2. Testing /api/contracts/deploy")
	
	deployRequest := map[string]interface{}{
		"code":            "PUSH 42\nSTORE \"value\"\nRETURN",
		"owner":           "test_user",
		"initial_balance": "1000",
		"gas_limit":       100000,
	}

	jsonData, err := json.Marshal(deployRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8888/api/contracts/deploy", "application/json", bytes.NewBuffer(jsonData))
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

func testList() {
	fmt.Println("3. Testing /api/contracts/list")
	
	resp, err := http.Get("http://localhost:8888/api/contracts/list")
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
