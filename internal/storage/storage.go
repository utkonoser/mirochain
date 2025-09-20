package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"mirochain/internal/blockchain"

	"github.com/dgraph-io/badger/v4"
)

// BadgerStorage реализует Storage с использованием BadgerDB
type BadgerStorage struct {
	db   *badger.DB
	path string
	mu   sync.RWMutex
}

// NewBadgerStorage создает новый экземпляр BadgerStorage
func NewBadgerStorage(dataDir string) (*BadgerStorage, error) {
	// Создаем директорию для данных, если она не существует
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Настраиваем опции BadgerDB
	opts := badger.DefaultOptions(dataDir)
	opts.Logger = nil // Отключаем логи BadgerDB, используем slog

	// Открываем базу данных
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &BadgerStorage{
		db:   db,
		path: dataDir,
	}

	slog.Info("BadgerDB storage initialized", "path", dataDir)
	return storage, nil
}

// Close закрывает соединение с базой данных
func (s *BadgerStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}

// Clear очищает все данные из базы
func (s *BadgerStorage) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("database is not open")
	}

	// Закрываем текущую базу
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	// Удаляем директорию с данными
	if err := os.RemoveAll(s.path); err != nil {
		return fmt.Errorf("failed to remove data directory: %w", err)
	}

	// Создаем новую директорию
	if err := os.MkdirAll(s.path, 0755); err != nil {
		return fmt.Errorf("failed to recreate data directory: %w", err)
	}

	// Открываем новую базу
	opts := badger.DefaultOptions(s.path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return fmt.Errorf("failed to reopen database: %w", err)
	}

	s.db = db
	slog.Info("Database cleared and reinitialized")
	return nil
}

// SaveBlock сохраняет блок в базу данных
func (s *BadgerStorage) SaveBlock(block *blockchain.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("database is not open")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		// Сериализуем блок
		blockData, err := s.serializeBlock(block)
		if err != nil {
			return fmt.Errorf("failed to serialize block: %w", err)
		}

		// Сохраняем блок по хешу
		blockKey := s.getBlockKey(block.Hash)
		if err := txn.Set(blockKey, blockData); err != nil {
			return fmt.Errorf("failed to save block by hash: %w", err)
		}

		// Сохраняем блок по высоте
		heightKey := s.getHeightKey(block.Height)
		if err := txn.Set(heightKey, blockData); err != nil {
			return fmt.Errorf("failed to save block by height: %w", err)
		}

		// Обновляем метаданные
		if err := s.saveHeight(txn, block.Height); err != nil {
			return fmt.Errorf("failed to save height: %w", err)
		}

		slog.Debug("Block saved", "hash", fmt.Sprintf("%x", block.Hash), "height", block.Height)
		return nil
	})
}

// GetBlock получает блок по хешу
func (s *BadgerStorage) GetBlock(hash []byte) (*blockchain.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var block *blockchain.Block
	err := s.db.View(func(txn *badger.Txn) error {
		blockKey := s.getBlockKey(hash)
		item, err := txn.Get(blockKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var err error
			block, err = s.deserializeBlock(val)
			return err
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("block not found")
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return block, nil
}

// GetBlockByHeight получает блок по высоте
func (s *BadgerStorage) GetBlockByHeight(height int64) (*blockchain.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var block *blockchain.Block
	err := s.db.View(func(txn *badger.Txn) error {
		heightKey := s.getHeightKey(height)
		item, err := txn.Get(heightKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var err error
			block, err = s.deserializeBlock(val)
			return err
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, fmt.Errorf("block at height %d not found", height)
		}
		return nil, fmt.Errorf("failed to get block by height: %w", err)
	}

	return block, nil
}

// GetAllBlocks получает все блоки
func (s *BadgerStorage) GetAllBlocks() ([]*blockchain.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var blocks []*blockchain.Block
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("block:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				block, err := s.deserializeBlock(val)
				if err != nil {
					return err
				}
				blocks = append(blocks, block)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all blocks: %w", err)
	}

	return blocks, nil
}

// GetLatestBlock получает последний блок
func (s *BadgerStorage) GetLatestBlock() (*blockchain.Block, error) {
	height, err := s.GetHeight()
	if err != nil {
		return nil, fmt.Errorf("failed to get height: %w", err)
	}

	if height < 0 {
		return nil, fmt.Errorf("no blocks found")
	}

	return s.GetBlockByHeight(height)
}

// SaveUTXOSet сохраняет UTXO набор
func (s *BadgerStorage) SaveUTXOSet(utxoSet *blockchain.UTXOSet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("database is not open")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		// Сериализуем UTXO набор
		utxoData, err := s.serializeUTXOSet(utxoSet)
		if err != nil {
			return fmt.Errorf("failed to serialize UTXO set: %w", err)
		}

		// Сохраняем UTXO набор
		utxoKey := []byte("utxo:set")
		if err := txn.Set(utxoKey, utxoData); err != nil {
			return fmt.Errorf("failed to save UTXO set: %w", err)
		}

		slog.Debug("UTXO set saved")
		return nil
	})
}

// GetUTXOSet получает UTXO набор
func (s *BadgerStorage) GetUTXOSet() (*blockchain.UTXOSet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return nil, fmt.Errorf("database is not open")
	}

	var utxoSet *blockchain.UTXOSet
	err := s.db.View(func(txn *badger.Txn) error {
		utxoKey := []byte("utxo:set")
		item, err := txn.Get(utxoKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var err error
			utxoSet, err = s.deserializeUTXOSet(val)
			return err
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Возвращаем пустой UTXO набор, если не найден
			return blockchain.NewUTXOSet(), nil
		}
		return nil, fmt.Errorf("failed to get UTXO set: %w", err)
	}

	return utxoSet, nil
}

// SaveDifficulty сохраняет сложность
func (s *BadgerStorage) SaveDifficulty(difficulty int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("database is not open")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		difficultyKey := []byte("meta:difficulty")
		difficultyData := []byte(fmt.Sprintf("%d", difficulty))
		return txn.Set(difficultyKey, difficultyData)
	})
}

