package mining

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/parallel"
	"mirochain/internal/wallet"
)

// Manager управляет майнерами и mempool
type Manager struct {
	Blockchain           *blockchain.Blockchain         `json:"-"`
	Mempool              *Mempool                       `json:"-"`
	Network              *network.Server                `json:"-"`
	Miners               map[string]*Miner              `json:"miners"`
	TransactionProcessor *parallel.TransactionProcessor `json:"-"`
	IsRunning            bool                           `json:"is_running"`
	mutex                sync.RWMutex
}

// NewManager создает новый менеджер майнинга
func NewManager(bc *blockchain.Blockchain, mempool *Mempool, net *network.Server) *Manager {
	// Создаем TransactionProcessor с 4 воркерами
	processor := parallel.NewTransactionProcessor(4, 100)

	return &Manager{
		Blockchain:           bc,
		Mempool:              mempool,
		Network:              net,
		Miners:               make(map[string]*Miner),
		TransactionProcessor: processor,
		IsRunning:            false,
	}
}

// Start запускает менеджер майнинга
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.IsRunning {
		return fmt.Errorf("mining manager is already running")
	}

	// Создаем UTXOSet для TransactionProcessor
	utxoSet := blockchain.NewUTXOSet()

	// Запускаем TransactionProcessor
	if err := m.TransactionProcessor.Start(utxoSet); err != nil {
		return fmt.Errorf("failed to start transaction processor: %w", err)
	}

	// Запускаем TransactionProcessor для mempool
	if err := m.Mempool.StartProcessor(); err != nil {
		return fmt.Errorf("failed to start mempool processor: %w", err)
	}

	m.IsRunning = true
	slog.Info("Mining manager started")

	return nil
}

// Stop останавливает менеджер майнинга
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.IsRunning {
		return fmt.Errorf("mining manager is not running")
	}

	// Останавливаем всех майнеров
	for _, miner := range m.Miners {
		if miner.IsRunning() {
			miner.StopMining()
		}
	}

	// Останавливаем TransactionProcessor
	m.TransactionProcessor.Stop()

	// Останавливаем TransactionProcessor для mempool
	m.Mempool.StopProcessor()

	m.IsRunning = false
	slog.Info("Mining manager stopped")

	return nil
}

// AddMiner добавляет майнера
func (m *Manager) AddMiner(address string, publicKey []byte, w *wallet.Wallet) (*Miner, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Проверяем, что майнер с таким адресом не существует
	for _, miner := range m.Miners {
		if miner.Address == address {
			return nil, fmt.Errorf("miner with address %s already exists", address)
		}
	}

	// Создаем майнера
	miner := NewMiner(address, publicKey, m.Blockchain, m.Mempool, m.Network, w)
	m.Miners[miner.ID] = miner

	slog.Info("Miner added", "miner_id", miner.ID, "address", address)
	return miner, nil
}

// RemoveMiner удаляет майнера
func (m *Manager) RemoveMiner(minerID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	miner, exists := m.Miners[minerID]
	if !exists {
		return fmt.Errorf("miner not found: %s", minerID)
	}

	// Останавливаем майнера, если он работает
	if miner.IsRunning() {
		miner.StopMining()
	}

	delete(m.Miners, minerID)
	slog.Info("Miner removed", "miner_id", minerID)
	return nil
}

// GetMiner возвращает майнера по ID
func (m *Manager) GetMiner(minerID string) (*Miner, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	miner, exists := m.Miners[minerID]
	return miner, exists
}

// GetMiners возвращает всех майнеров
func (m *Manager) GetMiners() []*Miner {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var miners []*Miner
	for _, miner := range m.Miners {
		miners = append(miners, miner)
	}
	return miners
}

// StartMining запускает майнинг для всех майнеров
func (m *Manager) StartMining() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.Miners) == 0 {
		return fmt.Errorf("no miners available")
	}

	for _, miner := range m.Miners {
		if !miner.IsRunning() {
			err := miner.StartMining()
			if err != nil {
				slog.Error("Failed to start miner", "miner_id", miner.ID, "error", err)
			}
		}
	}

	slog.Info("Mining started for all miners")
	return nil
}

