package nft

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

// NFTAPI предоставляет HTTP API для работы с NFT
type NFTAPI struct {
	manager *ERC721Manager
}

// NewNFTAPI создает новый API для NFT
func NewNFTAPI(manager *ERC721Manager) *NFTAPI {
	return &NFTAPI{manager: manager}
}

// CreateContractRequest представляет запрос на создание контракта
type CreateContractRequest struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Owner     string `json:"owner"`
	BaseURI   string `json:"base_uri"`
	MaxSupply string `json:"max_supply,omitempty"`
}

// CreateContractResponse представляет ответ на создание контракта
type CreateContractResponse struct {
	Success bool   `json:"success"`
	Address string `json:"address,omitempty"`
	Error   string `json:"error,omitempty"`
}

// MintRequest представляет запрос на создание NFT
type MintRequest struct {
	ContractAddress string                 `json:"contract_address"`
	To              string                 `json:"to"`
	TokenID         string                 `json:"token_id"`
	Metadata        *TokenMetadata         `json:"metadata"`
	Attributes      map[string]interface{} `json:"attributes,omitempty"`
}

// MintResponse представляет ответ на создание NFT
type MintResponse struct {
	Success bool         `json:"success"`
	Token   *ERC721Token `json:"token,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// TransferRequest представляет запрос на перевод NFT
type TransferRequest struct {
	ContractAddress string `json:"contract_address"`
	From            string `json:"from"`
	To              string `json:"to"`
	TokenID         string `json:"token_id"`
}

// TransferResponse представляет ответ на перевод NFT
type TransferResponse struct {
	Success bool           `json:"success"`
	Event   *TransferEvent `json:"event,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ApproveRequest представляет запрос на одобрение NFT
type ApproveRequest struct {
	ContractAddress string `json:"contract_address"`
	Owner           string `json:"owner"`
	Approved        string `json:"approved"`
	TokenID         string `json:"token_id"`
}

// ApproveResponse представляет ответ на одобрение NFT
type ApproveResponse struct {
	Success bool         `json:"success"`
	Event   *ApprovalEvent `json:"event,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// CreateContract создает новый NFT контракт
func (api *NFTAPI) CreateContract(w http.ResponseWriter, r *http.Request) {
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	var maxSupply *big.Int
	if req.MaxSupply != "" {
		var ok bool
		maxSupply, ok = new(big.Int).SetString(req.MaxSupply, 10)
		if !ok {
			http.Error(w, "Invalid max supply", http.StatusBadRequest)
			return
		}
	}
	
	// Создаем контракт
	contract, err := api.manager.CreateContract(req.Name, req.Symbol, req.Owner, req.BaseURI, maxSupply)
	if err != nil {
		response := CreateContractResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := CreateContractResponse{
		Success: true,
		Address: contract.Address,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Mint создает новый NFT токен
func (api *NFTAPI) Mint(w http.ResponseWriter, r *http.Request) {
	var req MintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenID, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	// Создаем NFT
	token, err := api.manager.Mint(req.ContractAddress, req.To, tokenID, req.Metadata, req.Attributes)
	if err != nil {
		response := MintResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := MintResponse{
		Success: true,
		Token:   token,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Transfer переводит NFT
func (api *NFTAPI) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenID, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	// Выполняем перевод
	event, err := api.manager.Transfer(req.ContractAddress, req.From, req.To, tokenID)
	if err != nil {
		response := TransferResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := TransferResponse{
		Success: true,
		Event:   event,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Approve одобряет адрес для управления NFT
func (api *NFTAPI) Approve(w http.ResponseWriter, r *http.Request) {
	var req ApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenID, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	// Выполняем одобрение
	event, err := api.manager.Approve(req.ContractAddress, req.Owner, req.Approved, tokenID)
	if err != nil {
		response := ApproveResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := ApproveResponse{
		Success: true,
		Event:   event,
	}
	
	json.NewEncoder(w).Encode(response)
}

// SetApprovalForAll одобряет или отзывает одобрение для всех токенов
func (api *NFTAPI) SetApprovalForAll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContractAddress string `json:"contract_address"`
		Owner           string `json:"owner"`
		Operator        string `json:"operator"`
		Approved        bool   `json:"approved"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Выполняем одобрение
	event, err := api.manager.SetApprovalForAll(req.ContractAddress, req.Owner, req.Operator, req.Approved)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"event":   event,
	}
	
	json.NewEncoder(w).Encode(response)
}

