package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"sync"
)

// Blockchain представляет блокчейн
type Blockchain struct {
	Chain      []*Block `json:"chain"`
	UTXOSet    *UTXOSet `json:"utxo_set"`
	Difficulty int      `json:"difficulty"`
	mutex      sync.RWMutex
}

// NewBlockchain создает новый блокчейн с genesis блоком
func NewBlockchain(address string, publicKey []byte, difficulty int) *Blockchain {
	bc := &Blockchain{
		UTXOSet:    NewUTXOSet(),
		Difficulty: difficulty,
	}

	// Создаем genesis блок
	genesisBlock := bc.createGenesisBlock(address, publicKey)
	bc.Chain = append(bc.Chain, genesisBlock)

	// Обрабатываем genesis блок в UTXOSet
	if err := bc.UTXOSet.ProcessBlock(genesisBlock); err != nil {
		slog.Error("Failed to process genesis block", "error", err)
	}

	return bc
}

// createGenesisBlock создает genesis блок
func (bc *Blockchain) createGenesisBlock(address string, publicKey []byte) *Block {
	// Создаем coinbase транзакцию для genesis блока
	coinbaseTx := NewCoinbaseTransaction(address, publicKey, 1000000) // 1M монет в genesis

	return NewBlock([]*Transaction{coinbaseTx}, []byte{}, 0, bc.Difficulty)
}

// AddBlock добавляет блок в блокчейн
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Получаем последний блок напрямую, без вызова GetLastBlock()
	var lastBlock *Block
	if len(bc.Chain) > 0 {
		lastBlock = bc.Chain[len(bc.Chain)-1]
	}

	// Проверяем валидность блока
	if lastBlock != nil && !block.IsValid(lastBlock) {
		return fmt.Errorf("invalid block")
	}

	// Проверяем, что блок не дублируется
	for _, existingBlock := range bc.Chain {
		if bytes.Equal(existingBlock.Hash, block.Hash) {
			return fmt.Errorf("block already exists")
		}
	}

	// Обрабатываем блок в UTXOSet
	if err := bc.UTXOSet.ProcessBlock(block); err != nil {
		return fmt.Errorf("failed to process block in UTXOSet: %w", err)
	}

	// Добавляем блок в цепочку
	bc.Chain = append(bc.Chain, block)

	slog.Info("Block added to blockchain", "height", block.Height, "hash", block.Hash)
	return nil
}

// GetLastBlock возвращает последний блок в цепочке
func (bc *Blockchain) GetLastBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Chain) == 0 {
		return nil
	}
	return bc.Chain[len(bc.Chain)-1]
}

// GetLastBlockAndDifficulty возвращает последний блок и сложность без дополнительных блокировок
func (bc *Blockchain) GetLastBlockAndDifficulty() (*Block, int) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Chain) == 0 {
		return nil, bc.Difficulty
	}
	return bc.Chain[len(bc.Chain)-1], bc.Difficulty
}

// GetBlock возвращает блок по хешу
func (bc *Blockchain) GetBlock(hash []byte) *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for _, block := range bc.Chain {
		if bytes.Equal(block.Hash, hash) {
			return block
		}
	}
	return nil
}

// GetBlockByHeight возвращает блок по высоте
func (bc *Blockchain) GetBlockByHeight(height int64) *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if height < 0 || int(height) >= len(bc.Chain) {
		return nil
	}
	return bc.Chain[height]
}

// GetHeight возвращает текущую высоту блокчейна
func (bc *Blockchain) GetHeight() int64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return int64(len(bc.Chain) - 1)
}

// GetGenesisHash возвращает хеш genesis блока
func (bc *Blockchain) GetGenesisHash() []byte {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Chain) == 0 {
		return nil
	}
	return bc.Chain[0].Hash
}

// IsValid проверяет валидность всего блокчейна
func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Chain) == 0 {
		return false
	}

	// Проверяем каждый блок
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		if !currentBlock.IsValid(previousBlock) {
			slog.Error("Invalid block in chain", "height", currentBlock.Height)
			return false
		}
	}

	return true
}

// GetBalance возвращает баланс для указанного адреса
func (bc *Blockchain) GetBalance(address string) int64 {
	return bc.UTXOSet.GetBalance(address)
}

// GetUTXOsByAddress возвращает все UTXO для указанного адреса
func (bc *Blockchain) GetUTXOsByAddress(address string) []*UTXO {
	return bc.UTXOSet.GetUTXOsByAddress(address)
}

