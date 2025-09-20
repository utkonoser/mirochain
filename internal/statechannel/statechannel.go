package statechannel

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// StateChannel представляет канал состояния между участниками
type StateChannel struct {
	ID              string                 `json:"id"`
	Participants    []string               `json:"participants"`
	Balances        map[string]*big.Int    `json:"balances"`
	Nonce           int64                  `json:"nonce"`
	State           ChannelState           `json:"state"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	DisputePeriod   time.Duration          `json:"dispute_period"`
	TotalDeposit    *big.Int               `json:"total_deposit"`
	ChannelType     ChannelType            `json:"channel_type"`
	ContractAddress string                 `json:"contract_address,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ChannelState представляет состояние канала
type ChannelState string

const (
	StateOpening     ChannelState = "opening"
	StateOpen        ChannelState = "open"
	StateClosing     ChannelState = "closing"
	StateClosed      ChannelState = "closed"
	StateDisputed    ChannelState = "disputed"
	StateSettled     ChannelState = "settled"
)

// ChannelType представляет тип канала
type ChannelType string

const (
	TypePayment     ChannelType = "payment"
	TypeMicropayment ChannelType = "micropayment"
	TypeGaming      ChannelType = "gaming"
	TypePrediction  ChannelType = "prediction"
	TypeCustom      ChannelType = "custom"
)

// StateUpdate представляет обновление состояния канала
type StateUpdate struct {
	ChannelID    string                 `json:"channel_id"`
	Nonce        int64                  `json:"nonce"`
	Balances     map[string]*big.Int    `json:"balances"`
	Participants []string               `json:"participants"`
	Data         map[string]interface{} `json:"data"`
	Signature    string                 `json:"signature"`
	Timestamp    time.Time              `json:"timestamp"`
	UpdateType   UpdateType             `json:"update_type"`
}

// UpdateType представляет тип обновления
type UpdateType string

const (
	UpdatePayment     UpdateType = "payment"
	UpdateDeposit     UpdateType = "deposit"
	UpdateWithdrawal  UpdateType = "withdrawal"
	UpdateSettlement  UpdateType = "settlement"
	UpdateDispute     UpdateType = "dispute"
	UpdateClose       UpdateType = "close"
)

// ChannelTransaction представляет транзакцию в канале
type ChannelTransaction struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channel_id"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Amount      *big.Int               `json:"amount"`
	Nonce       int64                  `json:"nonce"`
	Data        map[string]interface{} `json:"data"`
	Signature   string                 `json:"signature"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      TransactionStatus      `json:"status"`
	GasUsed     int64                  `json:"gas_used"`
	GasPrice    *big.Int               `json:"gas_price"`
}

// TransactionStatus представляет статус транзакции
type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusConfirmed TransactionStatus = "confirmed"
	StatusFailed    TransactionStatus = "failed"
	StatusReverted  TransactionStatus = "reverted"
)

// ChannelDeposit представляет депозит в канал
type ChannelDeposit struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Participant string  `json:"participant"`
	Amount    *big.Int  `json:"amount"`
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// ChannelWithdrawal представляет вывод из канала
type ChannelWithdrawal struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Participant string  `json:"participant"`
	Amount    *big.Int  `json:"amount"`
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// ChannelDispute представляет спор по каналу
type ChannelDispute struct {
	ID           string    `json:"id"`
	ChannelID    string    `json:"channel_id"`
	Initiator    string    `json:"initiator"`
	Reason       string    `json:"reason"`
	Evidence     string    `json:"evidence"`
	StateUpdate  *StateUpdate `json:"state_update"`
	Timestamp    time.Time `json:"timestamp"`
	Status       string    `json:"status"`
	Resolution   string    `json:"resolution,omitempty"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

// ChannelSettlement представляет урегулирование канала
type ChannelSettlement struct {
	ID           string                 `json:"id"`
	ChannelID    string                 `json:"channel_id"`
	FinalState   *StateUpdate           `json:"final_state"`
	TxHash       string                 `json:"tx_hash"`
	Timestamp    time.Time              `json:"timestamp"`
	GasUsed      int64                  `json:"gas_used"`
	GasPrice     *big.Int               `json:"gas_price"`
	Participants []string               `json:"participants"`
	Balances     map[string]*big.Int    `json:"balances"`
}

// StateChannelManager управляет state channels
type StateChannelManager struct {
	channels     map[string]*StateChannel
	transactions map[string]*ChannelTransaction
	deposits     map[string]*ChannelDeposit
	withdrawals  map[string]*ChannelWithdrawal
	disputes     map[string]*ChannelDispute
	settlements  map[string]*ChannelSettlement
	mu           sync.RWMutex
}

// NewStateChannelManager создает новый менеджер state channels
func NewStateChannelManager() *StateChannelManager {
	return &StateChannelManager{
		channels:     make(map[string]*StateChannel),
		transactions: make(map[string]*ChannelTransaction),
		deposits:     make(map[string]*ChannelDeposit),
		withdrawals:  make(map[string]*ChannelWithdrawal),
		disputes:     make(map[string]*ChannelDispute),
		settlements:  make(map[string]*ChannelSettlement),
	}
}

// CreateChannel создает новый state channel
func (scm *StateChannelManager) CreateChannel(participants []string, channelType ChannelType, disputePeriod time.Duration, metadata map[string]interface{}) (*StateChannel, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	if len(participants) < 2 {
		return nil, fmt.Errorf("channel must have at least 2 participants")
	}

	// Генерируем уникальный ID
	channelID := scm.generateChannelID(participants)

	// Инициализируем балансы
	balances := make(map[string]*big.Int)
	for _, participant := range participants {
		balances[participant] = big.NewInt(0)
	}

	channel := &StateChannel{
		ID:            channelID,
		Participants:  participants,
		Balances:      balances,
		Nonce:         0,
		State:         StateOpening,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(disputePeriod),
		DisputePeriod: disputePeriod,
		TotalDeposit:  big.NewInt(0),
		ChannelType:   channelType,
		Metadata:      metadata,
	}

	scm.channels[channelID] = channel
	return channel, nil
}

// GetChannel возвращает канал по ID
func (scm *StateChannelManager) GetChannel(channelID string) (*StateChannel, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}
	return channel, nil
}

