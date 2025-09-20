package persistent

import (
	"fmt"
	"log/slog"
	"sync"

	"mirochain/internal/blockchain"
	"mirochain/internal/storage"
)

// PersistentBlockchain представляет блокчейн с персистентным хранением
type PersistentBlockchain struct {
	storage    storage.Storage
	UTXOSet    *blockchain.UTXOSet
	Difficulty int
	mutex      sync.RWMutex
}

// NewPersistentBlockchain создает новый блокчейн с персистентным хранением
func NewPersistentBlockchain(dataDir string, address string, publicKey []byte, difficulty int) (*PersistentBlockchain, error) {
	// Создаем storage
	store, err := storage.NewBadgerStorage(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	pbc := &PersistentBlockchain{
		storage:    store,
		Difficulty: difficulty,
	}

	// Пытаемся загрузить существующий блокчейн
	err = pbc.loadFromStorage()
	if err != nil {
		slog.Info("No existing blockchain found, creating new one")

		// Создаем новый UTXO набор
		pbc.UTXOSet = blockchain.NewUTXOSet()

		// Создаем genesis блок
		genesisBlock := pbc.createGenesisBlock(address, publicKey)

		// Сохраняем genesis блок
		if err := pbc.saveBlock(genesisBlock); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to save genesis block: %w", err)
		}

		// Обрабатываем genesis блок в UTXOSet
		if err := pbc.UTXOSet.ProcessBlock(genesisBlock); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to process genesis block: %w", err)
		}

		// Сохраняем UTXO набор
		if err := pbc.saveUTXOSet(); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to save UTXO set: %w", err)
		}

		// Сохраняем метаданные
		if err := pbc.saveMetadata(); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to save metadata: %w", err)
		}

		slog.Info("New blockchain created with genesis block")
	} else {
		// Проверяем, есть ли хотя бы genesis блок
		height, err := pbc.GetHeight()
		if err != nil || height < 0 {
			slog.Info("No blocks found in storage, creating genesis block")

			// Создаем genesis блок
			genesisBlock := pbc.createGenesisBlock(address, publicKey)

			// Сохраняем genesis блок
			if err := pbc.saveBlock(genesisBlock); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to save genesis block: %w", err)
			}

			// Обрабатываем genesis блок в UTXOSet
			if err := pbc.UTXOSet.ProcessBlock(genesisBlock); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to process genesis block: %w", err)
			}

			// Сохраняем UTXO набор
			if err := pbc.saveUTXOSet(); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to save UTXO set: %w", err)
			}

			// Сохраняем метаданные
			if err := pbc.saveMetadata(); err != nil {
				store.Close()
				return nil, fmt.Errorf("failed to save metadata: %w", err)
			}

			slog.Info("Genesis block created and saved")
		} else {
			slog.Info("Existing blockchain loaded from storage")
		}
	}

	return pbc, nil
}

// loadFromStorage загружает блокчейн из storage
func (pbc *PersistentBlockchain) loadFromStorage() error {
	// Загружаем UTXO набор
	utxoSet, err := pbc.storage.GetUTXOSet()
	if err != nil {
		return fmt.Errorf("failed to load UTXO set: %w", err)
	}
	pbc.UTXOSet = utxoSet

	// Загружаем сложность (если не найдена, используем значение по умолчанию)
	difficulty, err := pbc.storage.GetDifficulty()
	if err != nil {
		// Если difficulty не найдена, используем значение по умолчанию
		pbc.Difficulty = 0
	} else {
		pbc.Difficulty = difficulty
	}

	return nil
}

// createGenesisBlock создает genesis блок
func (pbc *PersistentBlockchain) createGenesisBlock(address string, publicKey []byte) *blockchain.Block {
	// Создаем coinbase транзакцию для genesis блока
	coinbaseTx := blockchain.NewCoinbaseTransaction(address, publicKey, 1000000) // 1M монет в genesis

	return blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, []byte{}, 0, pbc.Difficulty)
}

