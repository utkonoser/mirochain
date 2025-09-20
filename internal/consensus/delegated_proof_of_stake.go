package consensus

import (
	"fmt"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// DelegatedProofOfStake представляет алгоритм консенсуса Delegated Proof of Stake
type DelegatedProofOfStake struct {
	blockchain     *blockchain.Blockchain
	delegates      map[string]*Delegate
	voters         map[string]*Voter
	totalVotes     *big.Int
	activeDelegates []*Delegate
	roundLength    int64
	currentRound   int64
	mutex          sync.RWMutex
	delegateMutex  sync.RWMutex
}

// Delegate представляет делегата
type Delegate struct {
	Address      string   `json:"address"`
	Votes        *big.Int `json:"votes"`
	IsActive     bool     `json:"is_active"`
	LastBlock    int64    `json:"last_block"`
	MissedBlocks int      `json:"missed_blocks"`
	Rewards      *big.Int `json:"rewards"`
	Voters       []string `json:"voters"`
	Rank         int      `json:"rank"`
}

// Voter представляет голосующего
type Voter struct {
	Address        string   `json:"address"`
	VotePower      *big.Int `json:"vote_power"`
	VotedFor       string   `json:"voted_for"`
	LastVote       int64    `json:"last_vote"`
	DelegatedVotes *big.Int `json:"delegated_votes"`
}

// VoteTransaction представляет транзакцию голосования
type VoteTransaction struct {
	Type      string   `json:"type"`      // vote, unvote, delegate
	From      string   `json:"from"`
	To        string   `json:"to"`
	Amount    *big.Int `json:"amount"`
	Timestamp int64    `json:"timestamp"`
	Signature string   `json:"signature"`
}

// NewDelegatedProofOfStake создает новый DPoS консенсус
func NewDelegatedProofOfStake(bc *blockchain.Blockchain) *DelegatedProofOfStake {
	return &DelegatedProofOfStake{
		blockchain:     bc,
		delegates:      make(map[string]*Delegate),
		voters:         make(map[string]*Voter),
		totalVotes:     big.NewInt(0),
		activeDelegates: make([]*Delegate, 0),
		roundLength:    21, // 21 делегат на раунд
		currentRound:   0,
	}
}

// RegisterDelegate регистрирует делегата
func (dpos *DelegatedProofOfStake) RegisterDelegate(address string) error {
	dpos.mutex.Lock()
	defer dpos.mutex.Unlock()
	
	if _, exists := dpos.delegates[address]; exists {
		return fmt.Errorf("delegate %s already registered", address)
	}
	
	dpos.delegates[address] = &Delegate{
		Address:      address,
		Votes:        big.NewInt(0),
		IsActive:     false,
		LastBlock:    0,
		MissedBlocks: 0,
		Rewards:      big.NewInt(0),
		Voters:       make([]string, 0),
		Rank:         0,
	}
	
	log.Printf("Delegate registered: %s", address)
	return nil
}

// Vote голосует за делегата
func (dpos *DelegatedProofOfStake) Vote(voterAddress, delegateAddress string, votePower *big.Int) error {
	dpos.mutex.Lock()
	defer dpos.mutex.Unlock()
	
	if votePower.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("vote power must be positive")
	}
	
	// Проверяем, что делегат существует
	delegate, exists := dpos.delegates[delegateAddress]
	if !exists {
		return fmt.Errorf("delegate %s not found", delegateAddress)
	}
	
	// Получаем или создаем голосующего
	voter, exists := dpos.voters[voterAddress]
	if !exists {
		voter = &Voter{
			Address:        voterAddress,
			VotePower:      big.NewInt(0),
			VotedFor:       "",
			LastVote:       0,
			DelegatedVotes: big.NewInt(0),
		}
		dpos.voters[voterAddress] = voter
	}
	
	// Если голосующий уже голосовал за другого делегата, отменяем предыдущий голос
	if voter.VotedFor != "" && voter.VotedFor != delegateAddress {
		dpos.unvote(voterAddress, voter.VotedFor, voter.VotePower)
	}
	
	// Добавляем голос
	voter.VotePower.Add(voter.VotePower, votePower)
	voter.VotedFor = delegateAddress
	voter.LastVote = time.Now().Unix()
	
	// Обновляем голоса делегата
	delegate.Votes.Add(delegate.Votes, votePower)
	dpos.totalVotes.Add(dpos.totalVotes, votePower)
	
	// Добавляем голосующего в список делегата
	if !dpos.contains(delegate.Voters, voterAddress) {
		delegate.Voters = append(delegate.Voters, voterAddress)
	}
	
	// Обновляем активных делегатов
	dpos.updateActiveDelegates()
	
	log.Printf("Vote cast: %s voted %s for %s", voterAddress, votePower.String(), delegateAddress)
	return nil
}

