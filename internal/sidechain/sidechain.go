package sidechain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// Sidechain представляет отдельную боковую цепь
type Sidechain struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ParentChain string            `json:"parent_chain"` // ID родительской цепи
	Creator     string            `json:"creator"`
	CreatedAt   time.Time         `json:"created_at"`
	Status      SidechainStatus   `json:"status"`
	Config      SidechainConfig   `json:"config"`
	Blocks      []*SidechainBlock `json:"blocks"`
	Height      int64             `json:"height"`
	Hash        string            `json:"hash"`
	Validators  []string          `json:"validators"`
	Assets      map[string]*Asset `json:"assets"` // Поддерживаемые активы
}

// SidechainStatus представляет статус sidechain
type SidechainStatus string

const (
	StatusActive    SidechainStatus = "active"
	StatusInactive  SidechainStatus = "inactive"
	StatusPaused    SidechainStatus = "paused"
	StatusTerminated SidechainStatus = "terminated"
)

// SidechainConfig представляет конфигурацию sidechain
type SidechainConfig struct {
	ConsensusAlgorithm string `json:"consensus_algorithm"` // PoW, PoS, DPoS, PBFT
	BlockTime         int64  `json:"block_time"`          // Время между блоками в секундах
	Difficulty        int    `json:"difficulty"`          // Сложность майнинга
	MaxBlockSize      int64  `json:"max_block_size"`      // Максимальный размер блока
	GasLimit          int64  `json:"gas_limit"`           // Лимит газа
	ValidatorCount    int    `json:"validator_count"`     // Количество валидаторов
	BridgeEnabled     bool   `json:"bridge_enabled"`      // Включен ли мост с основной цепью
	CrossChainEnabled bool   `json:"cross_chain_enabled"` // Включены ли кросс-чейн транзакции
}

// SidechainBlock представляет блок в sidechain
type SidechainBlock struct {
	Index        int64                  `json:"index"`
	Timestamp    time.Time              `json:"timestamp"`
	PreviousHash string                 `json:"previous_hash"`
	Hash         string                 `json:"hash"`
	MerkleRoot   string                 `json:"merkle_root"`
	Nonce        int64                  `json:"nonce"`
	Difficulty   int                    `json:"difficulty"`
	Transactions []*SidechainTransaction `json:"transactions"`
	Validator    string                 `json:"validator"`
	Signature    string                 `json:"signature"`
}

// SidechainTransaction представляет транзакцию в sidechain
type SidechainTransaction struct {
	ID          string                 `json:"id"`
	Type        TransactionType        `json:"type"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Amount      *big.Int               `json:"amount"`
	Asset       string                 `json:"asset"`        // ID актива
	Data        map[string]interface{} `json:"data"`
	GasLimit    int64                  `json:"gas_limit"`
	GasPrice    *big.Int               `json:"gas_price"`
	Nonce       int64                  `json:"nonce"`
	Signature   string                 `json:"signature"`
	Timestamp   time.Time              `json:"timestamp"`
	ParentTxID  string                 `json:"parent_tx_id,omitempty"` // ID транзакции в родительской цепи
}

// TransactionType представляет тип транзакции
type TransactionType string

const (
	TxTypeTransfer     TransactionType = "transfer"
	TxTypeCrossChain   TransactionType = "cross_chain"
	TxTypeBridge       TransactionType = "bridge"
	TxTypeValidator    TransactionType = "validator"
	TxTypeContract     TransactionType = "contract"
	TxTypeAssetCreate  TransactionType = "asset_create"
	TxTypeAssetMint    TransactionType = "asset_mint"
	TxTypeAssetBurn    TransactionType = "asset_burn"
)

// Asset представляет актив в sidechain
type Asset struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"`
	Decimals    int       `json:"decimals"`
	TotalSupply *big.Int  `json:"total_supply"`
	Creator     string    `json:"creator"`
	CreatedAt   time.Time `json:"created_at"`
	Type        AssetType `json:"type"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AssetType представляет тип актива
type AssetType string

const (
	AssetTypeNative    AssetType = "native"
	AssetTypeToken     AssetType = "token"
	AssetTypeNFT       AssetType = "nft"
	AssetTypeBridged   AssetType = "bridged"
)

// BridgeTransaction представляет транзакцию моста между цепями
type BridgeTransaction struct {
	ID              string    `json:"id"`
	SourceChain     string    `json:"source_chain"`
	TargetChain     string    `json:"target_chain"`
	Asset           string    `json:"asset"`
	Amount          *big.Int  `json:"amount"`
	From            string    `json:"from"`
	To              string    `json:"to"`
	Status          BridgeStatus `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	SourceTxID      string    `json:"source_tx_id"`
	TargetTxID      string    `json:"target_tx_id,omitempty"`
	ValidatorProof  string    `json:"validator_proof,omitempty"`
}

