package consensus

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// ProofOfStake представляет алгоритм консенсуса Proof of Stake
type ProofOfStake struct {
	blockchain     *blockchain.Blockchain
	stakes         map[string]*StakeInfo
	totalStake     *big.Int
	validators     []*Validator
	epochLength    int64
	currentEpoch   int64
	mutex          sync.RWMutex
	validatorMutex sync.RWMutex
}

// StakeInfo содержит информацию о стейке
type StakeInfo struct {
	Address     string   `json:"address"`
	Amount      *big.Int `json:"amount"`
	LockTime    int64    `json:"lock_time"`
	LastStake   int64    `json:"last_stake"`
	DelegatedTo string   `json:"delegated_to,omitempty"`
	Rewards     *big.Int `json:"rewards"`
}

// Validator представляет валидатора
type Validator struct {
	Address      string   `json:"address"`
	Stake        *big.Int `json:"stake"`
	VotingPower  *big.Int `json:"voting_power"`
	IsActive     bool     `json:"is_active"`
	LastBlock    int64    `json:"last_block"`
	MissedBlocks int      `json:"missed_blocks"`
	Rewards      *big.Int `json:"rewards"`
}

// StakeTransaction представляет транзакцию стейкинга
type StakeTransaction struct {
	Type        string   `json:"type"`        // stake, unstake, delegate
	From        string   `json:"from"`
	To          string   `json:"to,omitempty"` // для делегирования
	Amount      *big.Int `json:"amount"`
	LockTime    int64    `json:"lock_time"`
	Timestamp   int64    `json:"timestamp"`
	Signature   string   `json:"signature"`
}

// NewProofOfStake создает новый PoS консенсус
func NewProofOfStake(bc *blockchain.Blockchain) *ProofOfStake {
	return &ProofOfStake{
		blockchain:  bc,
		stakes:      make(map[string]*StakeInfo),
		totalStake:  big.NewInt(0),
		validators:  make([]*Validator, 0),
		epochLength: 100, // 100 блоков на эпоху
		currentEpoch: 0,
	}
}

// Stake добавляет стейк
func (pos *ProofOfStake) Stake(address string, amount *big.Int, lockTime int64) error {
	pos.mutex.Lock()
	defer pos.mutex.Unlock()
	
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("stake amount must be positive")
	}
	
	if lockTime < 0 {
		return fmt.Errorf("lock time cannot be negative")
	}
	
	now := time.Now().Unix()
	
	// Проверяем, есть ли уже стейк для этого адреса
	if stake, exists := pos.stakes[address]; exists {
		// Добавляем к существующему стейку
		stake.Amount.Add(stake.Amount, amount)
		stake.LastStake = now
		
		// Обновляем время блокировки, если новое больше
		if lockTime > stake.LockTime {
			stake.LockTime = lockTime
		}
	} else {
		// Создаем новый стейк
		pos.stakes[address] = &StakeInfo{
			Address:   address,
			Amount:    new(big.Int).Set(amount),
			LockTime:  lockTime,
			LastStake: now,
			Rewards:   big.NewInt(0),
		}
	}
	
	// Обновляем общий стейк
	pos.totalStake.Add(pos.totalStake, amount)
	
	// Обновляем валидаторов
	pos.updateValidators()
	
	log.Printf("Stake added: %s staked %s tokens", address, amount.String())
	return nil
}