// Unvote отменяет голос
func (dpos *DelegatedProofOfStake) Unvote(voterAddress string) error {
	dpos.mutex.Lock()
	defer dpos.mutex.Unlock()
	
	voter, exists := dpos.voters[voterAddress]
	if !exists {
		return fmt.Errorf("voter %s not found", voterAddress)
	}
	
	if voter.VotedFor == "" {
		return fmt.Errorf("voter %s has no active vote", voterAddress)
	}
	
	// Отменяем голос
	dpos.unvote(voterAddress, voter.VotedFor, voter.VotePower)
	
	// Сбрасываем голосующего
	voter.VotePower = big.NewInt(0)
	voter.VotedFor = ""
	voter.LastVote = time.Now().Unix()
	
	// Обновляем активных делегатов
	dpos.updateActiveDelegates()
	
	log.Printf("Vote cancelled: %s", voterAddress)
	return nil
}

// unvote внутренняя функция для отмены голоса
func (dpos *DelegatedProofOfStake) unvote(voterAddress, delegateAddress string, votePower *big.Int) {
	delegate := dpos.delegates[delegateAddress]
	if delegate != nil {
		delegate.Votes.Sub(delegate.Votes, votePower)
		dpos.totalVotes.Sub(dpos.totalVotes, votePower)
		
		// Удаляем голосующего из списка делегата
		for i, voter := range delegate.Voters {
			if voter == voterAddress {
				delegate.Voters = append(delegate.Voters[:i], delegate.Voters[i+1:]...)
				break
			}
		}
	}
}

// contains проверяет, содержит ли слайс элемент
func (dpos *DelegatedProofOfStake) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// updateActiveDelegates обновляет список активных делегатов
func (dpos *DelegatedProofOfStake) updateActiveDelegates() {
	dpos.delegateMutex.Lock()
	defer dpos.delegateMutex.Unlock()
	
	// Очищаем текущих активных делегатов
	dpos.activeDelegates = make([]*Delegate, 0)
	
	// Собираем всех делегатов с голосами
	candidates := make([]*Delegate, 0)
	for _, delegate := range dpos.delegates {
		if delegate.Votes.Cmp(big.NewInt(0)) > 0 {
			candidates = append(candidates, delegate)
		}
	}
	
	// Сортируем по количеству голосов
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Votes.Cmp(candidates[j].Votes) > 0
	})
	
	// Берем топ делегатов
	maxDelegates := int(dpos.roundLength)
	if len(candidates) > maxDelegates {
		candidates = candidates[:maxDelegates]
	}
	
	// Активируем делегатов
	for i, delegate := range candidates {
		delegate.IsActive = true
		delegate.Rank = i + 1
		dpos.activeDelegates = append(dpos.activeDelegates, delegate)
	}
	
	// Деактивируем остальных
	for _, delegate := range dpos.delegates {
		if !dpos.isActiveDelegate(delegate.Address) {
			delegate.IsActive = false
			delegate.Rank = 0
		}
	}
	
	log.Printf("Updated active delegates: %d active delegates", len(dpos.activeDelegates))
}

// isActiveDelegate проверяет, является ли делегат активным
func (dpos *DelegatedProofOfStake) isActiveDelegate(address string) bool {
	for _, delegate := range dpos.activeDelegates {
		if delegate.Address == address {
			return true
		}
	}
	return false
}