// BridgeStatus представляет статус мостовой транзакции
type BridgeStatus string

const (
	BridgeStatusPending   BridgeStatus = "pending"
	BridgeStatusConfirmed BridgeStatus = "confirmed"
	BridgeStatusCompleted BridgeStatus = "completed"
	BridgeStatusFailed    BridgeStatus = "failed"
)

// CrossChainMessage представляет сообщение между цепями
type CrossChainMessage struct {
	ID          string                 `json:"id"`
	SourceChain string                 `json:"source_chain"`
	TargetChain string                 `json:"target_chain"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Status      string                 `json:"status"`
}

// SidechainManager управляет sidechains
type SidechainManager struct {
	sidechains map[string]*Sidechain
	bridgeTxs  map[string]*BridgeTransaction
	messages   map[string]*CrossChainMessage
}

// NewSidechainManager создает новый менеджер sidechains
func NewSidechainManager() *SidechainManager {
	return &SidechainManager{
		sidechains: make(map[string]*Sidechain),
		bridgeTxs:  make(map[string]*BridgeTransaction),
		messages:   make(map[string]*CrossChainMessage),
	}
}

// CreateSidechain создает новую sidechain
func (sm *SidechainManager) CreateSidechain(name, description, creator, parentChain string, config SidechainConfig) (*Sidechain, error) {
	// Генерируем уникальный ID
	id := sm.generateSidechainID(name, creator)
	
	// Создаем genesis блок
	genesisBlock := &SidechainBlock{
		Index:        0,
		Timestamp:    time.Now(),
		PreviousHash: "0",
		Hash:         "genesis",
		MerkleRoot:   "0",
		Nonce:        0,
		Difficulty:   config.Difficulty,
		Transactions: []*SidechainTransaction{},
		Validator:    creator,
		Signature:    "",
	}
	
	sidechain := &Sidechain{
		ID:          id,
		Name:        name,
		Description: description,
		ParentChain: parentChain,
		Creator:     creator,
		CreatedAt:   time.Now(),
		Status:      StatusActive,
		Config:      config,
		Blocks:      []*SidechainBlock{genesisBlock},
		Height:      0,
		Hash:        "genesis",
		Validators:  []string{creator},
		Assets:      make(map[string]*Asset),
	}
	
	// Создаем нативный актив
	nativeAsset := &Asset{
		ID:          "native",
		Name:        fmt.Sprintf("%s Native", name),
		Symbol:      "NATIVE",
		Decimals:    18,
		TotalSupply: big.NewInt(0),
		Creator:     creator,
		CreatedAt:   time.Now(),
		Type:        AssetTypeNative,
		Metadata:    make(map[string]interface{}),
	}
	sidechain.Assets["native"] = nativeAsset
	
	sm.sidechains[id] = sidechain
	return sidechain, nil
}

// GetSidechain возвращает sidechain по ID
func (sm *SidechainManager) GetSidechain(id string) (*Sidechain, bool) {
	sidechain, exists := sm.sidechains[id]
	return sidechain, exists
}

// ListSidechains возвращает список всех sidechains
func (sm *SidechainManager) ListSidechains() []*Sidechain {
	sidechains := make([]*Sidechain, 0, len(sm.sidechains))
	for _, sidechain := range sm.sidechains {
		sidechains = append(sidechains, sidechain)
	}
	return sidechains
}

// AddBlock добавляет блок в sidechain
func (sm *SidechainManager) AddBlock(sidechainID string, block *SidechainBlock) error {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	// Проверяем валидность блока
	if err := sm.validateBlock(sidechain, block); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}
	
	// Добавляем блок
	sidechain.Blocks = append(sidechain.Blocks, block)
	sidechain.Height = block.Index
	sidechain.Hash = block.Hash
	
	return nil
}

// AddTransaction добавляет транзакцию в sidechain
func (sm *SidechainManager) AddTransaction(sidechainID string, tx *SidechainTransaction) error {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	// Проверяем валидность транзакции
	if err := sm.validateTransaction(sidechain, tx); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}
	
	// Добавляем транзакцию в последний блок
	if len(sidechain.Blocks) > 0 {
		lastBlock := sidechain.Blocks[len(sidechain.Blocks)-1]
		lastBlock.Transactions = append(lastBlock.Transactions, tx)
	}
	
	return nil
}

// CreateAsset создает новый актив в sidechain
func (sm *SidechainManager) CreateAsset(sidechainID, name, symbol string, decimals int, totalSupply *big.Int, creator string, assetType AssetType) (*Asset, error) {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return nil, fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	// Генерируем ID актива
	assetID := sm.generateAssetID(name, symbol, sidechainID)
	
	asset := &Asset{
		ID:          assetID,
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: new(big.Int).Set(totalSupply),
		Creator:     creator,
		CreatedAt:   time.Now(),
		Type:        assetType,
		Metadata:    make(map[string]interface{}),
	}
	
	sidechain.Assets[assetID] = asset
	return asset, nil
}

// GetAsset возвращает актив по ID
func (sm *SidechainManager) GetAsset(sidechainID, assetID string) (*Asset, error) {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return nil, fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	asset, exists := sidechain.Assets[assetID]
	if !exists {
		return nil, fmt.Errorf("asset not found: %s", assetID)
	}
	
	return asset, nil
}

// ListAssets возвращает список активов в sidechain
func (sm *SidechainManager) ListAssets(sidechainID string) ([]*Asset, error) {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return nil, fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	assets := make([]*Asset, 0, len(sidechain.Assets))
	for _, asset := range sidechain.Assets {
		assets = append(assets, asset)
	}
	
	return assets, nil
}

// CreateBridgeTransaction создает мостовую транзакцию
func (sm *SidechainManager) CreateBridgeTransaction(sourceChain, targetChain, asset string, amount *big.Int, from, to string) (*BridgeTransaction, error) {
	txID := sm.generateBridgeTxID(sourceChain, targetChain, from)
	
	bridgeTx := &BridgeTransaction{
		ID:          txID,
		SourceChain: sourceChain,
		TargetChain: targetChain,
		Asset:       asset,
		Amount:      new(big.Int).Set(amount),
		From:        from,
		To:          to,
		Status:      BridgeStatusPending,
		CreatedAt:   time.Now(),
	}
	
	sm.bridgeTxs[txID] = bridgeTx
	return bridgeTx, nil
}

// ProcessBridgeTransaction обрабатывает мостовую транзакцию
func (sm *SidechainManager) ProcessBridgeTransaction(txID, targetTxID, validatorProof string) error {
	bridgeTx, exists := sm.bridgeTxs[txID]
	if !exists {
		return fmt.Errorf("bridge transaction not found: %s", txID)
	}
	
	bridgeTx.TargetTxID = targetTxID
	bridgeTx.ValidatorProof = validatorProof
	bridgeTx.Status = BridgeStatusCompleted
	now := time.Now()
	bridgeTx.CompletedAt = &now
	
	return nil
}

// GetBridgeTransaction возвращает мостовую транзакцию
func (sm *SidechainManager) GetBridgeTransaction(txID string) (*BridgeTransaction, error) {
	bridgeTx, exists := sm.bridgeTxs[txID]
	if !exists {
		return nil, fmt.Errorf("bridge transaction not found: %s", txID)
	}
	
	return bridgeTx, nil
}

// ListBridgeTransactions возвращает список мостовых транзакций
func (sm *SidechainManager) ListBridgeTransactions() []*BridgeTransaction {
	transactions := make([]*BridgeTransaction, 0, len(sm.bridgeTxs))
	for _, tx := range sm.bridgeTxs {
		transactions = append(transactions, tx)
	}
	return transactions
}

// SendCrossChainMessage отправляет кросс-чейн сообщение
func (sm *SidechainManager) SendCrossChainMessage(sourceChain, targetChain, msgType string, data map[string]interface{}) (*CrossChainMessage, error) {
	msgID := sm.generateMessageID(sourceChain, targetChain)
	
	message := &CrossChainMessage{
		ID:          msgID,
		SourceChain: sourceChain,
		TargetChain: targetChain,
		Type:        msgType,
		Data:        data,
		CreatedAt:   time.Now(),
		Status:      "pending",
	}
	
	sm.messages[msgID] = message
	return message, nil
}

// ProcessCrossChainMessage обрабатывает кросс-чейн сообщение
func (sm *SidechainManager) ProcessCrossChainMessage(msgID string) error {
	message, exists := sm.messages[msgID]
	if !exists {
		return fmt.Errorf("message not found: %s", msgID)
	}
	
	message.Status = "processed"
	now := time.Now()
	message.ProcessedAt = &now
	
	return nil
}

// GetSidechainStats возвращает статистику sidechain
func (sm *SidechainManager) GetSidechainStats(sidechainID string) (map[string]interface{}, error) {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return nil, fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	// Подсчитываем статистику
	totalTxs := 0
	totalAssets := len(sidechain.Assets)
	activeValidators := len(sidechain.Validators)
	
	for _, block := range sidechain.Blocks {
		totalTxs += len(block.Transactions)
	}
	
	// Подсчитываем мостовые транзакции
	bridgeTxs := 0
	for _, tx := range sm.bridgeTxs {
		if tx.SourceChain == sidechainID || tx.TargetChain == sidechainID {
			bridgeTxs++
		}
	}
	
	return map[string]interface{}{
		"id":                sidechain.ID,
		"name":              sidechain.Name,
		"status":            sidechain.Status,
		"height":            sidechain.Height,
		"total_blocks":      len(sidechain.Blocks),
		"total_transactions": totalTxs,
		"total_assets":      totalAssets,
		"active_validators": activeValidators,
		"bridge_transactions": bridgeTxs,
		"created_at":        sidechain.CreatedAt,
		"consensus":         sidechain.Config.ConsensusAlgorithm,
		"block_time":        sidechain.Config.BlockTime,
	}, nil
}

// validateBlock проверяет валидность блока
func (sm *SidechainManager) validateBlock(sidechain *Sidechain, block *SidechainBlock) error {
	// Проверяем последовательность индексов
	if block.Index != sidechain.Height+1 {
		return fmt.Errorf("invalid block index: expected %d, got %d", sidechain.Height+1, block.Index)
	}
	
	// Проверяем предыдущий хеш
	if block.Index > 0 && block.PreviousHash != sidechain.Hash {
		return fmt.Errorf("invalid previous hash")
	}
	
	// Проверяем валидатора
	validValidator := false
	for _, validator := range sidechain.Validators {
		if validator == block.Validator {
			validValidator = true
			break
		}
	}
	if !validValidator {
		return fmt.Errorf("invalid validator: %s", block.Validator)
	}
	
	return nil
}

// validateTransaction проверяет валидность транзакции
func (sm *SidechainManager) validateTransaction(sidechain *Sidechain, tx *SidechainTransaction) error {
	// Проверяем тип транзакции
	switch tx.Type {
	case TxTypeTransfer, TxTypeCrossChain, TxTypeBridge, TxTypeValidator, TxTypeContract, TxTypeAssetCreate, TxTypeAssetMint, TxTypeAssetBurn:
		// Валидные типы
	default:
		return fmt.Errorf("invalid transaction type: %s", tx.Type)
	}
	
	// Проверяем актив
	if tx.Asset != "" {
		_, exists := sidechain.Assets[tx.Asset]
		if !exists {
			return fmt.Errorf("asset not found: %s", tx.Asset)
		}
	}
	
	return nil
}

// generateSidechainID генерирует уникальный ID для sidechain
func (sm *SidechainManager) generateSidechainID(name, creator string) string {
	data := fmt.Sprintf("%s_%s_%d", name, creator, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sidechain_%x", hash[:8])
}

// generateAssetID генерирует уникальный ID для актива
func (sm *SidechainManager) generateAssetID(name, symbol, sidechainID string) string {
	data := fmt.Sprintf("%s_%s_%s_%d", name, symbol, sidechainID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("asset_%x", hash[:8])
}

// generateBridgeTxID генерирует уникальный ID для мостовой транзакции
func (sm *SidechainManager) generateBridgeTxID(sourceChain, targetChain, from string) string {
	data := fmt.Sprintf("%s_%s_%s_%d", sourceChain, targetChain, from, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("bridge_%x", hash[:8])
}

// generateMessageID генерирует уникальный ID для сообщения
func (sm *SidechainManager) generateMessageID(sourceChain, targetChain string) string {
	data := fmt.Sprintf("%s_%s_%d", sourceChain, targetChain, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("msg_%x", hash[:8])
}

// ExportSidechain экспортирует sidechain
func (sm *SidechainManager) ExportSidechain(sidechainID string) ([]byte, error) {
	sidechain, exists := sm.GetSidechain(sidechainID)
	if !exists {
		return nil, fmt.Errorf("sidechain not found: %s", sidechainID)
	}
	
	return json.MarshalIndent(sidechain, "", "  ")
}

// ImportSidechain импортирует sidechain
func (sm *SidechainManager) ImportSidechain(data []byte) (*Sidechain, error) {
	var sidechain Sidechain
	err := json.Unmarshal(data, &sidechain)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal sidechain: %v", err)
	}
	
	// Проверяем, что sidechain с таким ID не существует
	if _, exists := sm.sidechains[sidechain.ID]; exists {
		return nil, fmt.Errorf("sidechain with ID %s already exists", sidechain.ID)
	}
	
	// Инициализируем карты, если они nil
	if sidechain.Assets == nil {
		sidechain.Assets = make(map[string]*Asset)
	}
	
	sm.sidechains[sidechain.ID] = &sidechain
	return &sidechain, nil
}