// TransferFrom переводит NFT от имени владельца
func (api *NFTAPI) TransferFrom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContractAddress string `json:"contract_address"`
		Spender         string `json:"spender"`
		From            string `json:"from"`
		To              string `json:"to"`
		TokenID         string `json:"token_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenID, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	// Выполняем перевод
	event, err := api.manager.TransferFrom(req.ContractAddress, req.Spender, req.From, req.To, tokenID)
	if err != nil {
		response := TransferResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := TransferResponse{
		Success: true,
		Event:   event,
	}
	
	json.NewEncoder(w).Encode(response)
}

// OwnerOf возвращает владельца NFT
func (api *NFTAPI) OwnerOf(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contract_address")
	tokenID := r.URL.Query().Get("token_id")
	
	if contractAddress == "" || tokenID == "" {
		http.Error(w, "contract_address and token_id parameters required", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenIDBig, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	owner, err := api.manager.OwnerOf(contractAddress, tokenIDBig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"owner":   owner,
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetApproved возвращает одобренный адрес для NFT
func (api *NFTAPI) GetApproved(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contract_address")
	tokenID := r.URL.Query().Get("token_id")
	
	if contractAddress == "" || tokenID == "" {
		http.Error(w, "contract_address and token_id parameters required", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenIDBig, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	approved, err := api.manager.GetApproved(contractAddress, tokenIDBig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success":  true,
		"approved": approved,
	}
	
	json.NewEncoder(w).Encode(response)
}

// IsApprovedForAll проверяет одобрение оператора
func (api *NFTAPI) IsApprovedForAll(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contract_address")
	owner := r.URL.Query().Get("owner")
	operator := r.URL.Query().Get("operator")
	
	if contractAddress == "" || owner == "" || operator == "" {
		http.Error(w, "contract_address, owner and operator parameters required", http.StatusBadRequest)
		return
	}
	
	approved, err := api.manager.IsApprovedForAll(contractAddress, owner, operator)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success":  true,
		"approved": approved,
	}
	
	json.NewEncoder(w).Encode(response)
}

// BalanceOf возвращает количество NFT у владельца
func (api *NFTAPI) BalanceOf(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contract_address")
	owner := r.URL.Query().Get("owner")
	
	if contractAddress == "" || owner == "" {
		http.Error(w, "contract_address and owner parameters required", http.StatusBadRequest)
		return
	}
	
	balance, err := api.manager.BalanceOf(contractAddress, owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"balance": balance.String(),
	}
	
	json.NewEncoder(w).Encode(response)
}

// GetToken возвращает информацию о NFT
func (api *NFTAPI) GetToken(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contract_address")
	tokenID := r.URL.Query().Get("token_id")
	
	if contractAddress == "" || tokenID == "" {
		http.Error(w, "contract_address and token_id parameters required", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenIDBig, ok := new(big.Int).SetString(tokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	token, err := api.manager.GetToken(contractAddress, tokenIDBig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	json.NewEncoder(w).Encode(token)
}

// GetTokensByOwner возвращает все NFT владельца
func (api *NFTAPI) GetTokensByOwner(w http.ResponseWriter, r *http.Request) {
	contractAddress := r.URL.Query().Get("contract_address")
	owner := r.URL.Query().Get("owner")
	
	if contractAddress == "" || owner == "" {
		http.Error(w, "contract_address and owner parameters required", http.StatusBadRequest)
		return
	}
	
	tokens, err := api.manager.GetTokensByOwner(contractAddress, owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	json.NewEncoder(w).Encode(tokens)
}

// GetContractInfo возвращает информацию о контракте
func (api *NFTAPI) GetContractInfo(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}
	
	info, err := api.manager.GetContractInfo(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(info)
}

// ListContracts возвращает список всех контрактов
func (api *NFTAPI) ListContracts(w http.ResponseWriter, r *http.Request) {
	contracts := api.manager.ListContracts()
	
	// Конвертируем в формат для API
	contractList := make([]map[string]interface{}, 0, len(contracts))
	for _, contract := range contracts {
		info, _ := api.manager.GetContractInfo(contract.Address)
		contractList = append(contractList, info)
	}
	
	json.NewEncoder(w).Encode(contractList)
}

// GetContractStats возвращает статистику контракта
func (api *NFTAPI) GetContractStats(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}
	
	stats, err := api.manager.GetContractStats(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(stats)
}

// SearchTokens ищет NFT по критериям
func (api *NFTAPI) SearchTokens(w http.ResponseWriter, r *http.Request) {
	var criteria map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	tokens, err := api.manager.SearchTokens(criteria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	json.NewEncoder(w).Encode(tokens)
}

// Burn сжигает NFT
func (api *NFTAPI) Burn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ContractAddress string `json:"contract_address"`
		Owner           string `json:"owner"`
		TokenID         string `json:"token_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Парсим TokenID
	tokenID, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		http.Error(w, "Invalid token ID", http.StatusBadRequest)
		return
	}
	
	// Сжигаем NFT
	err := api.manager.Burn(req.ContractAddress, req.Owner, tokenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"message": "NFT burned successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// ExportContract экспортирует контракт
func (api *NFTAPI) ExportContract(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}
	
	data, err := api.manager.ExportContract(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"nft_contract_%s.json\"", address))
	w.Write(data)
}