// AddBlock добавляет блок в блокчейн
func (pbc *PersistentBlockchain) AddBlock(block *blockchain.Block) error {
	pbc.mutex.Lock()
	defer pbc.mutex.Unlock()

	// Валидируем блок
	if !pbc.isValidBlock(block) {
		return fmt.Errorf("invalid block")
	}

	// Обрабатываем блок в UTXOSet
	if err := pbc.UTXOSet.ProcessBlock(block); err != nil {
		return fmt.Errorf("failed to process block in UTXO set: %w", err)
	}

	// Сохраняем блок
	if err := pbc.saveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %w", err)
	}

	// Сохраняем UTXO набор
	if err := pbc.saveUTXOSet(); err != nil {
		return fmt.Errorf("failed to save UTXO set: %w", err)
	}

	// Сохраняем метаданные
	if err := pbc.saveMetadata(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	slog.Info("Block added to persistent blockchain", "height", block.Height, "hash", fmt.Sprintf("%x", block.Hash))
	return nil
}

// isValidBlock проверяет валидность блока
func (pbc *PersistentBlockchain) isValidBlock(block *blockchain.Block) bool {
	// Получаем предыдущий блок напрямую из storage
	previousBlock, err := pbc.storage.GetBlockByHeight(block.Height - 1)
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
	height, err := pbc.storage.GetHeight()
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

// GetBlock получает блок по хешу
func (pbc *PersistentBlockchain) GetBlock(hash []byte) (*blockchain.Block, error) {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	return pbc.storage.GetBlock(hash)
}

// GetBlockByHeight получает блок по высоте
func (pbc *PersistentBlockchain) GetBlockByHeight(height int64) (*blockchain.Block, error) {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	return pbc.storage.GetBlockByHeight(height)
}

// GetGenesisHash получает хеш genesis блока
func (pbc *PersistentBlockchain) GetGenesisHash() []byte {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	genesisBlock, err := pbc.storage.GetBlockByHeight(0)
	if err != nil {
		slog.Error("Failed to get genesis block", "error", err)
		return nil
	}

	return genesisBlock.Hash
}

// GetHeight получает высоту блокчейна
func (pbc *PersistentBlockchain) GetHeight() (int64, error) {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	return pbc.storage.GetHeight()
}

// GetStats получает статистику блокчейна
func (pbc *PersistentBlockchain) GetStats() map[string]interface{} {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	height, _ := pbc.storage.GetHeight()
	difficulty, _ := pbc.storage.GetDifficulty()

	return map[string]interface{}{
		"height":     height,
		"difficulty": difficulty,
		"utxo_count": len(pbc.UTXOSet.UTXOs),
	}
}

// GetBalance получает баланс адреса
func (pbc *PersistentBlockchain) GetBalance(address string) int64 {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	return pbc.UTXOSet.GetBalance(address)
}

// GetUTXOs получает UTXO для адреса
func (pbc *PersistentBlockchain) GetUTXOs(address string) []*blockchain.UTXO {
	pbc.mutex.RLock()
	defer pbc.mutex.RUnlock()

	return pbc.UTXOSet.GetUTXOsByAddress(address)
}

// Close закрывает блокчейн и storage
func (pbc *PersistentBlockchain) Close() error {
	pbc.mutex.Lock()
	defer pbc.mutex.Unlock()

	if pbc.storage != nil {
		err := pbc.storage.Close()
		pbc.storage = nil
		return err
	}
	return nil
}

// Вспомогательные методы для сохранения

func (pbc *PersistentBlockchain) saveBlock(block *blockchain.Block) error {
	return pbc.storage.SaveBlock(block)
}

func (pbc *PersistentBlockchain) saveUTXOSet() error {
	return pbc.storage.SaveUTXOSet(pbc.UTXOSet)
}

func (pbc *PersistentBlockchain) saveMetadata() error {
	height, err := pbc.storage.GetHeight()
	if err != nil {
		return err
	}

	if err := pbc.storage.SaveHeight(height); err != nil {
		return err
	}

	return pbc.storage.SaveDifficulty(pbc.Difficulty)
}
