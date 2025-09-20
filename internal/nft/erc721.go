package nft

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// ERC721Token представляет NFT токен в стандарте ERC-721
type ERC721Token struct {
	TokenID     *big.Int         `json:"token_id"`
	Contract    string           `json:"contract"`
	Owner       string           `json:"owner"`
	Metadata    *TokenMetadata   `json:"metadata"`
	CreatedAt   time.Time        `json:"created_at"`
	TransferredAt time.Time      `json:"transferred_at"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// TokenMetadata представляет метаданные NFT
type TokenMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	ExternalURL string `json:"external_url"`
	AnimationURL string `json:"animation_url,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	YouTubeURL  string `json:"youtube_url,omitempty"`
}

// ERC721Contract представляет контракт NFT коллекции
type ERC721Contract struct {
	Address     string            `json:"address"`
	Name        string            `json:"name"`
	Symbol      string            `json:"symbol"`
	Owner       string            `json:"owner"`
	CreatedAt   time.Time         `json:"created_at"`
	TotalSupply *big.Int          `json:"total_supply"`
	MaxSupply   *big.Int          `json:"max_supply,omitempty"`
	BaseURI     string            `json:"base_uri"`
	Tokens      map[string]*ERC721Token `json:"tokens"` // tokenID -> token
	Approvals   map[string]string `json:"approvals"`   // tokenID -> approved address
	Operators   map[string]map[string]bool `json:"operators"` // owner -> operator -> approved
}

// TransferEvent представляет событие передачи NFT
type TransferEvent struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	TokenID *big.Int `json:"token_id"`
	Contract string  `json:"contract"`
}

// ApprovalEvent представляет событие одобрения NFT
type ApprovalEvent struct {
	Owner    string   `json:"owner"`
	Approved string   `json:"approved"`
	TokenID  *big.Int `json:"token_id"`
	Contract string   `json:"contract"`
}

// ApprovalForAllEvent представляет событие одобрения всех токенов
type ApprovalForAllEvent struct {
	Owner    string `json:"owner"`
	Operator string `json:"operator"`
	Approved bool   `json:"approved"`
	Contract string `json:"contract"`
}

// ERC721Manager управляет NFT токенами и контрактами
type ERC721Manager struct {
	contracts map[string]*ERC721Contract
}

// NewERC721Manager создает новый менеджер NFT
func NewERC721Manager() *ERC721Manager {
	return &ERC721Manager{
		contracts: make(map[string]*ERC721Contract),
	}
}

// CreateContract создает новый NFT контракт
func (em *ERC721Manager) CreateContract(name, symbol, owner, baseURI string, maxSupply *big.Int) (*ERC721Contract, error) {
	// Генерируем адрес контракта
	address := em.generateContractAddress()
	
	contract := &ERC721Contract{
		Address:     address,
		Name:        name,
		Symbol:      symbol,
		Owner:       owner,
		CreatedAt:   time.Now(),
		TotalSupply: big.NewInt(0),
		MaxSupply:   maxSupply,
		BaseURI:     baseURI,
		Tokens:      make(map[string]*ERC721Token),
		Approvals:   make(map[string]string),
		Operators:   make(map[string]map[string]bool),
	}
	
	em.contracts[address] = contract
	return contract, nil
}

// GetContract возвращает контракт по адресу
func (em *ERC721Manager) GetContract(address string) (*ERC721Contract, bool) {
	contract, exists := em.contracts[address]
	return contract, exists
}

// ListContracts возвращает список всех контрактов
func (em *ERC721Manager) ListContracts() []*ERC721Contract {
	contracts := make([]*ERC721Contract, 0, len(em.contracts))
	for _, contract := range em.contracts {
		contracts = append(contracts, contract)
	}
	return contracts
}

