package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"mirochain/internal/blockchain"
	"mirochain/internal/wallet"
)

// Server представляет API сервер
type Server struct {
	Blockchain    *blockchain.Blockchain
	WalletManager *wallet.WalletManager
	Mempool       interface{} // Mempool interface для избежания циклических импортов
	Port          int
}

// NewServer создает новый API сервер
func NewServer(bc *blockchain.Blockchain, wm *wallet.WalletManager, port int) *Server {
	return &Server{
		Blockchain:    bc,
		WalletManager: wm,
		Port:          port,
	}
}

// NewServerWithMempool создает новый API сервер с mempool
func NewServerWithMempool(bc *blockchain.Blockchain, wm *wallet.WalletManager, mempool interface{}, port int) *Server {
	return &Server{
		Blockchain:    bc,
		WalletManager: wm,
		Mempool:       mempool,
		Port:          port,
	}
}

// Start запускает API сервер
func (s *Server) Start() error {
	// Регистрируем маршруты
	http.HandleFunc("/api/status", s.handleStatus)
	http.HandleFunc("/api/blocks", s.handleBlocks)
	http.HandleFunc("/api/blocks/", s.handleBlockByHeight)
	http.HandleFunc("/api/transactions", s.handleTransactions)
	http.HandleFunc("/api/wallets", s.handleWallets)
	http.HandleFunc("/api/wallets/", s.handleWalletByAddress)
	http.HandleFunc("/api/balance/", s.handleBalance)
	http.HandleFunc("/api/utxos/", s.handleUTXOs)

	// Запускаем сервер
	addr := fmt.Sprintf(":%d", s.Port)
	slog.Info("Starting API server", "address", addr)

	return http.ListenAndServe(addr, nil)
}

// HandleStatus публичный метод для тестирования
func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	s.handleStatus(w, r)
}

// handleStatus обрабатывает запрос статуса
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.Blockchain.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data":   stats,
	})
}

// HandleBlocks публичный метод для тестирования
func (s *Server) HandleBlocks(w http.ResponseWriter, r *http.Request) {
	s.handleBlocks(w, r)
}

// handleBlocks обрабатывает запрос списка блоков
func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры запроса
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Получаем блоки
	height := s.Blockchain.GetHeight()
	var blocks []*blockchain.Block

	for i := int64(offset); i <= height && len(blocks) < limit; i++ {
		block := s.Blockchain.GetBlockByHeight(i)
		if block != nil {
			blocks = append(blocks, block)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blocks": blocks,
		"total":  height + 1,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleBlockByHeight публичный метод для тестирования
func (s *Server) HandleBlockByHeight(w http.ResponseWriter, r *http.Request) {
	s.handleBlockByHeight(w, r)
}

// handleBlockByHeight обрабатывает запрос блока по высоте
func (s *Server) handleBlockByHeight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем высоту из URL
	path := r.URL.Path
	heightStr := path[len("/api/blocks/"):]

	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid height", http.StatusBadRequest)
		return
	}

	block := s.Blockchain.GetBlockByHeight(height)
	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"block": block,
	})
}

// HandleTransactions публичный метод для тестирования
func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	s.handleTransactions(w, r)
}

// handleTransactions обрабатывает запрос транзакций
func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Реализовать получение транзакций из mempool
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": []interface{}{},
		"message":      "Transaction pool not implemented yet",
	})
}

// HandleWallets публичный метод для тестирования
func (s *Server) HandleWallets(w http.ResponseWriter, r *http.Request) {
	s.handleWallets(w, r)
}

// handleWallets обрабатывает запрос списка кошельков
func (s *Server) handleWallets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wallets := s.WalletManager.GetWallets()
	var walletList []map[string]interface{}

	for address, wallet := range wallets {
		walletList = append(walletList, map[string]interface{}{
			"address":    address,
			"public_key": fmt.Sprintf("%x", wallet.GetPublicKeyBytes()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallets": walletList,
		"count":   len(walletList),
	})
}

// HandleWalletByAddress публичный метод для тестирования
func (s *Server) HandleWalletByAddress(w http.ResponseWriter, r *http.Request) {
	s.handleWalletByAddress(w, r)
}

// handleWalletByAddress обрабатывает запрос кошелька по адресу
func (s *Server) handleWalletByAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем адрес из URL
	path := r.URL.Path
	address := path[len("/api/wallets/"):]

	wallet, exists := s.WalletManager.GetWallet(address)
	if !exists {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}

	balance := s.Blockchain.GetBalance(address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":    address,
		"public_key": fmt.Sprintf("%x", wallet.GetPublicKeyBytes()),
		"balance":    balance,
	})
}

// HandleBalance публичный метод для тестирования
func (s *Server) HandleBalance(w http.ResponseWriter, r *http.Request) {
	s.handleBalance(w, r)
}

// handleBalance обрабатывает запрос баланса
func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем адрес из URL
	path := r.URL.Path
	address := path[len("/api/balance/"):]

	balance := s.Blockchain.GetBalance(address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": address,
		"balance": balance,
	})
}

// HandleUTXOs публичный метод для тестирования
func (s *Server) HandleUTXOs(w http.ResponseWriter, r *http.Request) {
	s.handleUTXOs(w, r)
}

// handleUTXOs обрабатывает запрос UTXO
func (s *Server) handleUTXOs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем адрес из URL
	path := r.URL.Path
	address := path[len("/api/utxos/"):]

	utxos := s.Blockchain.GetUTXOsByAddress(address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": address,
		"utxos":   utxos,
		"count":   len(utxos),
	})
}