// ListChannels возвращает список всех каналов
func (scm *StateChannelManager) ListChannels() []*StateChannel {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	channels := make([]*StateChannel, 0, len(scm.channels))
	for _, channel := range scm.channels {
		channels = append(channels, channel)
	}
	return channels
}

// Deposit добавляет депозит в канал
func (scm *StateChannelManager) Deposit(channelID, participant string, amount *big.Int, txHash string) (*ChannelDeposit, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	if channel.State != StateOpening && channel.State != StateOpen {
		return nil, fmt.Errorf("channel is not in a state that accepts deposits")
	}

	// Проверяем, что участник является частью канала
	validParticipant := false
	for _, p := range channel.Participants {
		if p == participant {
			validParticipant = true
			break
		}
	}
	if !validParticipant {
		return nil, fmt.Errorf("participant %s is not part of channel %s", participant, channelID)
	}

	// Обновляем баланс
	channel.Balances[participant].Add(channel.Balances[participant], amount)
	channel.TotalDeposit.Add(channel.TotalDeposit, amount)
	channel.UpdatedAt = time.Now()

	// Создаем запись о депозите
	depositID := scm.generateDepositID(channelID, participant)
	deposit := &ChannelDeposit{
		ID:          depositID,
		ChannelID:   channelID,
		Participant: participant,
		Amount:      new(big.Int).Set(amount),
		TxHash:      txHash,
		Timestamp:   time.Now(),
		Status:      "confirmed",
	}

	scm.deposits[depositID] = deposit

	// Если канал в состоянии opening и все участники внесли депозиты, переводим в open
	if channel.State == StateOpening {
		allDeposited := true
		for _, p := range channel.Participants {
			if channel.Balances[p].Cmp(big.NewInt(0)) == 0 {
				allDeposited = false
				break
			}
		}
		if allDeposited {
			channel.State = StateOpen
		}
	}

	return deposit, nil
}

