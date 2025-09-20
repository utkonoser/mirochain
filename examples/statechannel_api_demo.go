//go:build statechannel_api_demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== State Channel API Demo ===")
	fmt.Println()

	// Тестируем endpoints
	testCreateChannel()
	testListChannels()
	testDeposit()
	testCreateTransaction()
	testWithdraw()
	testInitiateDispute()
	testGetChannelStats()
	testGetChannelHistory()
}

func testCreateChannel() {
	fmt.Println("1. Testing /api/statechannel/create")
	
	createRequest := map[string]interface{}{
		"participants":   []string{"alice", "bob"},
		"channel_type":   "payment",
		"dispute_period": 86400, // 24 hours in seconds
		"metadata": map[string]interface{}{
			"description": "Test payment channel",
			"max_amount":  "1000000",
		},
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8892/api/statechannel/create", "application/json", bytes.NewBuffer(jsonData))
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

func testListChannels() {
	fmt.Println("2. Testing /api/statechannel/list")
	
	resp, err := http.Get("http://localhost:8892/api/statechannel/list")
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

func testDeposit() {
	fmt.Println("3. Testing /api/statechannel/deposit")
	
	// Сначала создаем канал
	createChannel()
	
	depositRequest := map[string]interface{}{
		"channel_id":  "channel_placeholder", // Будет заменен реальным ID
		"participant": "alice",
		"amount":      "100000",
		"tx_hash":     "0x1234567890abcdef",
	}

	jsonData, err := json.Marshal(depositRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8892/api/statechannel/deposit", "application/json", bytes.NewBuffer(jsonData))
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

func testCreateTransaction() {
	fmt.Println("4. Testing /api/statechannel/create-transaction")
	
	createTransactionRequest := map[string]interface{}{
		"channel_id": "channel_placeholder",
		"from":       "alice",
		"to":         "bob",
		"amount":     "5000",
		"data": map[string]interface{}{
			"description": "Test payment",
			"category":    "payment",
		},
		"signature": "alice_signature_1",
	}

	jsonData, err := json.Marshal(createTransactionRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8892/api/statechannel/create-transaction", "application/json", bytes.NewBuffer(jsonData))
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

func testWithdraw() {
	fmt.Println("5. Testing /api/statechannel/withdraw")
	
	withdrawRequest := map[string]interface{}{
		"channel_id":  "channel_placeholder",
		"participant": "alice",
		"amount":      "20000",
	}

	jsonData, err := json.Marshal(withdrawRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8892/api/statechannel/withdraw", "application/json", bytes.NewBuffer(jsonData))
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

func testInitiateDispute() {
	fmt.Println("6. Testing /api/statechannel/initiate-dispute")
	
	disputeRequest := map[string]interface{}{
		"channel_id": "channel_placeholder",
		"initiator":  "alice",
		"reason":     "Balance calculation error",
		"evidence":   "Evidence of incorrect calculation",
		"state_update": map[string]interface{}{
			"channel_id":    "channel_placeholder",
			"nonce":         1,
			"balances": map[string]string{
				"alice": "50000",
				"bob":   "200000",
			},
			"participants": []string{"alice", "bob"},
			"data": map[string]interface{}{
				"dispute_reason": "Incorrect balance calculation",
			},
			"signature":    "disputed_signature",
			"timestamp":    time.Now(),
			"update_type":  "dispute",
		},
	}

	jsonData, err := json.Marshal(disputeRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return
	}

	resp, err := http.Post("http://localhost:8892/api/statechannel/initiate-dispute", "application/json", bytes.NewBuffer(jsonData))
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

func testGetChannelStats() {
	fmt.Println("7. Testing /api/statechannel/stats")
	
	resp, err := http.Get("http://localhost:8892/api/statechannel/stats?channel_id=channel_placeholder")
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

func testGetChannelHistory() {
	fmt.Println("8. Testing /api/statechannel/history")
	
	resp, err := http.Get("http://localhost:8892/api/statechannel/history?channel_id=channel_placeholder")
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

func createChannel() {
	createRequest := map[string]interface{}{
		"participants":   []string{"alice", "bob"},
		"channel_type":   "payment",
		"dispute_period": 86400,
		"metadata": map[string]interface{}{
			"description": "Test payment channel",
			"max_amount":  "1000000",
		},
	}

	jsonData, err := json.Marshal(createRequest)
	if err != nil {
		return
	}

	resp, err := http.Post("http://localhost:8892/api/statechannel/create", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
