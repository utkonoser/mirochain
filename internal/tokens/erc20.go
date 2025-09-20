package tokens

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// ERC20Token представляет токен в стандарте ERC-20
type ERC20Token struct {
	Address     string                         `json:"address"`
	Name        string                         `json:"name"`
	Symbol      string                         `json:"symbol"`
	Decimals    uint8                          `json:"decimals"`
	TotalSupply *big.Int                       `json:"total_supply"`
	Owner       string                         `json:"owner"`
	CreatedAt   time.Time                      `json:"created_at"`
	Balances    map[string]*big.Int            `json:"balances"`
	Allowances  map[string]map[string]*big.Int `json:"allowances"` // owner -> spender -> amount
}

// ERC20TransferEvent представляет событие передачи токенов
type ERC20TransferEvent struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Value   *big.Int `json:"value"`
	Address string   `json:"address"`
}

// ERC20ApprovalEvent представляет событие одобрения токенов
type ERC20ApprovalEvent struct {
	Owner   string   `json:"owner"`
	Spender string   `json:"spender"`
	Value   *big.Int `json:"value"`
	Address string   `json:"address"`
}

// ERC20Manager управляет токенами ERC-20
type ERC20Manager struct {
	tokens map[string]*ERC20Token
}

// NewERC20Manager создает новый менеджер токенов
func NewERC20Manager() *ERC20Manager {
	return &ERC20Manager{
		tokens: make(map[string]*ERC20Token),
	}
}

// CreateToken создает новый токен
func (em *ERC20Manager) CreateToken(name, symbol string, decimals uint8, totalSupply *big.Int, owner string) (*ERC20Token, error) {
	// Генерируем адрес токена
	address := em.generateTokenAddress()

	token := &ERC20Token{
		Address:     address,
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: new(big.Int).Set(totalSupply),
		Owner:       owner,
		CreatedAt:   time.Now(),
		Balances:    make(map[string]*big.Int),
		Allowances:  make(map[string]map[string]*big.Int),
	}

	// Устанавливаем начальный баланс владельца
	token.Balances[owner] = new(big.Int).Set(totalSupply)

	em.tokens[address] = token
	return token, nil
}

// GetToken возвращает токен по адресу
func (em *ERC20Manager) GetToken(address string) (*ERC20Token, bool) {
	token, exists := em.tokens[address]
	return token, exists
}

// ListTokens возвращает список всех токенов
func (em *ERC20Manager) ListTokens() []*ERC20Token {
	tokens := make([]*ERC20Token, 0, len(em.tokens))
	for _, token := range em.tokens {
		tokens = append(tokens, token)
	}
	return tokens
}

// Transfer переводит токены от одного адреса к другому
func (em *ERC20Manager) Transfer(tokenAddress, from, to string, amount *big.Int) (*ERC20TransferEvent, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Проверяем баланс отправителя
	fromBalance, exists := token.Balances[from]
	if !exists {
		fromBalance = big.NewInt(0)
		token.Balances[from] = fromBalance
	}

	if fromBalance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient balance: have %s, need %s", fromBalance.String(), amount.String())
	}

	// Выполняем перевод
	fromBalance.Sub(fromBalance, amount)

	toBalance, exists := token.Balances[to]
	if !exists {
		toBalance = big.NewInt(0)
		token.Balances[to] = toBalance
	}
	toBalance.Add(toBalance, amount)

	// Создаем событие
	event := &ERC20TransferEvent{
		From:    from,
		To:      to,
		Value:   new(big.Int).Set(amount),
		Address: tokenAddress,
	}

	return event, nil
}

// Approve одобряет расход токенов
func (em *ERC20Manager) Approve(tokenAddress, owner, spender string, amount *big.Int) (*ERC20ApprovalEvent, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Инициализируем карту разрешений для владельца, если не существует
	if token.Allowances[owner] == nil {
		token.Allowances[owner] = make(map[string]*big.Int)
	}

	// Устанавливаем разрешение
	token.Allowances[owner][spender] = new(big.Int).Set(amount)

	// Создаем событие
	event := &ERC20ApprovalEvent{
		Owner:   owner,
		Spender: spender,
		Value:   new(big.Int).Set(amount),
		Address: tokenAddress,
	}

	return event, nil
}

// TransferFrom переводит токены от имени владельца
func (em *ERC20Manager) TransferFrom(tokenAddress, spender, from, to string, amount *big.Int) (*ERC20TransferEvent, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Проверяем разрешение
	allowance, exists := token.Allowances[from][spender]
	if !exists || allowance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("insufficient allowance: have %s, need %s", allowance.String(), amount.String())
	}

	// Выполняем перевод
	event, err := em.Transfer(tokenAddress, from, to, amount)
	if err != nil {
		return nil, err
	}

	// Уменьшаем разрешение
	allowance.Sub(allowance, amount)

	return event, nil
}

// BalanceOf возвращает баланс токенов для адреса
func (em *ERC20Manager) BalanceOf(tokenAddress, address string) (*big.Int, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	balance, exists := token.Balances[address]
	if !exists {
		return big.NewInt(0), nil
	}

	return new(big.Int).Set(balance), nil
}

// Allowance возвращает разрешение на расход токенов
func (em *ERC20Manager) Allowance(tokenAddress, owner, spender string) (*big.Int, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	allowance, exists := token.Allowances[owner][spender]
	if !exists {
		return big.NewInt(0), nil
	}

	return new(big.Int).Set(allowance), nil
}

// GetTotalSupply возвращает общее количество токенов
func (em *ERC20Manager) GetTotalSupply(tokenAddress string) (*big.Int, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	return new(big.Int).Set(token.TotalSupply), nil
}