// SelectDelegate выбирает делегата для следующего блока
func (dpos *DelegatedProofOfStake) SelectDelegate(blockHeight int64) (*Delegate, error) {
	dpos.delegateMutex.RLock()
	defer dpos.delegateMutex.RUnlock()
	
	if len(dpos.activeDelegates) == 0 {
		return nil, fmt.Errorf("no active delegates available")
	}
	
	// Вычисляем раунд
	round := blockHeight / dpos.roundLength
	
	// Если раунд изменился, обновляем активных делегатов
	if round != dpos.currentRound {
		dpos.currentRound = round
		dpos.updateActiveDelegates()
	}
	
	// Выбираем делегата по позиции в раунде
	position := blockHeight % dpos.roundLength
	if int(position) >= len(dpos.activeDelegates) {
		position = 0
	}
	
	delegate := dpos.activeDelegates[position]
	
	// Обновляем информацию о делегате
	delegate.LastBlock = blockHeight
	delegate.MissedBlocks = 0
	
	return delegate, nil
}

// ValidateBlock валидирует блок в контексте DPoS
func (dpos *DelegatedProofOfStake) ValidateBlock(block *blockchain.Block) error {
	dpos.delegateMutex.RLock()
	defer dpos.delegateMutex.RUnlock()
	
	// Проверяем, что блок создан активным делегатом
	// В реальной реализации здесь должна быть проверка по хешу блока или другим полям
	delegate := dpos.findDelegateByAddress("delegate_1") // Заглушка
	if delegate == nil {
		return fmt.Errorf("block created by non-delegate")
	}
	
	if !delegate.IsActive {
		return fmt.Errorf("delegate is not active")
	}
	
	// Проверяем, что делегат не создавал блок слишком часто
	if block.Height-delegate.LastBlock < 2 {
		return fmt.Errorf("delegate created block too recently")
	}
	
	// Обновляем информацию о делегате
	delegate.LastBlock = block.Height
	delegate.MissedBlocks = 0
	
	return nil
}

// findDelegateByAddress находит делегата по адресу
func (dpos *DelegatedProofOfStake) findDelegateByAddress(address string) *Delegate {
	for _, delegate := range dpos.activeDelegates {
		if delegate.Address == address {
			return delegate
		}
	}
	return nil
}

// CalculateRewards вычисляет награды для делегатов
func (dpos *DelegatedProofOfStake) CalculateRewards(block *blockchain.Block) *big.Int {
	dpos.delegateMutex.RLock()
	defer dpos.delegateMutex.RUnlock()
	
	// Базовая награда за блок
	baseReward := big.NewInt(50)
	
	// Находим делегата
	// В реальной реализации здесь должна быть проверка по хешу блока или другим полям
	delegate := dpos.findDelegateByAddress("delegate_1") // Заглушка
	if delegate == nil {
		return big.NewInt(0)
	}
	
	// Вычисляем награду на основе голосов
	totalVotes := dpos.calculateTotalVotes()
	if totalVotes.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// Награда пропорциональна голосам
	reward := new(big.Int).Mul(baseReward, delegate.Votes)
	reward.Div(reward, totalVotes)
	
	// Добавляем награду к голосам делегата
	delegate.Rewards.Add(delegate.Rewards, reward)
	
	// Распределяем награду между голосующими
	dpos.distributeRewards(delegate, reward)
	
	return reward
}

// calculateTotalVotes вычисляет общее количество голосов
func (dpos *DelegatedProofOfStake) calculateTotalVotes() *big.Int {
	total := big.NewInt(0)
	for _, delegate := range dpos.activeDelegates {
		if delegate.IsActive {
			total.Add(total, delegate.Votes)
		}
	}
	return total
}

// distributeRewards распределяет награды между голосующими
func (dpos *DelegatedProofOfStake) distributeRewards(delegate *Delegate, totalReward *big.Int) {
	if delegate.Votes.Cmp(big.NewInt(0)) == 0 {
		return
	}
	
	// Распределяем награду пропорционально голосам
	for _, voterAddress := range delegate.Voters {
		voter := dpos.voters[voterAddress]
		if voter != nil {
			// Вычисляем долю голосующего
			share := new(big.Int).Mul(totalReward, voter.VotePower)
			share.Div(share, delegate.Votes)
			
			// Добавляем награду к делегированным голосам
			voter.DelegatedVotes.Add(voter.DelegatedVotes, share)
		}
	}
}

