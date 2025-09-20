package tokens

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

// TokenAPI предоставляет HTTP API для работы с токенами
type TokenAPI struct {
	manager *ERC20Manager
}

// NewTokenAPI создает новый API для токенов
func NewTokenAPI(manager *ERC20Manager) *TokenAPI {
	return &TokenAPI{manager: manager}
}

// CreateTokenRequest представляет запрос на создание токена
type CreateTokenRequest struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	TotalSupply string `json:"total_supply"`
	Owner       string `json:"owner"`
}

// CreateTokenResponse представляет ответ на создание токена
type CreateTokenResponse struct {
	Success bool   `json:"success"`
	Address string `json:"address,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TransferRequest представляет запрос на перевод токенов
type TransferRequest struct {
	TokenAddress string `json:"token_address"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
}

// TransferResponse представляет ответ на перевод токенов
type TransferResponse struct {
	Success bool                `json:"success"`
	Event   *ERC20TransferEvent `json:"event,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// ApproveRequest представляет запрос на одобрение токенов
type ApproveRequest struct {
	TokenAddress string `json:"token_address"`
	Owner        string `json:"owner"`
	Spender      string `json:"spender"`
	Amount       string `json:"amount"`
}

// ApproveResponse представляет ответ на одобрение токенов
type ApproveResponse struct {
	Success bool                `json:"success"`
	Event   *ERC20ApprovalEvent `json:"event,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// BalanceRequest представляет запрос на получение баланса
type BalanceRequest struct {
	TokenAddress string `json:"token_address"`
	Address      string `json:"address"`
}

// BalanceResponse представляет ответ на получение баланса
type BalanceResponse struct {
	Success bool   `json:"success"`
	Balance string `json:"balance,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreateToken создает новый токен
func (api *TokenAPI) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим общее предложение
	totalSupply, ok := new(big.Int).SetString(req.TotalSupply, 10)
	if !ok {
		http.Error(w, "Invalid total supply", http.StatusBadRequest)
		return
	}

	// Создаем токен
	token, err := api.manager.CreateToken(req.Name, req.Symbol, req.Decimals, totalSupply, req.Owner)
	if err != nil {
		response := CreateTokenResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateTokenResponse{
		Success: true,
		Address: token.Address,
	}

	json.NewEncoder(w).Encode(response)
}

// Transfer переводит токены
func (api *TokenAPI) Transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим количество
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Выполняем перевод
	event, err := api.manager.Transfer(req.TokenAddress, req.From, req.To, amount)
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

// Approve одобряет расход токенов
func (api *TokenAPI) Approve(w http.ResponseWriter, r *http.Request) {
	var req ApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим количество
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Выполняем одобрение
	event, err := api.manager.Approve(req.TokenAddress, req.Owner, req.Spender, amount)
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

// TransferFrom переводит токены от имени владельца
func (api *TokenAPI) TransferFrom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddress string `json:"token_address"`
		Spender      string `json:"spender"`
		From         string `json:"from"`
		To           string `json:"to"`
		Amount       string `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим количество
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Выполняем перевод от имени
	event, err := api.manager.TransferFrom(req.TokenAddress, req.Spender, req.From, req.To, amount)
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

// GetBalance возвращает баланс токенов
func (api *TokenAPI) GetBalance(w http.ResponseWriter, r *http.Request) {
	tokenAddress := r.URL.Query().Get("token_address")
	address := r.URL.Query().Get("address")

	if tokenAddress == "" || address == "" {
		http.Error(w, "token_address and address parameters required", http.StatusBadRequest)
		return
	}

	balance, err := api.manager.BalanceOf(tokenAddress, address)
	if err != nil {
		response := BalanceResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := BalanceResponse{
		Success: true,
		Balance: balance.String(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetAllowance возвращает разрешение на расход токенов
func (api *TokenAPI) GetAllowance(w http.ResponseWriter, r *http.Request) {
	tokenAddress := r.URL.Query().Get("token_address")
	owner := r.URL.Query().Get("owner")
	spender := r.URL.Query().Get("spender")

	if tokenAddress == "" || owner == "" || spender == "" {
		http.Error(w, "token_address, owner and spender parameters required", http.StatusBadRequest)
		return
	}

	allowance, err := api.manager.Allowance(tokenAddress, owner, spender)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"allowance": allowance.String(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetTokenInfo возвращает информацию о токене
func (api *TokenAPI) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}

	info, err := api.manager.GetTokenInfo(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(info)
}

// ListTokens возвращает список всех токенов
func (api *TokenAPI) ListTokens(w http.ResponseWriter, r *http.Request) {
	tokens := api.manager.ListTokens()

	// Конвертируем в формат для API
	tokenList := make([]map[string]interface{}, 0, len(tokens))
	for _, token := range tokens {
		info, _ := api.manager.GetTokenInfo(token.Address)
		tokenList = append(tokenList, info)
	}

	json.NewEncoder(w).Encode(tokenList)
}

// GetTokenStats возвращает статистику токена
func (api *TokenAPI) GetTokenStats(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}

	stats, err := api.manager.GetTokenStats(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// SearchTokens ищет токены по критериям
func (api *TokenAPI) SearchTokens(w http.ResponseWriter, r *http.Request) {
	var criteria map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Конвертируем строковые значения в big.Int где необходимо
	if minSupplyStr, ok := criteria["min_supply"].(string); ok {
		if minSupply, ok := new(big.Int).SetString(minSupplyStr, 10); ok {
			criteria["min_supply"] = minSupply
		} else {
			http.Error(w, "Invalid min_supply format", http.StatusBadRequest)
			return
		}
	}

	tokens := api.manager.SearchTokens(criteria)

	// Конвертируем в формат для API
	tokenList := make([]map[string]interface{}, 0, len(tokens))
	for _, token := range tokens {
		info, _ := api.manager.GetTokenInfo(token.Address)
		tokenList = append(tokenList, info)
	}

	json.NewEncoder(w).Encode(tokenList)
}

// Mint создает новые токены
func (api *TokenAPI) Mint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddress string `json:"token_address"`
		To           string `json:"to"`
		Amount       string `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим количество
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Создаем токены
	err := api.manager.Mint(req.TokenAddress, req.To, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Tokens minted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// Burn сжигает токены
func (api *TokenAPI) Burn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddress string `json:"token_address"`
		From         string `json:"from"`
		Amount       string `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим количество
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Сжигаем токены
	err := api.manager.Burn(req.TokenAddress, req.From, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Tokens burned successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// ExportToken экспортирует токен
func (api *TokenAPI) ExportToken(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}

	data, err := api.manager.ExportToken(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"token_%s.json\"", address))
	w.Write(data)
}

// ImportToken импортирует токен
func (api *TokenAPI) ImportToken(w http.ResponseWriter, r *http.Request) {
	var tokenData []byte
	if _, err := r.Body.Read(tokenData); err != nil {
		http.Error(w, "Failed to read token data", http.StatusBadRequest)
		return
	}

	token, err := api.manager.ImportToken(tokenData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"address": token.Address,
		"message": "Token imported successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetTotalSupply возвращает общее предложение токена
func (api *TokenAPI) GetTotalSupply(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}

	supply, err := api.manager.GetTotalSupply(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"total_supply": supply.String(),
	}

	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes регистрирует маршруты API
func (api *TokenAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/tokens/create", api.CreateToken)
	mux.HandleFunc("/api/tokens/transfer", api.Transfer)
	mux.HandleFunc("/api/tokens/approve", api.Approve)
	mux.HandleFunc("/api/tokens/transfer-from", api.TransferFrom)
	mux.HandleFunc("/api/tokens/balance", api.GetBalance)
	mux.HandleFunc("/api/tokens/allowance", api.GetAllowance)
	mux.HandleFunc("/api/tokens/info", api.GetTokenInfo)
	mux.HandleFunc("/api/tokens/list", api.ListTokens)
	mux.HandleFunc("/api/tokens/stats", api.GetTokenStats)
	mux.HandleFunc("/api/tokens/search", api.SearchTokens)
	mux.HandleFunc("/api/tokens/mint", api.Mint)
	mux.HandleFunc("/api/tokens/burn", api.Burn)
	mux.HandleFunc("/api/tokens/export", api.ExportToken)
	mux.HandleFunc("/api/tokens/import", api.ImportToken)
	mux.HandleFunc("/api/tokens/total-supply", api.GetTotalSupply)
}
