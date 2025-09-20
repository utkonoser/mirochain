package persistent

import (
	"fmt"
	"log/slog"
	"sync"

	"mirochain/internal/blockchain"
	"mirochain/internal/cache"
	"mirochain/internal/storage"
)

// CachedPersistentBlockchain представляет блокчейн с кэшированием
type CachedPersistentBlockchain struct {
	storage    storage.Storage
	cache      cache.Cache
	UTXOSet    *blockchain.UTXOSet
	Difficulty int
	mutex      sync.RWMutex
}

// NewCachedPersistentBlockchain создает новый блокчейн с кэшированием
func NewCachedPersistentBlockchain(dataDir string, address string, publicKey []byte, difficulty int) (*CachedPersistentBlockchain, error) {
	// Создаем storage
	store, err := storage.NewBadgerStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Создаем кэш
	cache := cache.NewLRUCache(1000, 500, 500, 0) // 0 TTL означает бесконечное время жизни

	cpbc := &CachedPersistentBlockchain{
		storage:    store,
		cache:      cache,
		Difficulty: difficulty,
	}

	// Пытаемся загрузить существующий блокчейн
	err = cpbc.loadFromStorage()
	if err != nil {
		slog.Info("No existing blockchain found, creating new one")

		// Создаем новый UTXO набор
		cpbc.UTXOSet = blockchain.NewUTXOSet()

		// Создаем genesis блок
		genesisBlock := cpbc.createGenesisBlock(address, publicKey)

		// Сохраняем genesis блок
		if err := cpbc.saveBlock(genesisBlock); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to save genesis block: %w", err)
		}

		// Обрабатываем genesis блок в UTXOSet
		if err := cpbc.UTXOSet.ProcessBlock(genesisBlock); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to process genesis block: %w", err)
		}

		// Сохраняем UTXO набор
		if err := cpbc.saveUTXOSet(); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to save UTXO set: %w", err)
		}

		// Сохраняем метаданные
		if err := cpbc.saveMetadata(); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to save metadata: %w", err)
		}

		slog.Info("New cached blockchain created with genesis block")
	} else {
		// Проверяем, есть ли хотя бы genesis блок
		height, err := cpbc.storage.GetHeight()
		if err != nil || height < 0 {
			slog.Info("No blocks found in storage, creating genesis block")

			// Создаем genesis блок
			genesisBlock := cpbc.createGenesisBlock(address, publicKey)

			// Сохраняем genesis блок
			if err := cpbc.saveBlock(genesisBlock); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to save genesis block: %w", err)
			}

			// Обрабатываем genesis блок в UTXOSet
			if err := cpbc.UTXOSet.ProcessBlock(genesisBlock); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to process genesis block: %w", err)
			}

			// Сохраняем UTXO набор
			if err := cpbc.saveUTXOSet(); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to save UTXO set: %w", err)
			}

			// Сохраняем метаданные
			if err := cpbc.saveMetadata(); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to save metadata: %w", err)
			}

			slog.Info("Genesis block created and saved")
		} else {
			slog.Info("Existing cached blockchain loaded from storage")
		}
	}

	return cpbc, nil
}

// loadFromStorage загружает блокчейн из storage
func (cpbc *CachedPersistentBlockchain) loadFromStorage() error {
	// Загружаем UTXO набор
	utxoSet, err := cpbc.storage.GetUTXOSet()
	if err != nil {
		return fmt.Errorf("failed to load UTXO set: %w", err)
	}
	cpbc.UTXOSet = utxoSet

	// Загружаем сложность (если не найдена, используем значение по умолчанию)
	difficulty, err := cpbc.storage.GetDifficulty()
	if err != nil {
		// Если difficulty не найдена, используем значение по умолчанию
		cpbc.Difficulty = 0
	} else {
		cpbc.Difficulty = difficulty
	}

	// Загружаем высоту в кэш
	height, err := cpbc.storage.GetHeight()
	if err == nil {
		cpbc.cache.SetHeight(height)
	}

	return nil
}

// createGenesisBlock создает genesis блок
func (cpbc *CachedPersistentBlockchain) createGenesisBlock(address string, publicKey []byte) *blockchain.Block {
	// Создаем coinbase транзакцию для genesis блока
	coinbaseTx := blockchain.NewCoinbaseTransaction(address, publicKey, 1000000) // 1M монет в genesis

	return blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, 0, cpbc.Difficulty)
}

// AddBlock добавляет блок в блокчейн
func (cpbc *CachedPersistentBlockchain) AddBlock(block *blockchain.Block) error {
	cpbc.mutex.Lock()
	defer cpbc.mutex.Unlock()

	// Валидируем блок
	if !cpbc.isValidBlock(block) {
		return fmt.Errorf("invalid block")
	}

	// Обрабатываем блок в UTXOSet
	if err := cpbc.UTXOSet.ProcessBlock(block); err != nil {
		return fmt.Errorf("failed to process block in UTXO set: %w", err)
	}

	// Сохраняем блок
	if err := cpbc.saveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	// Сохраняем UTXO набор
	if err := cpbc.saveUTXOSet(); err != nil {
		return fmt.Errorf("failed to save UTXO set: %w", err)
	}

	// Сохраняем метаданные
	if err := cpbc.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	slog.Info("Block added to cached persistent blockchain", "height", block.Height, "hash", fmt.Sprintf("%x", block.Hash))
	return nil
}