// Mint создает новый NFT токен
func (em *ERC721Manager) Mint(contractAddress, to string, tokenID *big.Int, metadata *TokenMetadata, attributes map[string]interface{}) (*ERC721Token, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	// Проверяем максимальное предложение
	if contract.MaxSupply != nil && contract.TotalSupply.Cmp(contract.MaxSupply) >= 0 {
		return nil, fmt.Errorf("max supply reached")
	}
	
	// Проверяем, что токен не существует
	tokenIDStr := tokenID.String()
	if _, exists := contract.Tokens[tokenIDStr]; exists {
		return nil, fmt.Errorf("token already exists: %s", tokenIDStr)
	}
	
	// Создаем NFT токен
	token := &ERC721Token{
		TokenID:     new(big.Int).Set(tokenID),
		Contract:    contractAddress,
		Owner:       to,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		TransferredAt: time.Now(),
		Attributes:  attributes,
	}
	
	contract.Tokens[tokenIDStr] = token
	contract.TotalSupply.Add(contract.TotalSupply, big.NewInt(1))
	
	return token, nil
}

// Transfer переводит NFT от одного владельца к другому
func (em *ERC721Manager) Transfer(contractAddress, from, to string, tokenID *big.Int) (*TransferEvent, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	token, exists := contract.Tokens[tokenIDStr]
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenIDStr)
	}
	
	// Проверяем владельца
	if token.Owner != from {
		return nil, fmt.Errorf("not the owner of token: %s", tokenIDStr)
	}
	
	// Выполняем перевод
	token.Owner = to
	token.TransferredAt = time.Now()
	
	// Удаляем одобрение
	delete(contract.Approvals, tokenIDStr)
	
	// Создаем событие
	event := &TransferEvent{
		From:    from,
		To:      to,
		TokenID: new(big.Int).Set(tokenID),
		Contract: contractAddress,
	}
	
	return event, nil
}

// Approve одобряет адрес для управления конкретным токеном
func (em *ERC721Manager) Approve(contractAddress, owner, approved string, tokenID *big.Int) (*ApprovalEvent, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	token, exists := contract.Tokens[tokenIDStr]
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenIDStr)
	}
	
	// Проверяем владельца
	if token.Owner != owner {
		return nil, fmt.Errorf("not the owner of token: %s", tokenIDStr)
	}
	
	// Устанавливаем одобрение
	contract.Approvals[tokenIDStr] = approved
	
	// Создаем событие
	event := &ApprovalEvent{
		Owner:    owner,
		Approved: approved,
		TokenID:  new(big.Int).Set(tokenID),
		Contract: contractAddress,
	}
	
	return event, nil
}

// SetApprovalForAll одобряет или отзывает одобрение для всех токенов владельца
func (em *ERC721Manager) SetApprovalForAll(contractAddress, owner, operator string, approved bool) (*ApprovalForAllEvent, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	// Инициализируем карту операторов для владельца
	if contract.Operators[owner] == nil {
		contract.Operators[owner] = make(map[string]bool)
	}
	
	// Устанавливаем одобрение
	contract.Operators[owner][operator] = approved
	
	// Создаем событие
	event := &ApprovalForAllEvent{
		Owner:    owner,
		Operator: operator,
		Approved: approved,
		Contract: contractAddress,
	}
	
	return event, nil
}

// TransferFrom переводит NFT от имени владельца
func (em *ERC721Manager) TransferFrom(contractAddress, spender, from, to string, tokenID *big.Int) (*TransferEvent, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	token, exists := contract.Tokens[tokenIDStr]
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenIDStr)
	}
	
	// Проверяем права на перевод
	if token.Owner != from {
		return nil, fmt.Errorf("not the owner of token: %s", tokenIDStr)
	}
	
	// Проверяем одобрение
	approved, hasApproval := contract.Approvals[tokenIDStr]
	operatorApproved := contract.Operators[from][spender]
	
	if !hasApproval && !operatorApproved {
		return nil, fmt.Errorf("not approved to transfer token: %s", tokenIDStr)
	}
	
	if hasApproval && approved != spender {
		return nil, fmt.Errorf("not approved to transfer token: %s", tokenIDStr)
	}
	
	// Выполняем перевод
	return em.Transfer(contractAddress, from, to, tokenID)
}