// Unstake снимает стейк
func (pos *ProofOfStake) Unstake(address string, amount *big.Int) error {
	pos.mutex.Lock()
	defer pos.mutex.Unlock()
	
	stake, exists := pos.stakes[address]
	if !exists {
		return fmt.Errorf("no stake found for address %s", address)
	}
	
	if amount.Cmp(stake.Amount) > 0 {
		return fmt.Errorf("unstake amount exceeds stake amount")
	}
	
	now := time.Now().Unix()
	if now < stake.LockTime {
		return fmt.Errorf("stake is still locked until %d", stake.LockTime)
	}
	
	// Уменьшаем стейк
	stake.Amount.Sub(stake.Amount, amount)
	pos.totalStake.Sub(pos.totalStake, amount)
	
	// Если стейк стал нулевым, удаляем запись
	if stake.Amount.Cmp(big.NewInt(0)) == 0 {
		delete(pos.stakes, address)
	}
	
	// Обновляем валидаторов
	pos.updateValidators()
	
	log.Printf("Stake removed: %s unstaked %s tokens", address, amount.String())
	return nil
}

// Delegate делегирует стейк
func (pos *ProofOfStake) Delegate(from, to string, amount *big.Int) error {
	pos.mutex.Lock()
	defer pos.mutex.Unlock()
	
	if from == to {
		return fmt.Errorf("cannot delegate to yourself")
	}
	
	stake, exists := pos.stakes[from]
	if !exists {
		return fmt.Errorf("no stake found for address %s", from)
	}
	
	if amount.Cmp(stake.Amount) > 0 {
		return fmt.Errorf("delegation amount exceeds stake amount")
	}
	
	// Уменьшаем стейк отправителя
	stake.Amount.Sub(stake.Amount, amount)
	stake.DelegatedTo = to
	
	// Увеличиваем стейк получателя
	if delegateStake, exists := pos.stakes[to]; exists {
		delegateStake.Amount.Add(delegateStake.Amount, amount)
	} else {
		pos.stakes[to] = &StakeInfo{
			Address:   to,
			Amount:    new(big.Int).Set(amount),
			LockTime:  stake.LockTime,
			LastStake: time.Now().Unix(),
			Rewards:   big.NewInt(0),
		}
	}
	
	// Обновляем валидаторов
	pos.updateValidators()
	
	log.Printf("Stake delegated: %s delegated %s tokens to %s", from, amount.String(), to)
	return nil
}

// updateValidators обновляет список валидаторов
func (pos *ProofOfStake) updateValidators() {
	pos.validatorMutex.Lock()
	defer pos.validatorMutex.Unlock()
	
	// Очищаем текущих валидаторов
	pos.validators = make([]*Validator, 0)
	
	// Добавляем валидаторов на основе стейка
	for address, stake := range pos.stakes {
		if stake.Amount.Cmp(big.NewInt(1000)) >= 0 { // минимум 1000 токенов для валидации
			validator := &Validator{
				Address:      address,
				Stake:        new(big.Int).Set(stake.Amount),
				VotingPower:  new(big.Int).Set(stake.Amount),
				IsActive:     true,
				LastBlock:    0,
				MissedBlocks: 0,
				Rewards:      new(big.Int).Set(stake.Rewards),
			}
			pos.validators = append(pos.validators, validator)
		}
	}
	
	// Сортируем валидаторов по стейку
	sort.Slice(pos.validators, func(i, j int) bool {
		return pos.validators[i].Stake.Cmp(pos.validators[j].Stake) > 0
	})
	
	log.Printf("Updated validators: %d active validators", len(pos.validators))
}

// SelectValidator выбирает валидатора для следующего блока
func (pos *ProofOfStake) SelectValidator(blockHeight int64) (*Validator, error) {
	pos.validatorMutex.RLock()
	defer pos.validatorMutex.RUnlock()
	
	if len(pos.validators) == 0 {
		return nil, fmt.Errorf("no validators available")
	}
	
	// Вычисляем эпоху
	epoch := blockHeight / pos.epochLength
	
	// Если эпоха изменилась, обновляем валидаторов
	if epoch != pos.currentEpoch {
		pos.currentEpoch = epoch
		pos.updateValidators()
	}
	
	// Выбираем валидатора на основе стейка и случайности
	validator := pos.selectValidatorByStake(blockHeight)
	
	return validator, nil
}