// CreateTransaction создает новую транзакцию
func (bc *Blockchain) CreateTransaction(from, to string, amount int64, privateKey []byte) (*Transaction, error) {
	// Получаем UTXO для отправителя
	utxos := bc.GetUTXOsByAddress(from)
	if len(utxos) == 0 {
		return nil, fmt.Errorf("no UTXOs found for address %s", from)
	}

	// Выбираем UTXO для транзакции (простой алгоритм)
	var selectedUTXOs []*UTXO
	var totalValue int64

	for _, utxo := range utxos {
		selectedUTXOs = append(selectedUTXOs, utxo)
		totalValue += utxo.Value
		if totalValue >= amount {
			break
		}
	}

	if totalValue < amount {
		return nil, fmt.Errorf("insufficient balance: %d < %d", totalValue, amount)
	}

	// Создаем входы транзакции
	var inputs []*TransactionInput
	for _, utxo := range selectedUTXOs {
		input := &TransactionInput{
			TransactionID: utxo.TransactionID,
			OutputIndex:   utxo.OutputIndex,
			PublicKey:     utxo.PublicKey,
			Signature:     []byte("dummy_signature"), // Заглушка для тестирования
		}
		inputs = append(inputs, input)
	}

	// Создаем выходы транзакции
	var outputs []*TransactionOutput
	outputs = append(outputs, &TransactionOutput{
		Value:     amount,
		Address:   to,
		PublicKey: []byte{}, // В реальной реализации нужно получить публичный ключ получателя
	})

	// Добавляем сдачу, если нужно
	change := totalValue - amount
	if change > 0 {
		outputs = append(outputs, &TransactionOutput{
			Value:     change,
			Address:   from,
			PublicKey: []byte{}, // В реальной реализации нужно получить публичный ключ отправителя
		})
	}

	// Создаем транзакцию
	tx := NewTransaction(inputs, outputs)

	// TODO: Подписать транзакцию приватным ключом
	// В реальной реализации здесь нужно подписать каждый вход

	return tx, nil
}

// Serialize сериализует блокчейн в байты
func (bc *Blockchain) Serialize() ([]byte, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(bc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize blockchain: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeBlockchain десериализует блокчейн из байтов
func DeserializeBlockchain(data []byte) (*Blockchain, error) {
	var bc Blockchain
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&bc)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize blockchain: %w", err)
	}
	return &bc, nil
}

// GetStats возвращает статистику блокчейна
func (bc *Blockchain) GetStats() map[string]interface{} {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	// Вычисляем статистику напрямую, без вызова других методов
	height := int64(len(bc.Chain) - 1)
	var lastBlock *Block
	if len(bc.Chain) > 0 {
		lastBlock = bc.Chain[len(bc.Chain)-1]
	}

	// Проверяем валидность блокчейна
	isValid := true
	if len(bc.Chain) > 0 {
		for i := 1; i < len(bc.Chain); i++ {
			currentBlock := bc.Chain[i]
			previousBlock := bc.Chain[i-1]
			if !currentBlock.IsValid(previousBlock) {
				isValid = false
				break
			}
		}
	}

	stats := map[string]interface{}{
		"height":      height,
		"block_count": len(bc.Chain),
		"difficulty":  bc.Difficulty,
		"is_valid":    isValid,
	}

	if lastBlock != nil {
		stats["last_block_hash"] = lastBlock.Hash
		stats["last_block_timestamp"] = lastBlock.Timestamp
		stats["last_block_transactions"] = len(lastBlock.Transactions)
	}

	// Добавляем статистику UTXOSet
	utxoStats := bc.UTXOSet.GetStats()
	for k, v := range utxoStats {
		stats["utxo_"+k] = v
	}

	return stats
}

// String возвращает строковое представление блокчейна
func (bc *Blockchain) String() string {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	height := int64(len(bc.Chain) - 1)

	// Проверяем валидность блокчейна
	isValid := true
	if len(bc.Chain) > 0 {
		for i := 1; i < len(bc.Chain); i++ {
			currentBlock := bc.Chain[i]
			previousBlock := bc.Chain[i-1]
			if !currentBlock.IsValid(previousBlock) {
				isValid = false
				break
			}
		}
	}

	return fmt.Sprintf("Blockchain{Height: %d, Blocks: %d, Difficulty: %d, Valid: %t}",
		height, len(bc.Chain), bc.Difficulty, isValid)
}