// OwnerOf возвращает владельца токена
func (em *ERC721Manager) OwnerOf(contractAddress string, tokenID *big.Int) (string, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return "", fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	token, exists := contract.Tokens[tokenIDStr]
	if !exists {
		return "", fmt.Errorf("token not found: %s", tokenIDStr)
	}
	
	return token.Owner, nil
}

// GetApproved возвращает одобренный адрес для токена
func (em *ERC721Manager) GetApproved(contractAddress string, tokenID *big.Int) (string, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return "", fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	approved, exists := contract.Approvals[tokenIDStr]
	if !exists {
		return "", nil
	}
	
	return approved, nil
}

// IsApprovedForAll проверяет, одобрен ли оператор для всех токенов владельца
func (em *ERC721Manager) IsApprovedForAll(contractAddress, owner, operator string) (bool, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return false, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	approved, exists := contract.Operators[owner][operator]
	if !exists {
		return false, nil
	}
	
	return approved, nil
}

// BalanceOf возвращает количество токенов у владельца
func (em *ERC721Manager) BalanceOf(contractAddress, owner string) (*big.Int, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	count := big.NewInt(0)
	for _, token := range contract.Tokens {
		if token.Owner == owner {
			count.Add(count, big.NewInt(1))
		}
	}
	
	return count, nil
}

// GetToken возвращает токен по ID
func (em *ERC721Manager) GetToken(contractAddress string, tokenID *big.Int) (*ERC721Token, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	token, exists := contract.Tokens[tokenIDStr]
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenIDStr)
	}
	
	return token, nil
}

// GetTokensByOwner возвращает все токены владельца
func (em *ERC721Manager) GetTokensByOwner(contractAddress, owner string) ([]*ERC721Token, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokens := make([]*ERC721Token, 0)
	for _, token := range contract.Tokens {
		if token.Owner == owner {
			tokens = append(tokens, token)
		}
	}
	
	return tokens, nil
}

// GetContractInfo возвращает информацию о контракте
func (em *ERC721Manager) GetContractInfo(contractAddress string) (map[string]interface{}, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	// Подсчитываем количество уникальных владельцев
	owners := make(map[string]bool)
	for _, token := range contract.Tokens {
		owners[token.Owner] = true
	}
	
	// Подсчитываем активные одобрения
	activeApprovals := 0
	for range contract.Approvals {
		activeApprovals++
	}
	
	// Подсчитываем активных операторов
	activeOperators := 0
	for _, operators := range contract.Operators {
		for _, approved := range operators {
			if approved {
				activeOperators++
			}
		}
	}
	
	return map[string]interface{}{
		"address":           contract.Address,
		"name":              contract.Name,
		"symbol":            contract.Symbol,
		"owner":             contract.Owner,
		"created_at":        contract.CreatedAt,
		"total_supply":      contract.TotalSupply.String(),
		"max_supply":        contract.MaxSupply.String(),
		"base_uri":          contract.BaseURI,
		"unique_owners":     len(owners),
		"active_approvals":  activeApprovals,
		"active_operators":  activeOperators,
		"token_count":       len(contract.Tokens),
	}, nil
}

// GetContractStats возвращает статистику контракта
func (em *ERC721Manager) GetContractStats(contractAddress string) (map[string]interface{}, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	// Подсчитываем статистику
	owners := make(map[string]int)
	transfers := 0
	oldestToken := time.Now()
	newestToken := time.Time{}
	
	for _, token := range contract.Tokens {
		owners[token.Owner]++
		transfers++
		
		if token.CreatedAt.Before(oldestToken) {
			oldestToken = token.CreatedAt
		}
		if token.CreatedAt.After(newestToken) {
			newestToken = token.CreatedAt
		}
	}
	
	// Находим владельца с наибольшим количеством токенов
	maxTokens := 0
	maxOwner := ""
	for owner, count := range owners {
		if count > maxTokens {
			maxTokens = count
			maxOwner = owner
		}
	}
	
	return map[string]interface{}{
		"total_owners":      len(owners),
		"total_transfers":   transfers,
		"oldest_token":      oldestToken,
		"newest_token":      newestToken,
		"max_tokens_owner":  maxOwner,
		"max_tokens_count":  maxTokens,
		"average_tokens":    float64(len(contract.Tokens)) / float64(len(owners)),
	}, nil
}

