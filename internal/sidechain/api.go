package sidechain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
)

// SidechainAPI предоставляет HTTP API для работы с sidechains
type SidechainAPI struct {
	manager *SidechainManager
}

// NewSidechainAPI создает новый API для sidechains
func NewSidechainAPI(manager *SidechainManager) *SidechainAPI {
	return &SidechainAPI{manager: manager}
}

// CreateSidechainRequest представляет запрос на создание sidechain
type CreateSidechainRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Creator     string          `json:"creator"`
	ParentChain string          `json:"parent_chain"`
	Config      SidechainConfig `json:"config"`
}

// CreateSidechainResponse представляет ответ на создание sidechain
type CreateSidechainResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddBlockRequest представляет запрос на добавление блока
type AddBlockRequest struct {
	SidechainID string           `json:"sidechain_id"`
	Block       *SidechainBlock  `json:"block"`
}

// AddBlockResponse представляет ответ на добавление блока
type AddBlockResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AddTransactionRequest представляет запрос на добавление транзакции
type AddTransactionRequest struct {
	SidechainID string                 `json:"sidechain_id"`
	Transaction *SidechainTransaction  `json:"transaction"`
}

// AddTransactionResponse представляет ответ на добавление транзакции
type AddTransactionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CreateAssetRequest представляет запрос на создание актива
type CreateAssetRequest struct {
	SidechainID string   `json:"sidechain_id"`
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Decimals    int      `json:"decimals"`
	TotalSupply string   `json:"total_supply"`
	Creator     string   `json:"creator"`
	Type        string   `json:"type"`
}

