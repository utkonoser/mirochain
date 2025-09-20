package statechannel

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// StateChannelAPI предоставляет HTTP API для работы с state channels
type StateChannelAPI struct {
	manager *StateChannelManager
}

// NewStateChannelAPI создает новый API для state channels
func NewStateChannelAPI(manager *StateChannelManager) *StateChannelAPI {
	return &StateChannelAPI{manager: manager}
}

// CreateChannelRequest представляет запрос на создание канала
type CreateChannelRequest struct {
	Participants   []string                 `json:"participants"`
	ChannelType    string                   `json:"channel_type"`
	DisputePeriod  int64                    `json:"dispute_period"` // в секундах
	Metadata       map[string]interface{}   `json:"metadata"`
}

// CreateChannelResponse представляет ответ на создание канала
type CreateChannelResponse struct {
	Success bool   `json:"success"`
	Channel *StateChannel `json:"channel,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DepositRequest представляет запрос на депозит
type DepositRequest struct {
	ChannelID   string `json:"channel_id"`
	Participant string `json:"participant"`
	Amount      string `json:"amount"`
	TxHash      string `json:"tx_hash"`
}

// DepositResponse представляет ответ на депозит
type DepositResponse struct {
	Success bool            `json:"success"`
	Deposit *ChannelDeposit `json:"deposit,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// UpdateStateRequest представляет запрос на обновление состояния
type UpdateStateRequest struct {
	ChannelID    string                 `json:"channel_id"`
	Nonce        int64                  `json:"nonce"`
	Balances     map[string]string      `json:"balances"`
	Participants []string               `json:"participants"`
	Data         map[string]interface{} `json:"data"`
	Signature    string                 `json:"signature"`
	UpdateType   string                 `json:"update_type"`
}

// UpdateStateResponse представляет ответ на обновление состояния
type UpdateStateResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CreateTransactionRequest представляет запрос на создание транзакции
type CreateTransactionRequest struct {
	ChannelID  string                 `json:"channel_id"`
	From       string                 `json:"from"`
	To         string                 `json:"to"`
	Amount     string                 `json:"amount"`
	Data       map[string]interface{} `json:"data"`
	Signature  string                 `json:"signature"`
}

// CreateTransactionResponse представляет ответ на создание транзакции
type CreateTransactionResponse struct {
	Success     bool                `json:"success"`
	Transaction *ChannelTransaction `json:"transaction,omitempty"`
	Error       string              `json:"error,omitempty"`
}

// WithdrawRequest представляет запрос на вывод
type WithdrawRequest struct {
	ChannelID   string `json:"channel_id"`
	Participant string `json:"participant"`
	Amount      string `json:"amount"`
}

// WithdrawResponse представляет ответ на вывод
type WithdrawResponse struct {
	Success    bool              `json:"success"`
	Withdrawal *ChannelWithdrawal `json:"withdrawal,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// ProcessWithdrawalRequest представляет запрос на обработку вывода
type ProcessWithdrawalRequest struct {
	WithdrawalID string `json:"withdrawal_id"`
	TxHash       string `json:"tx_hash"`
}

// ProcessWithdrawalResponse представляет ответ на обработку вывода
type ProcessWithdrawalResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// InitiateDisputeRequest представляет запрос на инициацию спора
type InitiateDisputeRequest struct {
	ChannelID   string        `json:"channel_id"`
	Initiator   string        `json:"initiator"`
	Reason      string        `json:"reason"`
	Evidence    string        `json:"evidence"`
	StateUpdate *StateUpdate  `json:"state_update"`
}

// InitiateDisputeResponse представляет ответ на инициацию спора
type InitiateDisputeResponse struct {
	Success bool           `json:"success"`
	Dispute *ChannelDispute `json:"dispute,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ResolveDisputeRequest представляет запрос на разрешение спора
type ResolveDisputeRequest struct {
	DisputeID  string `json:"dispute_id"`
	Resolution string `json:"resolution"`
}

// ResolveDisputeResponse представляет ответ на разрешение спора
type ResolveDisputeResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CloseChannelRequest представляет запрос на закрытие канала
type CloseChannelRequest struct {
	ChannelID string `json:"channel_id"`
}

// CloseChannelResponse представляет ответ на закрытие канала
type CloseChannelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SettleChannelRequest представляет запрос на урегулирование канала
type SettleChannelRequest struct {
	ChannelID  string       `json:"channel_id"`
	FinalState *StateUpdate `json:"final_state"`
	TxHash     string       `json:"tx_hash"`
}

// SettleChannelResponse представляет ответ на урегулирование канала
type SettleChannelResponse struct {
	Success    bool              `json:"success"`
	Settlement *ChannelSettlement `json:"settlement,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// CreateChannel создает новый state channel
func (api *StateChannelAPI) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим тип канала (по умолчанию payment)
	if req.ChannelType == "" {
		req.ChannelType = "payment"
	}
	
	var channelType ChannelType
	switch req.ChannelType {
	case "payment":
		channelType = TypePayment
	case "micropayment":
		channelType = TypeMicropayment
	case "gaming":
		channelType = TypeGaming
	case "prediction":
		channelType = TypePrediction
	case "custom":
		channelType = TypeCustom
	default:
		http.Error(w, "Invalid channel type", http.StatusBadRequest)
		return
	}

	// Создаем канал
	channel, err := api.manager.CreateChannel(req.Participants, channelType, time.Duration(req.DisputePeriod)*time.Second, req.Metadata)
	if err != nil {
		response := CreateChannelResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateChannelResponse{
		Success: true,
		Channel: channel,
	}

	json.NewEncoder(w).Encode(response)
}

// GetChannel возвращает информацию о канале
func (api *StateChannelAPI) GetChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	channel, err := api.manager.GetChannel(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(channel)
}

// ListChannels возвращает список всех каналов
func (api *StateChannelAPI) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels := api.manager.ListChannels()
	json.NewEncoder(w).Encode(channels)
}

// Deposit добавляет депозит в канал
func (api *StateChannelAPI) Deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositRequest
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

	// Создаем депозит
	deposit, err := api.manager.Deposit(req.ChannelID, req.Participant, amount, req.TxHash)
	if err != nil {
		response := DepositResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := DepositResponse{
		Success: true,
		Deposit: deposit,
	}

	json.NewEncoder(w).Encode(response)
}

// UpdateState обновляет состояние канала
func (api *StateChannelAPI) UpdateState(w http.ResponseWriter, r *http.Request) {
	var req UpdateStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим балансы
	balances := make(map[string]*big.Int)
	for participant, balanceStr := range req.Balances {
		balance, ok := new(big.Int).SetString(balanceStr, 10)
		if !ok {
			http.Error(w, fmt.Sprintf("Invalid balance for participant %s", participant), http.StatusBadRequest)
			return
		}
		balances[participant] = balance
	}

	// Парсим тип обновления
	var updateType UpdateType
	switch req.UpdateType {
	case "payment":
		updateType = UpdatePayment
	case "deposit":
		updateType = UpdateDeposit
	case "withdrawal":
		updateType = UpdateWithdrawal
	case "settlement":
		updateType = UpdateSettlement
	case "dispute":
		updateType = UpdateDispute
	case "close":
		updateType = UpdateClose
	default:
		http.Error(w, "Invalid update type", http.StatusBadRequest)
		return
	}

	// Создаем обновление состояния
	stateUpdate := &StateUpdate{
		ChannelID:    req.ChannelID,
		Nonce:        req.Nonce,
		Balances:     balances,
		Participants: req.Participants,
		Data:         req.Data,
		Signature:    req.Signature,
		Timestamp:    time.Now(),
		UpdateType:   updateType,
	}

	// Обновляем состояние
	err := api.manager.UpdateState(req.ChannelID, stateUpdate)
	if err != nil {
		response := UpdateStateResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := UpdateStateResponse{
		Success: true,
	}

	json.NewEncoder(w).Encode(response)
}

// CreateTransaction создает транзакцию в канале
func (api *StateChannelAPI) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
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

	// Создаем транзакцию
	transaction, err := api.manager.CreateTransaction(req.ChannelID, req.From, req.To, amount, req.Data, req.Signature)
	if err != nil {
		response := CreateTransactionResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateTransactionResponse{
		Success:     true,
		Transaction: transaction,
	}

	json.NewEncoder(w).Encode(response)
}

// Withdraw создает запрос на вывод из канала
func (api *StateChannelAPI) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
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

	// Создаем вывод
	withdrawal, err := api.manager.Withdraw(req.ChannelID, req.Participant, amount)
	if err != nil {
		response := WithdrawResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := WithdrawResponse{
		Success:    true,
		Withdrawal: withdrawal,
	}

	json.NewEncoder(w).Encode(response)
}

// ProcessWithdrawal обрабатывает вывод из канала
func (api *StateChannelAPI) ProcessWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req ProcessWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := api.manager.ProcessWithdrawal(req.WithdrawalID, req.TxHash)
	if err != nil {
		response := ProcessWithdrawalResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ProcessWithdrawalResponse{
		Success: true,
	}

	json.NewEncoder(w).Encode(response)
}

// InitiateDispute инициирует спор по каналу
func (api *StateChannelAPI) InitiateDispute(w http.ResponseWriter, r *http.Request) {
	var req InitiateDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Инициируем спор
	dispute, err := api.manager.InitiateDispute(req.ChannelID, req.Initiator, req.Reason, req.Evidence, req.StateUpdate)
	if err != nil {
		response := InitiateDisputeResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := InitiateDisputeResponse{
		Success: true,
		Dispute: dispute,
	}

	json.NewEncoder(w).Encode(response)
}

// ResolveDispute разрешает спор по каналу
func (api *StateChannelAPI) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	var req ResolveDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := api.manager.ResolveDispute(req.DisputeID, req.Resolution)
	if err != nil {
		response := ResolveDisputeResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ResolveDisputeResponse{
		Success: true,
	}

	json.NewEncoder(w).Encode(response)
}

// CloseChannel закрывает канал
func (api *StateChannelAPI) CloseChannel(w http.ResponseWriter, r *http.Request) {
	var req CloseChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := api.manager.CloseChannel(req.ChannelID)
	if err != nil {
		response := CloseChannelResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CloseChannelResponse{
		Success: true,
	}

	json.NewEncoder(w).Encode(response)
}

// SettleChannel урегулирует канал
func (api *StateChannelAPI) SettleChannel(w http.ResponseWriter, r *http.Request) {
	var req SettleChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Урегулируем канал
	settlement, err := api.manager.SettleChannel(req.ChannelID, req.FinalState, req.TxHash)
	if err != nil {
		response := SettleChannelResponse{
			Success: false,
			Error:   err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := SettleChannelResponse{
		Success:    true,
		Settlement: settlement,
	}

	json.NewEncoder(w).Encode(response)
}

// GetChannelStats возвращает статистику канала
func (api *StateChannelAPI) GetChannelStats(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	stats, err := api.manager.GetChannelStats(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// GetChannelTransactions возвращает транзакции канала
func (api *StateChannelAPI) GetChannelTransactions(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	transactions, err := api.manager.GetChannelTransactions(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

// GetChannelDeposits возвращает депозиты канала
func (api *StateChannelAPI) GetChannelDeposits(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	deposits, err := api.manager.GetChannelDeposits(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(deposits)
}

// GetChannelWithdrawals возвращает выводы канала
func (api *StateChannelAPI) GetChannelWithdrawals(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	withdrawals, err := api.manager.GetChannelWithdrawals(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(withdrawals)
}

// GetChannelDisputes возвращает споры канала
func (api *StateChannelAPI) GetChannelDisputes(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	disputes, err := api.manager.GetChannelDisputes(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(disputes)
}

// ExportChannel экспортирует канал
func (api *StateChannelAPI) ExportChannel(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	data, err := api.manager.ExportChannel(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"channel_%s.json\"", channelID))
	w.Write(data)
}

// ImportChannel импортирует канал
func (api *StateChannelAPI) ImportChannel(w http.ResponseWriter, r *http.Request) {
	var channelData []byte
	if _, err := r.Body.Read(channelData); err != nil {
		http.Error(w, "Failed to read channel data", http.StatusBadRequest)
		return
	}

	channel, err := api.manager.ImportChannel(channelData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"id":      channel.ID,
		"message": "Channel imported successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetChannelBalance возвращает баланс участника в канале
func (api *StateChannelAPI) GetChannelBalance(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	participant := r.URL.Query().Get("participant")

	if channelID == "" || participant == "" {
		http.Error(w, "channel_id and participant parameters required", http.StatusBadRequest)
		return
	}

	channel, err := api.manager.GetChannel(channelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	balance, exists := channel.Balances[participant]
	if !exists {
		http.Error(w, "Participant not found in channel", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"channel_id":  channelID,
		"participant": participant,
		"balance":     balance.String(),
	}

	json.NewEncoder(w).Encode(response)
}

// GetChannelHistory возвращает историю канала
func (api *StateChannelAPI) GetChannelHistory(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id parameter required", http.StatusBadRequest)
		return
	}

	// Получаем все данные канала
	transactions, _ := api.manager.GetChannelTransactions(channelID)
	deposits, _ := api.manager.GetChannelDeposits(channelID)
	withdrawals, _ := api.manager.GetChannelWithdrawals(channelID)
	disputes, _ := api.manager.GetChannelDisputes(channelID)

	// Объединяем все события
	var events []map[string]interface{}

	// Добавляем транзакции
	for _, tx := range transactions {
		events = append(events, map[string]interface{}{
			"type":      "transaction",
			"id":        tx.ID,
			"timestamp": tx.Timestamp,
			"data":      tx,
		})
	}

	// Добавляем депозиты
	for _, deposit := range deposits {
		events = append(events, map[string]interface{}{
			"type":      "deposit",
			"id":        deposit.ID,
			"timestamp": deposit.Timestamp,
			"data":      deposit,
		})
	}

	// Добавляем выводы
	for _, withdrawal := range withdrawals {
		events = append(events, map[string]interface{}{
			"type":      "withdrawal",
			"id":        withdrawal.ID,
			"timestamp": withdrawal.Timestamp,
			"data":      withdrawal,
		})
	}

	// Добавляем споры
	for _, dispute := range disputes {
		events = append(events, map[string]interface{}{
			"type":      "dispute",
			"id":        dispute.ID,
			"timestamp": dispute.Timestamp,
			"data":      dispute,
		})
	}

	// Сортируем по времени
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i]["timestamp"].(time.Time).After(events[j]["timestamp"].(time.Time)) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	json.NewEncoder(w).Encode(events)
}

// RegisterRoutes регистрирует маршруты API
func (api *StateChannelAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/statechannel/create", api.CreateChannel)
	mux.HandleFunc("/api/statechannel/get", api.GetChannel)
	mux.HandleFunc("/api/statechannel/list", api.ListChannels)
	mux.HandleFunc("/api/statechannel/deposit", api.Deposit)
	mux.HandleFunc("/api/statechannel/update-state", api.UpdateState)
	mux.HandleFunc("/api/statechannel/create-transaction", api.CreateTransaction)
	mux.HandleFunc("/api/statechannel/withdraw", api.Withdraw)
	mux.HandleFunc("/api/statechannel/process-withdrawal", api.ProcessWithdrawal)
	mux.HandleFunc("/api/statechannel/initiate-dispute", api.InitiateDispute)
	mux.HandleFunc("/api/statechannel/resolve-dispute", api.ResolveDispute)
	mux.HandleFunc("/api/statechannel/close", api.CloseChannel)
	mux.HandleFunc("/api/statechannel/settle", api.SettleChannel)
	mux.HandleFunc("/api/statechannel/stats", api.GetChannelStats)
	mux.HandleFunc("/api/statechannel/transactions", api.GetChannelTransactions)
	mux.HandleFunc("/api/statechannel/deposits", api.GetChannelDeposits)
	mux.HandleFunc("/api/statechannel/withdrawals", api.GetChannelWithdrawals)
	mux.HandleFunc("/api/statechannel/disputes", api.GetChannelDisputes)
	mux.HandleFunc("/api/statechannel/balance", api.GetChannelBalance)
	mux.HandleFunc("/api/statechannel/history", api.GetChannelHistory)
	mux.HandleFunc("/api/statechannel/export", api.ExportChannel)
	mux.HandleFunc("/api/statechannel/import", api.ImportChannel)
}