// SearchTokens ищет токены по критериям
func (em *ERC721Manager) SearchTokens(criteria map[string]interface{}) ([]*ERC721Token, error) {
	results := make([]*ERC721Token, 0)
	
	for _, contract := range em.contracts {
		for _, token := range contract.Tokens {
			matches := true
			
			// Фильтр по контракту
			if contractAddr, ok := criteria["contract"].(string); ok && contractAddr != "" {
				if token.Contract != contractAddr {
					matches = false
				}
			}
			
			// Фильтр по владельцу
			if owner, ok := criteria["owner"].(string); ok && owner != "" {
				if token.Owner != owner {
					matches = false
				}
			}
			
			// Фильтр по имени
			if name, ok := criteria["name"].(string); ok && name != "" {
				if token.Metadata.Name != name {
					matches = false
				}
			}
			
			// Фильтр по атрибутам
			if attributes, ok := criteria["attributes"].(map[string]interface{}); ok {
				for key, value := range attributes {
					if token.Attributes[key] != value {
						matches = false
						break
					}
				}
			}
			
			if matches {
				results = append(results, token)
			}
		}
	}
	
	return results, nil
}

// Burn сжигает NFT токен
func (em *ERC721Manager) Burn(contractAddress, owner string, tokenID *big.Int) error {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	tokenIDStr := tokenID.String()
	token, exists := contract.Tokens[tokenIDStr]
	if !exists {
		return fmt.Errorf("token not found: %s", tokenIDStr)
	}
	
	// Проверяем владельца
	if token.Owner != owner {
		return fmt.Errorf("not the owner of token: %s", tokenIDStr)
	}
	
	// Удаляем токен
	delete(contract.Tokens, tokenIDStr)
	delete(contract.Approvals, tokenIDStr)
	contract.TotalSupply.Sub(contract.TotalSupply, big.NewInt(1))
	
	return nil
}

// generateContractAddress генерирует адрес контракта
func (em *ERC721Manager) generateContractAddress() string {
	// Простая генерация адреса на основе времени и количества контрактов
	timestamp := time.Now().UnixNano()
	count := len(em.contracts)
	return fmt.Sprintf("nft_contract_%d_%d", timestamp, count)
}

// ExportContract экспортирует контракт в JSON
func (em *ERC721Manager) ExportContract(contractAddress string) ([]byte, error) {
	contract, exists := em.GetContract(contractAddress)
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	return json.MarshalIndent(contract, "", "  ")
}

// ImportContract импортирует контракт из JSON
func (em *ERC721Manager) ImportContract(data []byte) (*ERC721Contract, error) {
	var contract ERC721Contract
	err := json.Unmarshal(data, &contract)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract: %v", err)
	}
	
	// Проверяем, что контракт с таким адресом не существует
	if _, exists := em.contracts[contract.Address]; exists {
		return nil, fmt.Errorf("contract with address %s already exists", contract.Address)
	}
	
	// Инициализируем карты, если они nil
	if contract.Tokens == nil {
		contract.Tokens = make(map[string]*ERC721Token)
	}
	if contract.Approvals == nil {
		contract.Approvals = make(map[string]string)
	}
	if contract.Operators == nil {
		contract.Operators = make(map[string]map[string]bool)
	}
	
	em.contracts[contract.Address] = &contract
	return &contract, nil
}
