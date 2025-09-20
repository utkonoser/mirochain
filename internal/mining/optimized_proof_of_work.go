package mining

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"mirochain/internal/blockchain"
)

// OptimizedProofOfWork представляет оптимизированный алгоритм Proof of Work
type OptimizedProofOfWork struct {
	block      *blockchain.Block
	target     *big.Int
	Difficulty int
	Nonce      int64
	Hash       []byte
	Valid      bool
	stopChan   chan struct{}
}

// OptimizedMiner представляет оптимизированный майнер
type OptimizedMiner struct {
	workerCount int
	stopChan    chan struct{}
	mu          sync.RWMutex
	running     bool
	stats       *OptimizedMiningStats
}

// OptimizedMiningStats представляет статистику оптимизированного майнинга
type OptimizedMiningStats struct {
	BlocksMined     int64
	TotalTime       time.Duration
	AverageTime     time.Duration
	HashesPerSecond int64
	LastMined       time.Time
}

// NewOptimizedProofOfWork создает новый оптимизированный PoW
func NewOptimizedProofOfWork(block *blockchain.Block, difficulty int) *OptimizedProofOfWork {
	var target *big.Int
	if difficulty == 0 {
		// При нулевой сложности любое значение nonce должно быть валидным
		target = new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	} else {
		target = big.NewInt(1)
		target.Lsh(target, uint(256-difficulty))
	}

	return &OptimizedProofOfWork{
		block:      block,
		target:     target,
		Difficulty: difficulty,
		Nonce:      0,
		Hash:       nil,
		Valid:      false,
	}
}

// NewOptimizedMiner создает новый оптимизированный майнер
func NewOptimizedMiner(workerCount int) *OptimizedMiner {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	return &OptimizedMiner{
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
		running:     false,
		stats:       &OptimizedMiningStats{},
	}
}

// Mine выполняет майнинг блока с оптимизациями
func (pow *OptimizedProofOfWork) Mine() (int64, []byte, bool) {
	start := time.Now()

	// Предварительно вычисляем хеш блока без nonce
	blockData := pow.prepareBlockData()

	// Пробуем разные nonce значения
	for nonce := int64(0); nonce < 9223372036854775807; nonce++ {
		// Проверяем, нужно ли остановиться
		select {
		case <-pow.stopChan:
			return 0, nil, false
		default:
		}

		// Вычисляем хеш с текущим nonce
		hash := pow.calculateHash(blockData, nonce)

		// Проверяем, соответствует ли хеш цели
		if pow.isValidHash(hash) {
			pow.Nonce = nonce
			pow.Hash = hash
			pow.Valid = true

			elapsed := time.Since(start)
			slog.Info("Block mined successfully",
				"nonce", nonce,
				"hash", fmt.Sprintf("%x", hash),
				"time", elapsed,
				"difficulty", pow.Difficulty)

			return nonce, hash, true
		}
	}

	return 0, nil, false
}

// MineParallel выполняет параллельный майнинг
func (pow *OptimizedProofOfWork) MineParallel(workerCount int) (int64, []byte, bool) {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	start := time.Now()

	// Предварительно вычисляем хеш блока без nonce
	blockData := pow.prepareBlockData()

	// Каналы для координации
	resultChan := make(chan *MiningResult, 1)
	stopChan := make(chan struct{})

	// Запускаем воркеров
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go pow.mineWorker(i, workerCount, blockData, resultChan, stopChan, &wg)
	}

	// Ждем результат или остановку
	select {
	case result := <-resultChan:
		close(stopChan)
		wg.Wait()

		if result != nil {
			pow.Nonce = result.Nonce
			pow.Hash = result.Hash
			pow.Valid = true

			elapsed := time.Since(start)
			slog.Info("Block mined successfully (parallel)",
				"nonce", result.Nonce,
				"hash", fmt.Sprintf("%x", result.Hash),
				"time", elapsed,
				"difficulty", pow.Difficulty,
				"workers", workerCount)

			return result.Nonce, result.Hash, true
		}
	case <-pow.stopChan:
		close(stopChan)
		wg.Wait()
	}

	return 0, nil, false
}

// MiningResult представляет результат майнинга
type MiningResult struct {
	Nonce int64
	Hash  []byte
}