// StopMining останавливает майнинг для всех майнеров
func (m *Manager) StopMining() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, miner := range m.Miners {
		if miner.IsRunning() {
			err := miner.StopMining()
			if err != nil {
				slog.Error("Failed to stop miner", "miner_id", miner.ID, "error", err)
			}
		}
	}

	slog.Info("Mining stopped for all miners")
	return nil
}

// AddTransaction добавляет транзакцию в mempool
func (m *Manager) AddTransaction(tx *blockchain.Transaction) error {
	return m.Mempool.AddTransaction(tx)
}

// GetMempoolStats возвращает статистику mempool
func (m *Manager) GetMempoolStats() map[string]interface{} {
	return m.Mempool.GetStats()
}

// GetMiningStats возвращает общую статистику майнинга
func (m *Manager) GetMiningStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var totalBlocks int64
	var totalHashes int64
	var activeMiners int

	for _, miner := range m.Miners {
		stats := miner.GetStats()
		totalBlocks += stats.BlocksMined
		totalHashes += stats.TotalHashes
		if miner.IsRunning() {
			activeMiners++
		}
	}

	return map[string]interface{}{
		"total_blocks":  totalBlocks,
		"total_hashes":  totalHashes,
		"active_miners": activeMiners,
		"total_miners":  len(m.Miners),
		"mempool_size":  m.Mempool.Size(),
		"is_running":    m.IsRunning,
	}
}

// ProcessNewBlock обрабатывает новый блок (удаляет транзакции из mempool)
func (m *Manager) ProcessNewBlock(block *blockchain.Block) {
	// Получаем ID транзакций в блоке
	var txIDs []string
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			txIDs = append(txIDs, string(tx.ID))
		}
	}

	// Удаляем транзакции из mempool
	if len(txIDs) > 0 {
		m.Mempool.RemoveTransactions(txIDs)
		slog.Info("Removed transactions from mempool", "count", len(txIDs))
	}
}

// String возвращает строковое представление менеджера
func (m *Manager) String() string {
	return fmt.Sprintf("MiningManager{Running: %t, Miners: %d, Mempool: %s}",
		m.IsRunning, len(m.Miners), m.Mempool.String())
}

// ProcessTransactionAsync обрабатывает транзакцию асинхронно через TransactionProcessor
func (m *Manager) ProcessTransactionAsync(tx *blockchain.Transaction, utxoSet *blockchain.UTXOSet) <-chan *parallel.TransactionResult {
	resultChan := make(chan *parallel.TransactionResult, 1)

	go func() {
		result, err := m.TransactionProcessor.ProcessTransaction(tx, utxoSet, 1)
		if err != nil {
			result = &parallel.TransactionResult{
				Transaction: tx,
				Valid:       false,
				Error:       err,
				ProcessTime: 0,
			}
		}
		resultChan <- result
	}()

	return resultChan
}

// ProcessTransactionsBatch обрабатывает несколько транзакций параллельно
func (m *Manager) ProcessTransactionsBatch(transactions []*blockchain.Transaction, utxoSet *blockchain.UTXOSet) []*parallel.TransactionResult {
	results := make([]*parallel.TransactionResult, len(transactions))

	// Создаем каналы для результатов
	channels := make([]<-chan *parallel.TransactionResult, len(transactions))
	for i, tx := range transactions {
		channels[i] = m.ProcessTransactionAsync(tx, utxoSet)
	}

	// Собираем результаты
	for i, ch := range channels {
		select {
		case result := <-ch:
			results[i] = result
		case <-time.After(30 * time.Second):
			results[i] = &parallel.TransactionResult{
				Transaction: transactions[i],
				Valid:       false,
				Error:       fmt.Errorf("transaction processing timeout"),
				ProcessTime: 30 * time.Second,
			}
		}
	}

	return results
}

// GetTransactionProcessorStats возвращает статистику TransactionProcessor
func (m *Manager) GetTransactionProcessorStats() map[string]interface{} {
	return m.TransactionProcessor.GetStats()
}
