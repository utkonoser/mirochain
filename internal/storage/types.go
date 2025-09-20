package storage

import "mirochain/internal/blockchain"

// Storage представляет интерфейс для персистентного хранения
type Storage interface {
	// Блоки
	SaveBlock(block *blockchain.Block) error
	GetBlock(hash []byte) (*blockchain.Block, error)
	GetBlockByHeight(height int64) (*blockchain.Block, error)
	GetAllBlocks() ([]*blockchain.Block, error)
	GetLatestBlock() (*blockchain.Block, error)

	// UTXO
	SaveUTXOSet(utxoSet *blockchain.UTXOSet) error
	GetUTXOSet() (*blockchain.UTXOSet, error)

	// Метаданные
	SaveDifficulty(difficulty int) error
	GetDifficulty() (int, error)
	SaveHeight(height int64) error
	GetHeight() (int64, error)

	// Управление
	Close() error
	Clear() error
}