// GetDelegateInfo возвращает информацию о делегате
func (dpos *DelegatedProofOfStake) GetDelegateInfo(address string) (*Delegate, error) {
	dpos.mutex.RLock()
	defer dpos.mutex.RUnlock()
	
	delegate, exists := dpos.delegates[address]
	if !exists {
		return nil, fmt.Errorf("delegate %s not found", address)
	}
	
	return delegate, nil
}

// GetVoterInfo возвращает информацию о голосующем
func (dpos *DelegatedProofOfStake) GetVoterInfo(address string) (*Voter, error) {
	dpos.mutex.RLock()
	defer dpos.mutex.RUnlock()
	
	voter, exists := dpos.voters[address]
	if !exists {
		return nil, fmt.Errorf("voter %s not found", address)
	}
	
	return voter, nil
}

// GetActiveDelegates возвращает список активных делегатов
func (dpos *DelegatedProofOfStake) GetActiveDelegates() []*Delegate {
	dpos.delegateMutex.RLock()
	defer dpos.delegateMutex.RUnlock()
	
	// Возвращаем копию списка активных делегатов
	delegates := make([]*Delegate, len(dpos.activeDelegates))
	for i, delegate := range dpos.activeDelegates {
		delegates[i] = &Delegate{
			Address:      delegate.Address,
			Votes:        new(big.Int).Set(delegate.Votes),
			IsActive:     delegate.IsActive,
			LastBlock:    delegate.LastBlock,
			MissedBlocks: delegate.MissedBlocks,
			Rewards:      new(big.Int).Set(delegate.Rewards),
			Voters:       make([]string, len(delegate.Voters)),
			Rank:         delegate.Rank,
		}
		copy(delegates[i].Voters, delegate.Voters)
	}
	
	return delegates
}

// GetTotalVotes возвращает общее количество голосов
func (dpos *DelegatedProofOfStake) GetTotalVotes() *big.Int {
	dpos.mutex.RLock()
	defer dpos.mutex.RUnlock()
	
	return new(big.Int).Set(dpos.totalVotes)
}

// GetStats возвращает статистику DPoS
func (dpos *DelegatedProofOfStake) GetStats() map[string]interface{} {
	dpos.mutex.RLock()
	dpos.delegateMutex.RLock()
	defer dpos.mutex.RUnlock()
	defer dpos.delegateMutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_votes":        dpos.totalVotes.String(),
		"total_delegates":    len(dpos.delegates),
		"active_delegates":   len(dpos.activeDelegates),
		"total_voters":       len(dpos.voters),
		"current_round":      dpos.currentRound,
		"round_length":       dpos.roundLength,
	}
	
	// Добавляем информацию об активных делегатах
	delegates := make([]map[string]interface{}, 0, len(dpos.activeDelegates))
	for _, delegate := range dpos.activeDelegates {
		delegates = append(delegates, map[string]interface{}{
			"address":       delegate.Address,
			"votes":         delegate.Votes.String(),
			"is_active":     delegate.IsActive,
			"last_block":    delegate.LastBlock,
			"missed_blocks": delegate.MissedBlocks,
			"rewards":       delegate.Rewards.String(),
			"voters_count":  len(delegate.Voters),
			"rank":          delegate.Rank,
		})
	}
	stats["active_delegates"] = delegates
	
	return stats
}

// ProcessVoteTransaction обрабатывает транзакцию голосования
func (dpos *DelegatedProofOfStake) ProcessVoteTransaction(tx *VoteTransaction) error {
	switch tx.Type {
	case "vote":
		return dpos.Vote(tx.From, tx.To, tx.Amount)
	case "unvote":
		return dpos.Unvote(tx.From)
	case "delegate":
		return dpos.RegisterDelegate(tx.To)
	default:
		return fmt.Errorf("unknown vote transaction type: %s", tx.Type)
	}
}

// CreateVoteTransaction создает транзакцию голосования
func (dpos *DelegatedProofOfStake) CreateVoteTransaction(voteType, from, to string, amount *big.Int) *VoteTransaction {
	return &VoteTransaction{
		Type:      voteType,
		From:      from,
		To:        to,
		Amount:    new(big.Int).Set(amount),
		Timestamp: time.Now().Unix(),
		Signature: "", // Подпись должна быть добавлена отдельно
	}
}