// GetTokenInfo возвращает информацию о токене
func (em *ERC20Manager) GetTokenInfo(tokenAddress string) (map[string]interface{}, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Подсчитываем количество держателей
	holderCount := 0
	for _, balance := range token.Balances {
		if balance.Cmp(big.NewInt(0)) > 0 {
			holderCount++
		}
	}

	// Подсчитываем общий объем разрешений
	totalAllowances := big.NewInt(0)
	for _, allowances := range token.Allowances {
		for _, allowance := range allowances {
			totalAllowances.Add(totalAllowances, allowance)
		}
	}

	return map[string]interface{}{
		"address":          token.Address,
		"name":             token.Name,
		"symbol":           token.Symbol,
		"decimals":         token.Decimals,
		"total_supply":     token.TotalSupply.String(),
		"owner":            token.Owner,
		"created_at":       token.CreatedAt,
		"holder_count":     holderCount,
		"total_allowances": totalAllowances.String(),
		"balance_count":    len(token.Balances),
		"allowance_count":  len(token.Allowances),
	}, nil
}

// generateTokenAddress генерирует адрес токена
func (em *ERC20Manager) generateTokenAddress() string {
	// Простая генерация адреса на основе времени и количества токенов
	timestamp := time.Now().UnixNano()
	count := len(em.tokens)
	return fmt.Sprintf("token_%d_%d", timestamp, count)
}

// Burn сжигает токены (уменьшает общее предложение)
func (em *ERC20Manager) Burn(tokenAddress, from string, amount *big.Int) error {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Проверяем баланс
	balance, exists := token.Balances[from]
	if !exists || balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance for burning")
	}

	// Уменьшаем баланс и общее предложение
	balance.Sub(balance, amount)
	token.TotalSupply.Sub(token.TotalSupply, amount)

	return nil
}

// Mint создает новые токены (увеличивает общее предложение)
func (em *ERC20Manager) Mint(tokenAddress, to string, amount *big.Int) error {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Проверяем, что только владелец может создавать токены
	if to != token.Owner {
		return fmt.Errorf("only token owner can mint tokens")
	}

	// Увеличиваем баланс и общее предложение
	balance, exists := token.Balances[to]
	if !exists {
		balance = big.NewInt(0)
		token.Balances[to] = balance
	}
	balance.Add(balance, amount)
	token.TotalSupply.Add(token.TotalSupply, amount)

	return nil
}

// GetTokenStats возвращает статистику токена
func (em *ERC20Manager) GetTokenStats(tokenAddress string) (map[string]interface{}, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	// Подсчитываем статистику
	totalHolders := 0
	totalBalance := big.NewInt(0)
	maxBalance := big.NewInt(0)

	for _, balance := range token.Balances {
		if balance.Cmp(big.NewInt(0)) > 0 {
			totalHolders++
			totalBalance.Add(totalBalance, balance)
			if balance.Cmp(maxBalance) > 0 {
				maxBalance.Set(balance)
			}
		}
	}

	// Подсчитываем активные разрешения
	activeAllowances := 0
	for _, allowances := range token.Allowances {
		for _, allowance := range allowances {
			if allowance.Cmp(big.NewInt(0)) > 0 {
				activeAllowances++
			}
		}
	}

	return map[string]interface{}{
		"total_holders":     totalHolders,
		"total_balance":     totalBalance.String(),
		"max_balance":       maxBalance.String(),
		"active_allowances": activeAllowances,
		"circulation":       totalBalance.String(),
		"burned":            new(big.Int).Sub(token.TotalSupply, totalBalance).String(),
	}, nil
}

// SearchTokens ищет токены по критериям
func (em *ERC20Manager) SearchTokens(criteria map[string]interface{}) []*ERC20Token {
	results := make([]*ERC20Token, 0)

	for _, token := range em.tokens {
		matches := true

		// Фильтр по имени
		if name, ok := criteria["name"].(string); ok && name != "" {
			if token.Name != name {
				matches = false
			}
		}

		// Фильтр по символу
		if symbol, ok := criteria["symbol"].(string); ok && symbol != "" {
			if token.Symbol != symbol {
				matches = false
			}
		}

		// Фильтр по владельцу
		if owner, ok := criteria["owner"].(string); ok && owner != "" {
			if token.Owner != owner {
				matches = false
			}
		}

		// Фильтр по минимальному общему предложению
		if minSupply, ok := criteria["min_supply"].(*big.Int); ok && minSupply != nil {
			if token.TotalSupply.Cmp(minSupply) < 0 {
				matches = false
			}
		}

		if matches {
			results = append(results, token)
		}
	}

	return results
}

// ExportToken экспортирует токен в JSON
func (em *ERC20Manager) ExportToken(tokenAddress string) ([]byte, error) {
	token, exists := em.GetToken(tokenAddress)
	if !exists {
		return nil, fmt.Errorf("token not found: %s", tokenAddress)
	}

	return json.MarshalIndent(token, "", "  ")
}

// ImportToken импортирует токен из JSON
func (em *ERC20Manager) ImportToken(data []byte) (*ERC20Token, error) {
	var token ERC20Token
	err := json.Unmarshal(data, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %v", err)
	}

	// Проверяем, что токен с таким адресом не существует
	if _, exists := em.tokens[token.Address]; exists {
		return nil, fmt.Errorf("token with address %s already exists", token.Address)
	}

	// Инициализируем карты, если они nil
	if token.Balances == nil {
		token.Balances = make(map[string]*big.Int)
	}
	if token.Allowances == nil {
		token.Allowances = make(map[string]map[string]*big.Int)
	}

	em.tokens[token.Address] = &token
	return &token, nil
}
