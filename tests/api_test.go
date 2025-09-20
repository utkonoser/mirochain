package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"mirochain/internal/api"
	"mirochain/internal/blockchain"
	"mirochain/internal/wallet"
)

func TestAPIServer(t *testing.T) {
	// Создаем кошелек
	wm := wallet.NewWalletManager()
	testWallet, err := wm.CreateWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Создаем блокчейн
	bc := blockchain.NewBlockchain(testWallet.GetAddress(), testWallet.GetPublicKeyBytes(), 0)

	// Создаем API сервер
	server := api.NewServer(bc, wm, 0) // Порт 0 для автоматического выбора

	// Тестируем /api/status
	t.Run("StatusEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()
		server.HandleStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", response["status"])
		}
	})

	// Тестируем /api/blocks
	t.Run("BlocksEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/blocks", nil)
		w := httptest.NewRecorder()
		server.HandleBlocks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := response["blocks"]; !ok {
			t.Error("Response should contain 'blocks' field")
		}
	})

	// Тестируем /api/blocks с параметрами
	t.Run("BlocksEndpointWithParams", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/blocks?limit=5&offset=0", nil)
		w := httptest.NewRecorder()
		server.HandleBlocks(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["limit"] != float64(5) {
			t.Errorf("Expected limit 5, got %v", response["limit"])
		}
	})

	// Тестируем /api/blocks/{height}
	t.Run("BlockByHeightEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/blocks/0", nil)
		w := httptest.NewRecorder()
		server.HandleBlockByHeight(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := response["block"]; !ok {
			t.Error("Response should contain 'block' field")
		}
	})

	// Тестируем /api/blocks/{height} с несуществующим блоком
	t.Run("BlockByHeightNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/blocks/999", nil)
		w := httptest.NewRecorder()
		server.HandleBlockByHeight(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	// Тестируем /api/blocks/{height} с некорректной высотой
	t.Run("BlockByHeightInvalid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/blocks/invalid", nil)
		w := httptest.NewRecorder()
		server.HandleBlockByHeight(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	// Тестируем /api/transactions
	t.Run("TransactionsEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/transactions", nil)
		w := httptest.NewRecorder()
		server.HandleTransactions(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := response["transactions"]; !ok {
			t.Error("Response should contain 'transactions' field")
		}
	})

	// Тестируем /api/wallets
	t.Run("WalletsEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wallets", nil)
		w := httptest.NewRecorder()
		server.HandleWallets(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if _, ok := response["wallets"]; !ok {
			t.Error("Response should contain 'wallets' field")
		}

		if response["count"] != float64(1) {
			t.Errorf("Expected count 1, got %v", response["count"])
		}
	})

	// Тестируем /api/wallets/{address}
	t.Run("WalletByAddressEndpoint", func(t *testing.T) {
		address := testWallet.GetAddress()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/wallets/%s", address), nil)
		w := httptest.NewRecorder()
		server.HandleWalletByAddress(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["address"] != address {
			t.Errorf("Expected address %s, got %v", address, response["address"])
		}
	})

	// Тестируем /api/wallets/{address} с несуществующим кошельком
	t.Run("WalletByAddressNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/wallets/nonexistent", nil)
		w := httptest.NewRecorder()
		server.HandleWalletByAddress(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	// Тестируем /api/balance/{address}
	t.Run("BalanceEndpoint", func(t *testing.T) {
		address := testWallet.GetAddress()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/balance/%s", address), nil)
		w := httptest.NewRecorder()
		server.HandleBalance(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["address"] != address {
			t.Errorf("Expected address %s, got %v", address, response["address"])
		}

		if _, ok := response["balance"]; !ok {
			t.Error("Response should contain 'balance' field")
		}
	})

	// Тестируем /api/utxos/{address}
	t.Run("UTXOsEndpoint", func(t *testing.T) {
		address := testWallet.GetAddress()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/utxos/%s", address), nil)
		w := httptest.NewRecorder()
		server.HandleUTXOs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["address"] != address {
			t.Errorf("Expected address %s, got %v", address, response["address"])
		}

		if _, ok := response["utxos"]; !ok {
			t.Error("Response should contain 'utxos' field")
		}
	})

	// Тестируем неподдерживаемые HTTP методы
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/status", nil)
		w := httptest.NewRecorder()
		server.HandleStatus(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Log("API server tests completed successfully")
}