// selectValidatorByStake выбирает валидатора на основе стейка
func (pos *ProofOfStake) selectValidatorByStake(blockHeight int64) *Validator {
	if len(pos.validators) == 0 {
		return nil
	}
	
	// Вычисляем общий стейк
	totalStake := big.NewInt(0)
	for _, validator := range pos.validators {
		if validator.IsActive {
			totalStake.Add(totalStake, validator.Stake)
		}
	}
	
	if totalStake.Cmp(big.NewInt(0)) == 0 {
		return pos.validators[0]
	}
	
	// Создаем хеш для детерминированного выбора
	hash := pos.createSelectionHash(blockHeight, totalStake)
	
	// Выбираем валидатора на основе хеша
	selectionValue := new(big.Int).SetBytes(hash)
	selectionValue.Mod(selectionValue, totalStake)
	
	// Находим валидатора
	currentStake := big.NewInt(0)
	for _, validator := range pos.validators {
		if !validator.IsActive {
			continue
		}
		
		currentStake.Add(currentStake, validator.Stake)
		if selectionValue.Cmp(currentStake) < 0 {
			return validator
		}
	}
	
	// Fallback - возвращаем первого валидатора
	return pos.validators[0]
}

// createSelectionHash создает хеш для выбора валидатора
func (pos *ProofOfStake) createSelectionHash(blockHeight int64, totalStake *big.Int) []byte {
	// Используем высоту блока и общий стейк для создания детерминированного хеша
	data := fmt.Sprintf("%d:%s", blockHeight, totalStake.String())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// ValidateBlock валидирует блок в контексте PoS
func (pos *ProofOfStake) ValidateBlock(block *blockchain.Block) error {
	pos.validatorMutex.RLock()
	defer pos.validatorMutex.RUnlock()
	
	// Проверяем, что блок создан валидатором
	// В реальной реализации здесь должна быть проверка по хешу блока или другим полям
	validator := pos.findValidatorByAddress("validator_1") // Заглушка
	if validator == nil {
		return fmt.Errorf("block created by non-validator")
	}
	
	// Проверяем, что валидатор активен
	if !validator.IsActive {
		return fmt.Errorf("validator is not active")
	}
	
	// Проверяем, что валидатор не создавал блок слишком часто
	if block.Height-validator.LastBlock < 2 {
		return fmt.Errorf("validator created block too recently")
	}
	
	// Обновляем информацию о валидаторе
	validator.LastBlock = block.Height
	validator.MissedBlocks = 0
	
	return nil
}

// findValidatorByAddress находит валидатора по адресу
func (pos *ProofOfStake) findValidatorByAddress(address string) *Validator {
	for _, validator := range pos.validators {
		if validator.Address == address {
			return validator
		}
	}
	return nil
}

// CalculateRewards вычисляет награды для валидаторов
func (pos *ProofOfStake) CalculateRewards(block *blockchain.Block) *big.Int {
	pos.validatorMutex.RLock()
	defer pos.validatorMutex.RUnlock()
	
	// Базовая награда за блок
	baseReward := big.NewInt(50)
	
	// Находим валидатора
	// В реальной реализации здесь должна быть проверка по хешу блока или другим полям
	validator := pos.findValidatorByAddress("validator_1") // Заглушка
	if validator == nil {
		return big.NewInt(0)
	}
	
	// Вычисляем награду на основе стейка
	totalStake := pos.calculateTotalStake()
	if totalStake.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// Награда пропорциональна стейку
	reward := new(big.Int).Mul(baseReward, validator.Stake)
	reward.Div(reward, totalStake)
	
	// Добавляем награду к стейку валидатора
	validator.Rewards.Add(validator.Rewards, reward)
	
	// Обновляем стейк валидатора
	validator.Stake.Add(validator.Stake, reward)
	
	// Обновляем общий стейк
	pos.totalStake.Add(pos.totalStake, reward)
	
	return reward
}

// calculateTotalStake вычисляет общий стейк
func (pos *ProofOfStake) calculateTotalStake() *big.Int {
	total := big.NewInt(0)
	for _, validator := range pos.validators {
		if validator.IsActive {
			total.Add(total, validator.Stake)
		}
	}
	return total
}

// GetStakeInfo возвращает информацию о стейке
func (pos *ProofOfStake) GetStakeInfo(address string) (*StakeInfo, error) {
	pos.mutex.RLock()
	defer pos.mutex.RUnlock()
	
	stake, exists := pos.stakes[address]
	if !exists {
		return nil, fmt.Errorf("no stake found for address %s", address)
	}
	
	return stake, nil
}

// GetValidators возвращает список валидаторов
func (pos *ProofOfStake) GetValidators() []*Validator {
	pos.validatorMutex.RLock()
	defer pos.validatorMutex.RUnlock()
	
	// Возвращаем копию списка валидаторов
	validators := make([]*Validator, len(pos.validators))
	for i, validator := range pos.validators {
		validators[i] = &Validator{
			Address:      validator.Address,
			Stake:        new(big.Int).Set(validator.Stake),
			VotingPower:  new(big.Int).Set(validator.VotingPower),
			IsActive:     validator.IsActive,
			LastBlock:    validator.LastBlock,
			MissedBlocks: validator.MissedBlocks,
			Rewards:      new(big.Int).Set(validator.Rewards),
		}
	}
	
	return validators
}

// GetTotalStake возвращает общий стейк
func (pos *ProofOfStake) GetTotalStake() *big.Int {
	pos.mutex.RLock()
	defer pos.mutex.RUnlock()
	
	return new(big.Int).Set(pos.totalStake)
}

// GetStats возвращает статистику PoS
func (pos *ProofOfStake) GetStats() map[string]interface{} {
	pos.mutex.RLock()
	pos.validatorMutex.RLock()
	defer pos.mutex.RUnlock()
	defer pos.validatorMutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_stake":     pos.totalStake.String(),
		"active_stakes":   len(pos.stakes),
		"validators":      len(pos.validators),
		"current_epoch":   pos.currentEpoch,
		"epoch_length":    pos.epochLength,
	}
	
	// Добавляем информацию о валидаторах
	validators := make([]map[string]interface{}, 0, len(pos.validators))
	for _, validator := range pos.validators {
		validators = append(validators, map[string]interface{}{
			"address":       validator.Address,
			"stake":         validator.Stake.String(),
			"voting_power":  validator.VotingPower.String(),
			"is_active":     validator.IsActive,
			"last_block":    validator.LastBlock,
			"missed_blocks": validator.MissedBlocks,
			"rewards":       validator.Rewards.String(),
		})
	}
	stats["validators"] = validators
	
	return stats
}

// ProcessStakeTransaction обрабатывает транзакцию стейкинга
func (pos *ProofOfStake) ProcessStakeTransaction(tx *StakeTransaction) error {
	switch tx.Type {
	case "stake":
		return pos.Stake(tx.From, tx.Amount, tx.LockTime)
	case "unstake":
		return pos.Unstake(tx.From, tx.Amount)
	case "delegate":
		return pos.Delegate(tx.From, tx.To, tx.Amount)
	default:
		return fmt.Errorf("unknown stake transaction type: %s", tx.Type)
	}
}

// CreateStakeTransaction создает транзакцию стейкинга
func (pos *ProofOfStake) CreateStakeTransaction(stakeType, from, to string, amount *big.Int, lockTime int64) *StakeTransaction {
	return &StakeTransaction{
		Type:      stakeType,
		From:      from,
		To:        to,
		Amount:    new(big.Int).Set(amount),
		LockTime:  lockTime,
		Timestamp: time.Now().Unix(),
		Signature: "", // Подпись должна быть добавлена отдельно
	}
}