// CreateAssetResponse представляет ответ на создание актива
type CreateAssetResponse struct {
	Success bool   `json:"success"`
	Asset   *Asset `json:"asset,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreateBridgeTransactionRequest представляет запрос на создание мостовой транзакции
type CreateBridgeTransactionRequest struct {
	SourceChain string `json:"source_chain"`
	TargetChain string `json:"target_chain"`
	Asset       string `json:"asset"`
	Amount      string `json:"amount"`
	From        string `json:"from"`
	To          string `json:"to"`
}

// CreateBridgeTransactionResponse представляет ответ на создание мостовой транзакции
type CreateBridgeTransactionResponse struct {
	Success bool              `json:"success"`
	Tx      *BridgeTransaction `json:"transaction,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// SendMessageRequest представляет запрос на отправку кросс-чейн сообщения
type SendMessageRequest struct {
	SourceChain string                 `json:"source_chain"`
	TargetChain string                 `json:"target_chain"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
}

// SendMessageResponse представляет ответ на отправку сообщения
type SendMessageResponse struct {
	Success bool               `json:"success"`
	Message *CrossChainMessage `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// CreateSidechain создает новую sidechain
func (api *SidechainAPI) CreateSidechain(w http.ResponseWriter, r *http.Request) {
	var req CreateSidechainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Создаем sidechain
	sidechain, err := api.manager.CreateSidechain(req.Name, req.Description, req.Creator, req.ParentChain, req.Config)
	if err != nil {
		response := CreateSidechainResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := CreateSidechainResponse{
		Success: true,
		ID:      sidechain.ID,
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetSidechain возвращает информацию о sidechain
func (api *SidechainAPI) GetSidechain(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}
	
	sidechain, exists := api.manager.GetSidechain(id)
	if !exists {
		http.Error(w, "sidechain not found", http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(sidechain)
}

// ListSidechains возвращает список всех sidechains
func (api *SidechainAPI) ListSidechains(w http.ResponseWriter, r *http.Request) {
	sidechains := api.manager.ListSidechains()
	json.NewEncoder(w).Encode(sidechains)
}

// AddBlock добавляет блок в sidechain
func (api *SidechainAPI) AddBlock(w http.ResponseWriter, r *http.Request) {
	var req AddBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := api.manager.AddBlock(req.SidechainID, req.Block)
	if err != nil {
		response := AddBlockResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := AddBlockResponse{
		Success: true,
	}
	
	json.NewEncoder(w).Encode(response)
}

// AddTransaction добавляет транзакцию в sidechain
func (api *SidechainAPI) AddTransaction(w http.ResponseWriter, r *http.Request) {
	var req AddTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := api.manager.AddTransaction(req.SidechainID, req.Transaction)
	if err != nil {
		response := AddTransactionResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := AddTransactionResponse{
		Success: true,
	}
	
	json.NewEncoder(w).Encode(response)
}

// CreateAsset создает новый актив в sidechain
func (api *SidechainAPI) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим TotalSupply
	totalSupply, ok := new(big.Int).SetString(req.TotalSupply, 10)
	if !ok {
		http.Error(w, "Invalid total supply", http.StatusBadRequest)
		return
	}
	
	// Парсим тип актива
	var assetType AssetType
	switch req.Type {
	case "native":
		assetType = AssetTypeNative
	case "token":
		assetType = AssetTypeToken
	case "nft":
		assetType = AssetTypeNFT
	case "bridged":
		assetType = AssetTypeBridged
	default:
		http.Error(w, "Invalid asset type", http.StatusBadRequest)
		return
	}
	
	// Создаем актив
	asset, err := api.manager.CreateAsset(req.SidechainID, req.Name, req.Symbol, req.Decimals, totalSupply, req.Creator, assetType)
	if err != nil {
		response := CreateAssetResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := CreateAssetResponse{
		Success: true,
		Asset:   asset,
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetAsset возвращает информацию об активе
func (api *SidechainAPI) GetAsset(w http.ResponseWriter, r *http.Request) {
	sidechainID := r.URL.Query().Get("sidechain_id")
	assetID := r.URL.Query().Get("asset_id")
	
	if sidechainID == "" || assetID == "" {
		http.Error(w, "sidechain_id and asset_id parameters required", http.StatusBadRequest)
		return
	}
	
	asset, err := api.manager.GetAsset(sidechainID, assetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(asset)
}

// ListAssets возвращает список активов в sidechain
func (api *SidechainAPI) ListAssets(w http.ResponseWriter, r *http.Request) {
	sidechainID := r.URL.Query().Get("sidechain_id")
	if sidechainID == "" {
		http.Error(w, "sidechain_id parameter required", http.StatusBadRequest)
		return
	}
	
	assets, err := api.manager.ListAssets(sidechainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(assets)
}

// CreateBridgeTransaction создает мостовую транзакцию
func (api *SidechainAPI) CreateBridgeTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateBridgeTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим Amount
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}
	
	// Создаем мостовую транзакцию
	bridgeTx, err := api.manager.CreateBridgeTransaction(req.SourceChain, req.TargetChain, req.Asset, amount, req.From, req.To)
	if err != nil {
		response := CreateBridgeTransactionResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := CreateBridgeTransactionResponse{
		Success: true,
		Tx:      bridgeTx,
	}
	
	json.NewEncoder(w).Encode(response)
}

// ProcessBridgeTransaction обрабатывает мостовую транзакцию
func (api *SidechainAPI) ProcessBridgeTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TxID           string `json:"tx_id"`
		TargetTxID     string `json:"target_tx_id"`
		ValidatorProof string `json:"validator_proof"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := api.manager.ProcessBridgeTransaction(req.TxID, req.TargetTxID, req.ValidatorProof)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "Bridge transaction processed successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetBridgeTransaction возвращает мостовую транзакцию
func (api *SidechainAPI) GetBridgeTransaction(w http.ResponseWriter, r *http.Request) {
	txID := r.URL.Query().Get("tx_id")
	if txID == "" {
		http.Error(w, "tx_id parameter required", http.StatusBadRequest)
		return
	}
	
	bridgeTx, err := api.manager.GetBridgeTransaction(txID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(bridgeTx)
}

// ListBridgeTransactions возвращает список мостовых транзакций
func (api *SidechainAPI) ListBridgeTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := api.manager.ListBridgeTransactions()
	json.NewEncoder(w).Encode(transactions)
}

// SendCrossChainMessage отправляет кросс-чейн сообщение
func (api *SidechainAPI) SendCrossChainMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Отправляем сообщение
	message, err := api.manager.SendCrossChainMessage(req.SourceChain, req.TargetChain, req.Type, req.Data)
	if err != nil {
		response := SendMessageResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := SendMessageResponse{
		Success: true,
		Message: message,
	}
	
	json.NewEncoder(w).Encode(response)
}

// ProcessCrossChainMessage обрабатывает кросс-чейн сообщение
func (api *SidechainAPI) ProcessCrossChainMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MessageID string `json:"message_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	err := api.manager.ProcessCrossChainMessage(req.MessageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "Cross-chain message processed successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetSidechainStats возвращает статистику sidechain
func (api *SidechainAPI) GetSidechainStats(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}
	
	stats, err := api.manager.GetSidechainStats(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(stats)
}

// GetSidechainBlocks возвращает блоки sidechain
func (api *SidechainAPI) GetSidechainBlocks(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}
	
	sidechain, exists := api.manager.GetSidechain(id)
	if !exists {
		http.Error(w, "sidechain not found", http.StatusNotFound)
		return
	}
	
	// Параметры пагинации
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 100 // По умолчанию
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Применяем пагинацию
	blocks := sidechain.Blocks
	if offset >= len(blocks) {
		blocks = []*SidechainBlock{}
	} else {
		end := offset + limit
		if end > len(blocks) {
			end = len(blocks)
		}
		blocks = blocks[offset:end]
	}
	
	json.NewEncoder(w).Encode(blocks)
}

// GetSidechainTransactions возвращает транзакции sidechain
func (api *SidechainAPI) GetSidechainTransactions(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}
	
	sidechain, exists := api.manager.GetSidechain(id)
	if !exists {
		http.Error(w, "sidechain not found", http.StatusNotFound)
		return
	}
	
	// Собираем все транзакции из всех блоков
	var allTransactions []*SidechainTransaction
	for _, block := range sidechain.Blocks {
		allTransactions = append(allTransactions, block.Transactions...)
	}
	
	// Параметры пагинации
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 100 // По умолчанию
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Применяем пагинацию
	if offset >= len(allTransactions) {
		allTransactions = []*SidechainTransaction{}
	} else {
		end := offset + limit
		if end > len(allTransactions) {
			end = len(allTransactions)
		}
		allTransactions = allTransactions[offset:end]
	}
	
	json.NewEncoder(w).Encode(allTransactions)
}

// UpdateSidechainStatus обновляет статус sidechain
func (api *SidechainAPI) UpdateSidechainStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	sidechain, exists := api.manager.GetSidechain(req.ID)
	if !exists {
		http.Error(w, "sidechain not found", http.StatusNotFound)
		return
	}
	
	// Проверяем валидность статуса
	switch req.Status {
	case "active", "inactive", "paused", "terminated":
		sidechain.Status = SidechainStatus(req.Status)
	default:
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "Sidechain status updated successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// AddValidator добавляет валидатора в sidechain
func (api *SidechainAPI) AddValidator(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        string `json:"id"`
		Validator string `json:"validator"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	sidechain, exists := api.manager.GetSidechain(req.ID)
	if !exists {
		http.Error(w, "sidechain not found", http.StatusNotFound)
		return
	}
	
	// Проверяем, что валидатор еще не добавлен
	for _, validator := range sidechain.Validators {
		if validator == req.Validator {
			http.Error(w, "validator already exists", http.StatusBadRequest)
			return
		}
	}
	
	sidechain.Validators = append(sidechain.Validators, req.Validator)
	
	response := map[string]interface{}{
		"success": true,
		"message": "Validator added successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// RemoveValidator удаляет валидатора из sidechain
func (api *SidechainAPI) RemoveValidator(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        string `json:"id"`
		Validator string `json:"validator"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	sidechain, exists := api.manager.GetSidechain(req.ID)
	if !exists {
		http.Error(w, "sidechain not found", http.StatusNotFound)
		return
	}
	
	// Удаляем валидатора
	for i, validator := range sidechain.Validators {
		if validator == req.Validator {
			sidechain.Validators = append(sidechain.Validators[:i], sidechain.Validators[i+1:]...)
			break
		}
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "Validator removed successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// ExportSidechain экспортирует sidechain
func (api *SidechainAPI) ExportSidechain(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id parameter required", http.StatusBadRequest)
		return
	}
	
	data, err := api.manager.ExportSidechain(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sidechain_%s.json\"", id))
	w.Write(data)
}

// ImportSidechain импортирует sidechain
func (api *SidechainAPI) ImportSidechain(w http.ResponseWriter, r *http.Request) {
	var sidechainData []byte
	if _, err := r.Body.Read(sidechainData); err != nil {
		http.Error(w, "Failed to read sidechain data", http.StatusBadRequest)
		return
	}
	
	sidechain, err := api.manager.ImportSidechain(sidechainData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"id":      sidechain.ID,
		"message": "Sidechain imported successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes регистрирует маршруты API
func (api *SidechainAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/sidechain/create", api.CreateSidechain)
	mux.HandleFunc("/api/sidechain/get", api.GetSidechain)
	mux.HandleFunc("/api/sidechain/list", api.ListSidechains)
	mux.HandleFunc("/api/sidechain/add-block", api.AddBlock)
	mux.HandleFunc("/api/sidechain/add-transaction", api.AddTransaction)
	mux.HandleFunc("/api/sidechain/create-asset", api.CreateAsset)
	mux.HandleFunc("/api/sidechain/get-asset", api.GetAsset)
	mux.HandleFunc("/api/sidechain/list-assets", api.ListAssets)
	mux.HandleFunc("/api/sidechain/create-bridge-tx", api.CreateBridgeTransaction)
	mux.HandleFunc("/api/sidechain/process-bridge-tx", api.ProcessBridgeTransaction)
	mux.HandleFunc("/api/sidechain/get-bridge-tx", api.GetBridgeTransaction)
	mux.HandleFunc("/api/sidechain/list-bridge-txs", api.ListBridgeTransactions)
	mux.HandleFunc("/api/sidechain/send-message", api.SendCrossChainMessage)
	mux.HandleFunc("/api/sidechain/process-message", api.ProcessCrossChainMessage)
	mux.HandleFunc("/api/sidechain/stats", api.GetSidechainStats)
	mux.HandleFunc("/api/sidechain/blocks", api.GetSidechainBlocks)
	mux.HandleFunc("/api/sidechain/transactions", api.GetSidechainTransactions)
	mux.HandleFunc("/api/sidechain/update-status", api.UpdateSidechainStatus)
	mux.HandleFunc("/api/sidechain/add-validator", api.AddValidator)
	mux.HandleFunc("/api/sidechain/remove-validator", api.RemoveValidator)
	mux.HandleFunc("/api/sidechain/export", api.ExportSidechain)
	mux.HandleFunc("/api/sidechain/import", api.ImportSidechain)
}