// UpdateState обновляет состояние канала
func (scm *StateChannelManager) UpdateState(channelID string, stateUpdate *StateUpdate) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	if channel.State != StateOpen {
		return fmt.Errorf("channel is not open, cannot update state")
	}

	// Проверяем nonce
	if stateUpdate.Nonce <= channel.Nonce {
		return fmt.Errorf("invalid nonce: expected > %d, got %d", channel.Nonce, stateUpdate.Nonce)
	}

	// Проверяем подпись (упрощенная проверка)
	if stateUpdate.Signature == "" {
		return fmt.Errorf("state update must be signed")
	}

	// Обновляем состояние
	channel.Nonce = stateUpdate.Nonce
	channel.Balances = stateUpdate.Balances
	channel.UpdatedAt = time.Now()

	// Создаем транзакцию
	txID := scm.generateTransactionID(channelID)
	tx := &ChannelTransaction{
		ID:        txID,
		ChannelID: channelID,
		From:      stateUpdate.Participants[0], // Упрощение
		To:        stateUpdate.Participants[1], // Упрощение
		Amount:    big.NewInt(0), // Будет рассчитано из изменений балансов
		Nonce:     stateUpdate.Nonce,
		Data:      stateUpdate.Data,
		Signature: stateUpdate.Signature,
		Timestamp: stateUpdate.Timestamp,
		Status:    StatusConfirmed,
		GasUsed:   21000, // Базовый газ
		GasPrice:  big.NewInt(20),
	}

	scm.transactions[txID] = tx

	return nil
}

// CreateTransaction создает транзакцию в канале
func (scm *StateChannelManager) CreateTransaction(channelID, from, to string, amount *big.Int, data map[string]interface{}, signature string) (*ChannelTransaction, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	if channel.State != StateOpen {
		return nil, fmt.Errorf("channel is not open, cannot create transaction")
	}

	// Проверяем баланс отправителя
	if channel.Balances[from].Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient balance: %s < %s", channel.Balances[from].String(), amount.String())
	}

	// Создаем транзакцию
	txID := scm.generateTransactionID(channelID)
	tx := &ChannelTransaction{
		ID:        txID,
		ChannelID: channelID,
		From:      from,
		To:        to,
		Amount:    new(big.Int).Set(amount),
		Nonce:     channel.Nonce + 1,
		Data:      data,
		Signature: signature,
		Timestamp: time.Now(),
		Status:    StatusPending,
		GasUsed:   21000,
		GasPrice:  big.NewInt(20),
	}

	scm.transactions[txID] = tx

	// Обновляем балансы
	channel.Balances[from].Sub(channel.Balances[from], amount)
	channel.Balances[to].Add(channel.Balances[to], amount)
	channel.Nonce++
	channel.UpdatedAt = time.Now()

	tx.Status = StatusConfirmed

	return tx, nil
}

// Withdraw создает запрос на вывод из канала
func (scm *StateChannelManager) Withdraw(channelID, participant string, amount *big.Int) (*ChannelWithdrawal, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	if channel.State != StateOpen && channel.State != StateClosing {
		return nil, fmt.Errorf("channel is not in a state that allows withdrawals")
	}

	// Проверяем баланс
	if channel.Balances[participant].Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient balance for withdrawal")
	}

	// Создаем запрос на вывод
	withdrawalID := scm.generateWithdrawalID(channelID, participant)
	withdrawal := &ChannelWithdrawal{
		ID:          withdrawalID,
		ChannelID:   channelID,
		Participant: participant,
		Amount:      new(big.Int).Set(amount),
		TxHash:      "",
		Timestamp:   time.Now(),
		Status:      "pending",
	}

	scm.withdrawals[withdrawalID] = withdrawal

	return withdrawal, nil
}

// ProcessWithdrawal обрабатывает вывод из канала
func (scm *StateChannelManager) ProcessWithdrawal(withdrawalID, txHash string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	withdrawal, exists := scm.withdrawals[withdrawalID]
	if !exists {
		return fmt.Errorf("withdrawal not found: %s", withdrawalID)
	}

	channel, exists := scm.channels[withdrawal.ChannelID]
	if !exists {
		return fmt.Errorf("channel not found: %s", withdrawal.ChannelID)
	}

	// Обновляем баланс
	channel.Balances[withdrawal.Participant].Sub(channel.Balances[withdrawal.Participant], withdrawal.Amount)
	channel.TotalDeposit.Sub(channel.TotalDeposit, withdrawal.Amount)
	channel.UpdatedAt = time.Now()

	// Обновляем статус вывода
	withdrawal.TxHash = txHash
	withdrawal.Status = "confirmed"

	return nil
}