// ImportContract импортирует контракт
func (api *NFTAPI) ImportContract(w http.ResponseWriter, r *http.Request) {
	var contractData []byte
	if _, err := r.Body.Read(contractData); err != nil {
		http.Error(w, "Failed to read contract data", http.StatusBadRequest)
		return
	}
	
	contract, err := api.manager.ImportContract(contractData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"address": contract.Address,
		"message": "Contract imported successfully",
	}
	
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes регистрирует маршруты API
func (api *NFTAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/nft/create-contract", api.CreateContract)
	mux.HandleFunc("/api/nft/mint", api.Mint)
	mux.HandleFunc("/api/nft/transfer", api.Transfer)
	mux.HandleFunc("/api/nft/approve", api.Approve)
	mux.HandleFunc("/api/nft/set-approval-for-all", api.SetApprovalForAll)
	mux.HandleFunc("/api/nft/transfer-from", api.TransferFrom)
	mux.HandleFunc("/api/nft/owner-of", api.OwnerOf)
	mux.HandleFunc("/api/nft/get-approved", api.GetApproved)
	mux.HandleFunc("/api/nft/is-approved-for-all", api.IsApprovedForAll)
	mux.HandleFunc("/api/nft/balance-of", api.BalanceOf)
	mux.HandleFunc("/api/nft/get-token", api.GetToken)
	mux.HandleFunc("/api/nft/get-tokens-by-owner", api.GetTokensByOwner)
	mux.HandleFunc("/api/nft/contract-info", api.GetContractInfo)
	mux.HandleFunc("/api/nft/list-contracts", api.ListContracts)
	mux.HandleFunc("/api/nft/contract-stats", api.GetContractStats)
	mux.HandleFunc("/api/nft/search", api.SearchTokens)
	mux.HandleFunc("/api/nft/burn", api.Burn)
	mux.HandleFunc("/api/nft/export", api.ExportContract)
	mux.HandleFunc("/api/nft/import", api.ImportContract)
}