// mineWorker выполняет майнинг в отдельной горутине
func (pow *OptimizedProofOfWork) mineWorker(workerID, totalWorkers int, blockData []byte, resultChan chan<- *MiningResult, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// При нулевой сложности все воркеры начинают с nonce = 0
	var startNonce, step int64
	if pow.Difficulty == 0 {
		startNonce = 0
		step = 1
	} else {
		// Каждый воркер начинает с разного nonce
		startNonce = int64(workerID)
		step = int64(totalWorkers)
	}

	for nonce := startNonce; nonce < 9223372036854775807; nonce += step {
		// Проверяем, нужно ли остановиться
		select {
		case <-stopChan:
			return
		default:
		}

		// Вычисляем хеш с текущим nonce
		hash := pow.calculateHash(blockData, nonce)

		// Проверяем, соответствует ли хеш цели
		if pow.isValidHash(hash) {
			// Отправляем результат
			select {
			case resultChan <- &MiningResult{Nonce: nonce, Hash: hash}:
				return
			case <-stopChan:
				return
			}
		}
	}
}

// prepareBlockData подготавливает данные блока для хеширования
func (pow *OptimizedProofOfWork) prepareBlockData() []byte {
	// Используем тот же алгоритм, что и в обычном ProofOfWork
	return bytes.Join([][]byte{
		pow.block.PreviousHash,
		pow.block.MerkleRoot,
		[]byte(fmt.Sprintf("%d", pow.block.Timestamp)),
		[]byte(fmt.Sprintf("%d", pow.Difficulty)),
	}, []byte{})
}

// calculateHash вычисляет хеш блока
func (pow *OptimizedProofOfWork) calculateHash(blockData []byte, nonce int64) []byte {
	// Создаем копию данных блока
	data := make([]byte, len(blockData))
	copy(data, blockData)

	// Добавляем nonce
	nonceBytes := []byte(fmt.Sprintf("%d", nonce))
	data = append(data, nonceBytes...)

	// Вычисляем хеш
	hash := sha256.Sum256(data)
	return hash[:]
}

// isValidHash проверяет, соответствует ли хеш цели
func (pow *OptimizedProofOfWork) isValidHash(hash []byte) bool {
	// При нулевой сложности любой хеш валиден
	if pow.Difficulty == 0 {
		return true
	}

	var hashInt big.Int
	hashInt.SetBytes(hash)
	return hashInt.Cmp(pow.target) == -1
}

// Start запускает майнер
func (m *OptimizedMiner) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("miner is already running")
	}

	m.running = true
	m.stopChan = make(chan struct{})

	slog.Info("Optimized miner started", "workers", m.workerCount)
	return nil
}

// Stop останавливает майнер
func (m *OptimizedMiner) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("miner is not running")
	}

	close(m.stopChan)
	m.running = false

	slog.Info("Optimized miner stopped")
	return nil
}

// MineBlock майнит блок
func (m *OptimizedMiner) MineBlock(block *blockchain.Block) (int64, []byte, bool) {
	m.mu.RLock()
	if !m.running {
		m.mu.RUnlock()
		return 0, nil, false
	}
	m.mu.RUnlock()

	start := time.Now()

	// Создаем PoW
	pow := NewOptimizedProofOfWork(block, block.Difficulty)
	pow.stopChan = m.stopChan

	// Выполняем майнинг
	nonce, hash, success := pow.MineParallel(m.workerCount)

	if success {
		// Обновляем статистику
		elapsed := time.Since(start)
		atomic.AddInt64(&m.stats.BlocksMined, 1)
		atomic.AddInt64((*int64)(&m.stats.TotalTime), int64(elapsed))
		m.stats.LastMined = time.Now()

		// Вычисляем среднее время
		blocksMined := atomic.LoadInt64(&m.stats.BlocksMined)
		totalTime := time.Duration(atomic.LoadInt64((*int64)(&m.stats.TotalTime)))
		if blocksMined > 0 {
			m.stats.AverageTime = totalTime / time.Duration(blocksMined)
		}

		// Вычисляем хеши в секунду (приблизительно)
		if elapsed > 0 {
			seconds := elapsed.Nanoseconds() / 1000000000
			if seconds > 0 {
				hashesPerSecond := int64(1) / seconds
				atomic.StoreInt64(&m.stats.HashesPerSecond, hashesPerSecond)
			}
		}
	}

	return nonce, hash, success
}

// GetStats возвращает статистику майнера
func (m *OptimizedMiner) GetStats() *OptimizedMiningStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Возвращаем копию статистики
	return &OptimizedMiningStats{
		BlocksMined:     atomic.LoadInt64(&m.stats.BlocksMined),
		TotalTime:       time.Duration(atomic.LoadInt64((*int64)(&m.stats.TotalTime))),
		AverageTime:     m.stats.AverageTime,
		HashesPerSecond: atomic.LoadInt64(&m.stats.HashesPerSecond),
		LastMined:       m.stats.LastMined,
	}
}

// IsRunning проверяет, запущен ли майнер
func (m *OptimizedMiner) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}
