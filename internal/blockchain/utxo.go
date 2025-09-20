package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
)

// UTXO представляет неиспользованный выход транзакции
type UTXO struct {
	TransactionID []byte `json:"transaction_id"` // ID транзакции
	OutputIndex   int    `json:"output_index"`   // Индекс выхода в транзакции
	Value         int64  `json:"value"`          // Количество монет
	Address       string `json:"address"`        // Адрес владельца
	PublicKey     []byte `json:"public_key"`     // Публичный ключ владельца
	Spent         bool   `json:"spent"`          // Потрачен ли UTXO
}

// UTXOSet управляет множеством UTXO
type UTXOSet struct {
	UTXOs map[string]*UTXO `json:"utxos"` // Ключ: transactionID:outputIndex
	mutex sync.RWMutex
}

// NewUTXOSet создает новый UTXOSet
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		UTXOs: make(map[string]*UTXO),
	}
}

// AddUTXO добавляет UTXO в множество
func (us *UTXOSet) AddUTXO(utxo *UTXO) {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	key := us.getKey(utxo.TransactionID, utxo.OutputIndex)
	us.UTXOs[key] = utxo
}

// GetUTXO возвращает UTXO по ключу
func (us *UTXOSet) GetUTXO(transactionID []byte, outputIndex int) (*UTXO, bool) {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	key := us.getKey(transactionID, outputIndex)
	utxo, exists := us.UTXOs[key]
	return utxo, exists
}

// RemoveUTXO удаляет UTXO из множества
func (us *UTXOSet) RemoveUTXO(transactionID []byte, outputIndex int) {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	key := us.getKey(transactionID, outputIndex)
	delete(us.UTXOs, key)
}

// MarkAsSpent помечает UTXO как потраченный
func (us *UTXOSet) MarkAsSpent(transactionID []byte, outputIndex int) {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	key := us.getKey(transactionID, outputIndex)
	if utxo, exists := us.UTXOs[key]; exists {
		utxo.Spent = true
	}
}

// GetUTXOsByAddress возвращает все UTXO для указанного адреса
func (us *UTXOSet) GetUTXOsByAddress(address string) []*UTXO {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	var utxos []*UTXO
	for _, utxo := range us.UTXOs {
		if utxo.Address == address {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}

// GetBalance возвращает баланс для указанного адреса
func (us *UTXOSet) GetBalance(address string) int64 {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	var balance int64
	for _, utxo := range us.UTXOs {
		if utxo.Address == address {
			balance += utxo.Value
		}
	}
	return balance
}

// ProcessTransaction обрабатывает транзакцию и обновляет UTXO
func (us *UTXOSet) ProcessTransaction(tx *Transaction) error {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	// Удаляем потраченные UTXO (входы транзакции)
	for _, input := range tx.Inputs {
		key := us.getKey(input.TransactionID, input.OutputIndex)
		if utxo, exists := us.UTXOs[key]; exists {
			if utxo.Spent {
				return fmt.Errorf("UTXO already spent: %s", key)
			}
			// Удаляем UTXO из множества, а не помечаем как потраченный
			delete(us.UTXOs, key)
		} else {
			return fmt.Errorf("UTXO not found: %s", key)
		}
	}

	// Добавляем новые UTXO (выходы транзакции)
	for i, output := range tx.Outputs {
		utxo := &UTXO{
			TransactionID: tx.ID,
			OutputIndex:   i,
			Value:         output.Value,
			Address:       output.Address,
			PublicKey:     output.PublicKey,
			Spent:         false,
		}
		key := us.getKey(tx.ID, i)
		us.UTXOs[key] = utxo
	}

	return nil
}

// ProcessBlock обрабатывает блок и обновляет UTXO
func (us *UTXOSet) ProcessBlock(block *Block) error {
	for _, tx := range block.Transactions {
		if err := us.ProcessTransaction(tx); err != nil {
			return fmt.Errorf("failed to process transaction %x: %w", tx.ID, err)
		}
	}
	return nil
}

// getKey создает ключ для UTXO
func (us *UTXOSet) getKey(transactionID []byte, outputIndex int) string {
	return fmt.Sprintf("%x:%d", transactionID, outputIndex)
}

// Serialize сериализует UTXOSet в байты
func (us *UTXOSet) Serialize() ([]byte, error) {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(us.UTXOs)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize UTXOSet: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeUTXOSet десериализует UTXOSet из байтов
func DeserializeUTXOSet(data []byte) (*UTXOSet, error) {
	var utxos map[string]*UTXO
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&utxos)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize UTXOSet: %w", err)
	}

	return &UTXOSet{
		UTXOs: utxos,
	}, nil
}

// GetStats возвращает статистику UTXOSet
func (us *UTXOSet) GetStats() map[string]interface{} {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	total := len(us.UTXOs)
	var totalValue int64

	for _, utxo := range us.UTXOs {
		totalValue += utxo.Value
	}

	return map[string]interface{}{
		"total_utxos":   total,
		"unspent_utxos": total, // Все UTXO в множестве не потрачены
		"total_value":   totalValue,
	}
}

// String возвращает строковое представление UTXO
func (u *UTXO) String() string {
	return fmt.Sprintf("UTXO{TxID: %x, Index: %d, Value: %d, Address: %s, Spent: %t}",
		u.TransactionID, u.OutputIndex, u.Value, u.Address, u.Spent)
}
