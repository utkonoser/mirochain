package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log/slog"
	"time"
)

// Block представляет блок в блокчейне
type Block struct {
	Timestamp    int64          `json:"timestamp"`
	PreviousHash []byte         `json:"previous_hash"`
	Hash         []byte         `json:"hash"`
	Nonce        int            `json:"nonce"`
	MerkleRoot   []byte         `json:"merkle_root"`
	Transactions []*Transaction `json:"transactions"`
	Height       int64          `json:"height"`
	Difficulty   int            `json:"difficulty"`
}

// NewBlock создает новый блок
func NewBlock(transactions []*Transaction, previousHash []byte, height int64, difficulty int) *Block {
	block := &Block{
		Timestamp:    time.Now().Unix(),
		PreviousHash: previousHash,
		Transactions: transactions,
		Height:       height,
		Difficulty:   difficulty,
	}

	// Вычисляем Merkle root
	block.MerkleRoot = block.calculateMerkleRoot()

	// Вычисляем хеш блока
	block.Hash = block.calculateHash()

	return block
}

// calculateHash вычисляет хеш блока
func (b *Block) calculateHash() []byte {
	// Создаем данные для хеширования (включая nonce)
	data := bytes.Join([][]byte{
		b.PreviousHash,
		b.MerkleRoot,
		[]byte(fmt.Sprintf("%d", b.Timestamp)),
		[]byte(fmt.Sprintf("%d", b.Difficulty)),
		[]byte(fmt.Sprintf("%d", b.Nonce)),
	}, []byte{})

	hash := sha256.Sum256(data)
	return hash[:]
}

// calculateMerkleRoot вычисляет Merkle root для транзакций
func (b *Block) calculateMerkleRoot() []byte {
	if len(b.Transactions) == 0 {
		return []byte{}
	}

	if len(b.Transactions) == 1 {
		return b.Transactions[0].ID
	}

	// Создаем листья Merkle tree
	leaves := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		leaves[i] = tx.ID
	}

	// Вычисляем Merkle root
	return b.merkleTree(leaves)
}

// merkleTree рекурсивно вычисляет Merkle tree
func (b *Block) merkleTree(hashes [][]byte) []byte {
	if len(hashes) == 1 {
		return hashes[0]
	}

	var nextLevel [][]byte
	for i := 0; i < len(hashes); i += 2 {
		var hash []byte
		if i+1 < len(hashes) {
			// Объединяем два хеша
			combined := bytes.Join([][]byte{hashes[i], hashes[i+1]}, []byte{})
			hashBytes := sha256.Sum256(combined)
			hash = hashBytes[:]
		} else {
			// Нечетное количество - дублируем последний хеш
			combined := bytes.Join([][]byte{hashes[i], hashes[i]}, []byte{})
			hashBytes := sha256.Sum256(combined)
			hash = hashBytes[:]
		}
		nextLevel = append(nextLevel, hash)
	}

	return b.merkleTree(nextLevel)
}

// IsValid проверяет валидность блока
func (b *Block) IsValid(previousBlock *Block) bool {
	// Проверяем хеш блока
	calculatedHash := b.calculateHash()
	if !bytes.Equal(b.Hash, calculatedHash) {
		slog.Error("Invalid block hash", "expected", calculatedHash, "actual", b.Hash)
		return false
	}

	// Проверяем связь с предыдущим блоком
	if previousBlock != nil {
		if !bytes.Equal(b.PreviousHash, previousBlock.Hash) {
			slog.Error("Invalid previous hash", "expected", previousBlock.Hash, "actual", b.PreviousHash)
			return false
		}

		if b.Height != previousBlock.Height+1 {
			slog.Error("Invalid block height", "expected", previousBlock.Height+1, "actual", b.Height)
			return false
		}
	}

	// Проверяем Merkle root
	calculatedMerkleRoot := b.calculateMerkleRoot()
	if !bytes.Equal(b.MerkleRoot, calculatedMerkleRoot) {
		slog.Error("Invalid merkle root", "expected", calculatedMerkleRoot, "actual", b.MerkleRoot)
		return false
	}

	// Проверяем транзакции
	for _, tx := range b.Transactions {
		if !tx.IsValid() {
			slog.Error("Invalid transaction in block", "tx_id", tx.ID)
			return false
		}
	}

	return true
}

// Serialize сериализует блок в байты
func (b *Block) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(b)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize block: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeBlock десериализует блок из байтов
func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize block: %w", err)
	}
	return &block, nil
}

// String возвращает строковое представление блока
func (b *Block) String() string {
	return fmt.Sprintf("Block{Height: %d, Hash: %x, PreviousHash: %x, Transactions: %d, Timestamp: %d}",
		b.Height, b.Hash, b.PreviousHash, len(b.Transactions), b.Timestamp)
}