// isValidBlock проверяет валидность блока
func (cpbc *CachedPersistentBlockchain) isValidBlock(block *blockchain.Block) bool {
	// Получаем предыдущий блок напрямую из storage
	previousBlock, err := cpbc.storage.GetBlockByHeight(block.Height - 1)
	if err != nil && block.Height > 0 {
		slog.Error("Failed to get previous block", "error", err)
		return false
	}

	// Проверяем валидность блока
	if !block.IsValid(previousBlock) {
		slog.Error("Block validation failed")
		return false
	}

	// Проверяем, что высота блока корректна
	height, err := cpbc.storage.GetHeight()
	if err != nil {
		slog.Error("Failed to get blockchain height", "error", err)
		return false
	}

	if block.Height != height+1 {
		slog.Error("Invalid block height", "expected", height+1, "actual", block.Height)
		return false
	}

	return true
}

// GetBlock получает блок по хешу (с кэшированием)
func (cpbc *CachedPersistentBlockchain) GetBlock(hash []byte) (*blockchain.Block, error) {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	// Пытаемся получить из кэша
	if block, found := cpbc.cache.GetBlock(hash); found {
		return block, nil
	}

	// Если не найдено в кэше, загружаем из storage
	block, err := cpbc.storage.GetBlock(hash)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	cpbc.cache.SetBlock(hash, block)

	return block, nil
}

// GetBlockByHeight получает блок по высоте (с кэшированием)
func (cpbc *CachedPersistentBlockchain) GetBlockByHeight(height int64) (*blockchain.Block, error) {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	// Пытаемся получить из кэша
	if block, found := cpbc.cache.GetBlockByHeight(height); found {
		return block, nil
	}

	// Если не найдено в кэше, загружаем из storage
	block, err := cpbc.storage.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	cpbc.cache.SetBlockByHeight(height, block)

	return block, nil
}

// GetGenesisHash получает хеш genesis блока
func (cpbc *CachedPersistentBlockchain) GetGenesisHash() []byte {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	genesisBlock, err := cpbc.storage.GetBlockByHeight(0)
	if err != nil {
		slog.Error("Failed to get genesis block", "error", err)
		return nil
	}

	return genesisBlock.Hash
}

// GetHeight получает высоту блокчейна (с кэшированием)
func (cpbc *CachedPersistentBlockchain) GetHeight() (int64, error) {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	// Пытаемся получить из кэша
	if height, found := cpbc.cache.GetHeight(); found {
		return height, nil
	}

	// Если не найдено в кэше, загружаем из storage
	height, err := cpbc.storage.GetHeight()
	if err != nil {
		return 0, err
	}

	// Сохраняем в кэш
	cpbc.cache.SetHeight(height)

	return height, nil
}

// GetStats получает статистику блокчейна
func (cpbc *CachedPersistentBlockchain) GetStats() map[string]interface{} {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	height, _ := cpbc.storage.GetHeight()
	difficulty, _ := cpbc.storage.GetDifficulty()
	cacheStats := cpbc.cache.Stats()

	return map[string]interface{}{
		"height":       height,
		"difficulty":   difficulty,
		"utxo_count":   len(cpbc.UTXOSet.UTXOs),
		"cache_hits":   cacheStats.Hits,
		"cache_misses": cacheStats.Misses,
		"cache_size":   cacheStats.Size,
	}
}

// GetBalance получает баланс адреса (с кэшированием)
func (cpbc *CachedPersistentBlockchain) GetBalance(address string) int64 {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	// Пытаемся получить из кэша
	if balance, found := cpbc.cache.GetBalance(address); found {
		return balance
	}

	// Если не найдено в кэше, вычисляем
	balance := cpbc.UTXOSet.GetBalance(address)

	// Сохраняем в кэш
	cpbc.cache.SetBalance(address, balance)

	return balance
}

// GetUTXOs получает UTXO для адреса (с кэшированием)
func (cpbc *CachedPersistentBlockchain) GetUTXOs(address string) []*blockchain.UTXO {
	cpbc.mutex.RLock()
	defer cpbc.mutex.RUnlock()

	// Пытаемся получить из кэша
	if utxos, found := cpbc.cache.GetUTXOs(address); found {
		return utxos
	}

	// Если не найдено в кэше, получаем из UTXOSet
	utxos := cpbc.UTXOSet.GetUTXOsByAddress(address)

	// Сохраняем в кэш
	cpbc.cache.SetUTXOs(address, utxos)

	return utxos
}

// Close закрывает блокчейн и storage
func (cpbc *CachedPersistentBlockchain) Close() error {
	cpbc.mutex.Lock()
	defer cpbc.mutex.Unlock()

	if cpbc.storage != nil {
		err := cpbc.storage.Close()
		cpbc.storage = nil
		return err
	}
	return nil
}

// GetStorage возвращает storage для использования в других компонентах
func (cpbc *CachedPersistentBlockchain) GetStorage() storage.Storage {
	return cpbc.storage
}

// Вспомогательные методы для сохранения

func (cpbc *CachedPersistentBlockchain) saveBlock(block *blockchain.Block) error {
	// Сохраняем в storage
	err := cpbc.storage.SaveBlock(block)
	if err != nil {
		return err
	}

	// Сохраняем в кэш
	cpbc.cache.SetBlock(block.Hash, block)
	cpbc.cache.SetBlockByHeight(block.Height, block)

	return nil
}

func (cpbc *CachedPersistentBlockchain) saveUTXOSet() error {
	return cpbc.storage.SaveUTXOSet(cpbc.UTXOSet)
}

func (cpbc *CachedPersistentBlockchain) saveMetadata() error {
	height, err := cpbc.storage.GetHeight()
	if err != nil {
		return err
	}

	if err := cpbc.storage.SaveHeight(height); err != nil {
		return err
	}

	if err := cpbc.storage.SaveDifficulty(cpbc.Difficulty); err != nil {
		return err
	}

	// Обновляем кэш
	cpbc.cache.SetHeight(height)
	cpbc.cache.SetDifficulty(cpbc.Difficulty)

	return nil
}