// InitiateDispute инициирует спор по каналу
func (scm *StateChannelManager) InitiateDispute(channelID, initiator, reason, evidence string, stateUpdate *StateUpdate) (*ChannelDispute, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	// Создаем спор
	disputeID := scm.generateDisputeID(channelID)
	dispute := &ChannelDispute{
		ID:          disputeID,
		ChannelID:   channelID,
		Initiator:   initiator,
		Reason:      reason,
		Evidence:    evidence,
		StateUpdate: stateUpdate,
		Timestamp:   time.Now(),
		Status:      "open",
	}

	scm.disputes[disputeID] = dispute

	// Переводим канал в состояние спора
	channel.State = StateDisputed
	channel.UpdatedAt = time.Now()

	return dispute, nil
}

// ResolveDispute разрешает спор по каналу
func (scm *StateChannelManager) ResolveDispute(disputeID, resolution string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	dispute, exists := scm.disputes[disputeID]
	if !exists {
		return fmt.Errorf("dispute not found: %s", disputeID)
	}

	channel, exists := scm.channels[dispute.ChannelID]
	if !exists {
		return fmt.Errorf("channel not found: %s", dispute.ChannelID)
	}

	// Обновляем спор
	dispute.Status = "resolved"
	dispute.Resolution = resolution
	now := time.Now()
	dispute.ResolvedAt = &now

	// Переводим канал в состояние закрытия
	channel.State = StateClosing
	channel.UpdatedAt = time.Now()

	return nil
}

// CloseChannel закрывает канал
func (scm *StateChannelManager) CloseChannel(channelID string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	if channel.State != StateOpen && channel.State != StateClosing {
		return fmt.Errorf("channel cannot be closed in current state: %s", channel.State)
	}

	channel.State = StateClosed
	channel.UpdatedAt = time.Now()

	return nil
}

// SettleChannel урегулирует канал
func (scm *StateChannelManager) SettleChannel(channelID string, finalState *StateUpdate, txHash string) (*ChannelSettlement, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	// Создаем урегулирование
	settlementID := scm.generateSettlementID(channelID)
	settlement := &ChannelSettlement{
		ID:           settlementID,
		ChannelID:    channelID,
		FinalState:   finalState,
		TxHash:       txHash,
		Timestamp:    time.Now(),
		GasUsed:      100000, // Примерный газ для урегулирования
		GasPrice:     big.NewInt(20),
		Participants: channel.Participants,
		Balances:     channel.Balances,
	}

	scm.settlements[settlementID] = settlement

	// Переводим канал в состояние урегулированного
	channel.State = StateSettled
	channel.UpdatedAt = time.Now()

	return settlement, nil
}

// GetChannelStats возвращает статистику канала
func (scm *StateChannelManager) GetChannelStats(channelID string) (map[string]interface{}, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	// Подсчитываем статистику
	totalTxs := 0
	totalDeposits := 0
	totalWithdrawals := 0
	activeDisputes := 0

	for _, tx := range scm.transactions {
		if tx.ChannelID == channelID {
			totalTxs++
		}
	}

	for _, deposit := range scm.deposits {
		if deposit.ChannelID == channelID {
			totalDeposits++
		}
	}

	for _, withdrawal := range scm.withdrawals {
		if withdrawal.ChannelID == channelID {
			totalWithdrawals++
		}
	}

	for _, dispute := range scm.disputes {
		if dispute.ChannelID == channelID && dispute.Status == "open" {
			activeDisputes++
		}
	}

	return map[string]interface{}{
		"id":                channel.ID,
		"participants":      channel.Participants,
		"state":             channel.State,
		"nonce":             channel.Nonce,
		"total_deposit":     channel.TotalDeposit.String(),
		"balances":          channel.Balances,
		"channel_type":      channel.ChannelType,
		"created_at":        channel.CreatedAt,
		"updated_at":        channel.UpdatedAt,
		"expires_at":        channel.ExpiresAt,
		"total_transactions": totalTxs,
		"total_deposits":    totalDeposits,
		"total_withdrawals": totalWithdrawals,
		"active_disputes":   activeDisputes,
	}, nil
}

