package mining

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/wallet"
)

// Miner представляет майнер блоков
type Miner struct {
	ID         string                 `json:"id"`
	Address    string                 `json:"address"`
	PublicKey  []byte                 `json:"public_key"`
	Blockchain *blockchain.Blockchain `json:"-"`
	Mempool    *Mempool               `json:"-"`
	Network    *network.Server        `json:"-"`
	Wallet     *wallet.Wallet         `json:"-"`
	IsMining   bool                   `json:"is_mining"`
	StopChan   chan bool              `json:"-"`
	MiningChan chan bool              `json:"-"`
	Stats      *MiningStats           `json:"stats"`
	mutex      sync.RWMutex
}

// MiningStats содержит статистику майнинга
type MiningStats struct {
	BlocksMined      int64         `json:"blocks_mined"`
	TotalHashes      int64         `json:"total_hashes"`
	StartTime        time.Time     `json:"start_time"`
	LastBlockTime    time.Time     `json:"last_block_time"`
	AverageBlockTime time.Duration `json:"average_block_time"`
	HashRate         float64       `json:"hash_rate"`
}

// NewMiner создает нового майнера
func NewMiner(address string, publicKey []byte, bc *blockchain.Blockchain, mempool *Mempool, net *network.Server, w *wallet.Wallet) *Miner {
	return &Miner{
		ID:         generateMinerID(),
		Address:    address,
		PublicKey:  publicKey,
		Blockchain: bc,
		Mempool:    mempool,
		Network:    net,
		Wallet:     w,
		IsMining:   false,
		StopChan:   make(chan bool),
		MiningChan: make(chan bool),
		Stats: &MiningStats{
			StartTime: time.Now(),
		},
	}
}

// StartMining начинает майнинг
func (m *Miner) StartMining() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.IsMining {
		return fmt.Errorf("miner is already running")
	}

	m.IsMining = true
	m.Stats.StartTime = time.Now()

	slog.Info("Starting miner", "miner_id", m.ID, "address", m.Address)

	// Запускаем майнинг в отдельной горутине
	go m.mine()

	return nil
}

// StopMining останавливает майнинг
func (m *Miner) StopMining() error {
	m.mutex.Lock()
	if !m.IsMining {
		m.mutex.Unlock()
		return fmt.Errorf("miner is not running")
	}

	m.IsMining = false
	m.mutex.Unlock()

	// Отправляем сигнал остановки без блокировки
	select {
	case m.StopChan <- true:
	default:
		// Канал уже заполнен, это нормально
	}

	slog.Info("Stopping miner", "miner_id", m.ID)
	return nil
}

// mine выполняет майнинг в отдельной горутине
func (m *Miner) mine() {
	for {
		// Проверяем, нужно ли остановиться
		select {
		case <-m.StopChan:
			return
		default:
		}

		// Проверяем статус майнинга
		m.mutex.RLock()
		isMining := m.IsMining
		m.mutex.RUnlock()

		if !isMining {
			return
		}

		// Создаем новый блок для майнинга
		block, err := m.createMiningBlock()
		if err != nil {
			slog.Error("Failed to create mining block", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Майним блок
		success, err := m.mineBlock(block)
		if err != nil {
			slog.Error("Mining failed", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if success {
			// Добавляем блок в блокчейн
			err = m.Blockchain.AddBlock(block)
			if err != nil {
				slog.Error("Failed to add mined block", "error", err)
				continue
			}

			// Обновляем статистику
			m.updateStats(block)

			// Распространяем блок по сети
			if m.Network != nil {
				m.Network.BroadcastNewBlock(block)
			}

			slog.Info("Block mined successfully",
				"height", block.Height,
				"hash", fmt.Sprintf("%x", block.Hash),
				"nonce", block.Nonce)
		}
	}
}

// createMiningBlock создает блок для майнинга
func (m *Miner) createMiningBlock() (*blockchain.Block, error) {
	// Получаем последний блок и сложность одним вызовом
	lastBlock, difficulty := m.Blockchain.GetLastBlockAndDifficulty()
	if lastBlock == nil {
		return nil, fmt.Errorf("no last block found")
	}

	// Создаем coinbase транзакцию
	coinbaseTx := blockchain.NewCoinbaseTransaction(m.Address, m.PublicKey, 50) // 50 монет награда

	// Получаем транзакции из mempool
	mempoolTxs := m.Mempool.GetTransactionsForBlock(100) // Максимум 100 транзакций

	// Создаем список транзакций (coinbase + mempool)
	transactions := []*blockchain.Transaction{coinbaseTx}
	transactions = append(transactions, mempoolTxs...)

	// Создаем блок
	block := blockchain.NewBlock(
		transactions,
		lastBlock.Hash,
		lastBlock.Height+1,
		difficulty,
	)

	return block, nil
}

// mineBlock майнит блок
func (m *Miner) mineBlock(block *blockchain.Block) (bool, error) {
	// Создаем оптимизированный Proof of Work
	pow := NewOptimizedProofOfWork(block, block.Difficulty)

	// Майним блок (используем параллельный майнинг с 2 воркерами)
	nonce, hash, success := pow.MineParallel(2)
	if !success {
		return false, fmt.Errorf("mining failed: no valid nonce found")
	}

	// Обновляем блок
	block.Nonce = int(nonce)
	block.Hash = hash

	// Обновляем статистику хешей
	m.mutex.Lock()
	m.Stats.TotalHashes += nonce
	m.mutex.Unlock()

	return true, nil
}

// updateStats обновляет статистику майнинга
func (m *Miner) updateStats(block *blockchain.Block) {
	// Используем неблокирующий подход для обновления статистики
	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.Stats.BlocksMined++
		m.Stats.LastBlockTime = time.Now()

		// Вычисляем среднее время блока
		if m.Stats.BlocksMined > 0 {
			elapsed := time.Since(m.Stats.StartTime)
			m.Stats.AverageBlockTime = elapsed / time.Duration(m.Stats.BlocksMined)
		}

		// Вычисляем hash rate
		if m.Stats.TotalHashes > 0 {
			elapsed := time.Since(m.Stats.StartTime)
			m.Stats.HashRate = float64(m.Stats.TotalHashes) / elapsed.Seconds()
		}
	}()
}

// GetStats возвращает статистику майнинга
func (m *Miner) GetStats() *MiningStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Создаем копию статистики
	stats := *m.Stats
	return &stats
}

// IsRunning проверяет, работает ли майнер
func (m *Miner) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.IsMining
}

// generateMinerID генерирует уникальный ID майнера
func generateMinerID() string {
	return fmt.Sprintf("miner_%d", time.Now().UnixNano())
}

// String возвращает строковое представление майнера
func (m *Miner) String() string {
	return fmt.Sprintf("Miner{ID: %s, Address: %s, Mining: %t, Blocks: %d}",
		m.ID, m.Address, m.IsMining, m.Stats.BlocksMined)
}