// GetDifficulty получает сложность
func (s *BadgerStorage) GetDifficulty() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return 0, fmt.Errorf("database is not open")
	}

	var difficulty int
	err := s.db.View(func(txn *badger.Txn) error {
		difficultyKey := []byte("meta:difficulty")
		item, err := txn.Get(difficultyKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			_, err := fmt.Sscanf(string(val), "%d", &difficulty)
			return err
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return 0, nil // Возвращаем 0 по умолчанию
		}
		return 0, fmt.Errorf("failed to get difficulty: %w", err)
	}

	return difficulty, nil
}

// SaveHeight сохраняет высоту блокчейна
func (s *BadgerStorage) SaveHeight(height int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db == nil {
		return fmt.Errorf("database is not open")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return s.saveHeight(txn, height)
	})
}

// saveHeight внутренний метод для сохранения высоты
func (s *BadgerStorage) saveHeight(txn *badger.Txn, height int64) error {
	heightKey := []byte("meta:height")
	heightData := []byte(fmt.Sprintf("%d", height))
	return txn.Set(heightKey, heightData)
}

// GetHeight получает высоту блокчейна
func (s *BadgerStorage) GetHeight() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.db == nil {
		return 0, fmt.Errorf("database is not open")
	}

	var height int64
	err := s.db.View(func(txn *badger.Txn) error {
		heightKey := []byte("meta:height")
		item, err := txn.Get(heightKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			_, err := fmt.Sscanf(string(val), "%d", &height)
			return err
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return -1, nil // Возвращаем -1, если блоков нет
		}
		return 0, fmt.Errorf("failed to get height: %w", err)
	}

	return height, nil
}

// Вспомогательные методы для работы с ключами

func (s *BadgerStorage) getBlockKey(hash []byte) []byte {
	return []byte(fmt.Sprintf("block:%x", hash))
}

func (s *BadgerStorage) getHeightKey(height int64) []byte {
	return []byte(fmt.Sprintf("height:%d", height))
}

// Сериализация и десериализация

func (s *BadgerStorage) serializeBlock(block *blockchain.Block) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(block); err != nil {
		return nil, fmt.Errorf("failed to encode block: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *BadgerStorage) deserializeBlock(data []byte) (*blockchain.Block, error) {
	var block blockchain.Block
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&block); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}
	return &block, nil
}

func (s *BadgerStorage) serializeUTXOSet(utxoSet *blockchain.UTXOSet) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(utxoSet); err != nil {
		return nil, fmt.Errorf("failed to encode UTXO set: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *BadgerStorage) deserializeUTXOSet(data []byte) (*blockchain.UTXOSet, error) {
	var utxoSet blockchain.UTXOSet
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&utxoSet); err != nil {
		return nil, fmt.Errorf("failed to decode UTXO set: %w", err)
	}
	return &utxoSet, nil
}