// GetChannelTransactions возвращает транзакции канала
func (scm *StateChannelManager) GetChannelTransactions(channelID string) ([]*ChannelTransaction, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	var transactions []*ChannelTransaction
	for _, tx := range scm.transactions {
		if tx.ChannelID == channelID {
			transactions = append(transactions, tx)
		}
	}
	return transactions, nil
}

// GetChannelDeposits возвращает депозиты канала
func (scm *StateChannelManager) GetChannelDeposits(channelID string) ([]*ChannelDeposit, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	var deposits []*ChannelDeposit
	for _, deposit := range scm.deposits {
		if deposit.ChannelID == channelID {
			deposits = append(deposits, deposit)
		}
	}
	return deposits, nil
}

// GetChannelWithdrawals возвращает выводы канала
func (scm *StateChannelManager) GetChannelWithdrawals(channelID string) ([]*ChannelWithdrawal, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	var withdrawals []*ChannelWithdrawal
	for _, withdrawal := range scm.withdrawals {
		if withdrawal.ChannelID == channelID {
			withdrawals = append(withdrawals, withdrawal)
		}
	}
	return withdrawals, nil
}

// GetChannelDisputes возвращает споры канала
func (scm *StateChannelManager) GetChannelDisputes(channelID string) ([]*ChannelDispute, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	var disputes []*ChannelDispute
	for _, dispute := range scm.disputes {
		if dispute.ChannelID == channelID {
			disputes = append(disputes, dispute)
		}
	}
	return disputes, nil
}

// ExportChannel экспортирует канал
func (scm *StateChannelManager) ExportChannel(channelID string) ([]byte, error) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	channel, exists := scm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	return json.MarshalIndent(channel, "", "  ")
}

// ImportChannel импортирует канал
func (scm *StateChannelManager) ImportChannel(data []byte) (*StateChannel, error) {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	var channel StateChannel
	err := json.Unmarshal(data, &channel)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel: %v", err)
	}

	// Проверяем, что канал с таким ID не существует
	if _, exists := scm.channels[channel.ID]; exists {
		return nil, fmt.Errorf("channel with ID %s already exists", channel.ID)
	}

	// Инициализируем карты, если они nil
	if channel.Balances == nil {
		channel.Balances = make(map[string]*big.Int)
	}
	if channel.Metadata == nil {
		channel.Metadata = make(map[string]interface{})
	}

	scm.channels[channel.ID] = &channel
	return &channel, nil
}

// generateChannelID генерирует уникальный ID для канала
func (scm *StateChannelManager) generateChannelID(participants []string) string {
	data := fmt.Sprintf("%v_%d", participants, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("channel_%x", hash[:8])
}

// generateTransactionID генерирует уникальный ID для транзакции
func (scm *StateChannelManager) generateTransactionID(channelID string) string {
	data := fmt.Sprintf("%s_%d", channelID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("tx_%x", hash[:8])
}

// generateDepositID генерирует уникальный ID для депозита
func (scm *StateChannelManager) generateDepositID(channelID, participant string) string {
	data := fmt.Sprintf("%s_%s_%d", channelID, participant, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("deposit_%x", hash[:8])
}

// generateWithdrawalID генерирует уникальный ID для вывода
func (scm *StateChannelManager) generateWithdrawalID(channelID, participant string) string {
	data := fmt.Sprintf("%s_%s_%d", channelID, participant, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("withdrawal_%x", hash[:8])
}

// generateDisputeID генерирует уникальный ID для спора
func (scm *StateChannelManager) generateDisputeID(channelID string) string {
	data := fmt.Sprintf("%s_%d", channelID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("dispute_%x", hash[:8])
}

// generateSettlementID генерирует уникальный ID для урегулирования
func (scm *StateChannelManager) generateSettlementID(channelID string) string {
	data := fmt.Sprintf("%s_%d", channelID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("settlement_%x", hash[:8])
}
