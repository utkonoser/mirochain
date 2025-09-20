package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log/slog"
	"time"
)

// TransactionInput представляет вход транзакции (UTXO)
type TransactionInput struct {
	TransactionID []byte `json:"transaction_id"` // ID транзакции, содержащей UTXO
	OutputIndex   int    `json:"output_index"`   // Индекс выхода в той транзакции
	Signature     []byte `json:"signature"`      // Подпись владельца UTXO
	PublicKey     []byte `json:"public_key"`     // Публичный ключ владельца UTXO
}

// TransactionOutput представляет выход транзакции
type TransactionOutput struct {
	Value     int64  `json:"value"`      // Количество монет
	Address   string `json:"address"`    // Адрес получателя
	PublicKey []byte `json:"public_key"` // Публичный ключ получателя
}

// Transaction представляет транзакцию в блокчейне
type Transaction struct {
	ID        []byte               `json:"id"`
	Inputs    []*TransactionInput  `json:"inputs"`
	Outputs   []*TransactionOutput `json:"outputs"`
	Timestamp int64                `json:"timestamp"`
	Fee       int64                `json:"fee"`
}

// NewTransaction создает новую транзакцию
func NewTransaction(inputs []*TransactionInput, outputs []*TransactionOutput) *Transaction {
	tx := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now().Unix(),
	}

	// Вычисляем комиссию
	tx.calculateFee()

	// Вычисляем ID транзакции
	tx.ID = tx.calculateHash()

	return tx
}

// NewCoinbaseTransaction создает coinbase транзакцию (для майнинга)
func NewCoinbaseTransaction(address string, publicKey []byte, reward int64) *Transaction {
	// Coinbase транзакция не имеет входов
	output := &TransactionOutput{
		Value:     reward,
		Address:   address,
		PublicKey: publicKey,
	}

	tx := &Transaction{
		Inputs:    []*TransactionInput{},
		Outputs:   []*TransactionOutput{output},
		Timestamp: time.Now().Unix(),
		Fee:       0, // Coinbase транзакция не имеет комиссии
	}

	tx.ID = tx.calculateHash()
	return tx
}

// calculateHash вычисляет хеш транзакции
func (tx *Transaction) calculateHash() []byte {
	var inputs []byte
	for _, input := range tx.Inputs {
		inputs = append(inputs, input.TransactionID...)
		inputs = append(inputs, []byte(fmt.Sprintf("%d", input.OutputIndex))...)
	}

	var outputs []byte
	for _, output := range tx.Outputs {
		outputs = append(outputs, []byte(output.Address)...)
		outputs = append(outputs, []byte(fmt.Sprintf("%d", output.Value))...)
	}

	data := bytes.Join([][]byte{
		inputs,
		outputs,
		[]byte(fmt.Sprintf("%d", tx.Timestamp)),
	}, []byte{})

	hash := sha256.Sum256(data)
	return hash[:]
}

// calculateFee вычисляет комиссию транзакции
func (tx *Transaction) calculateFee() {
	// Для coinbase транзакции комиссия = 0
	if len(tx.Inputs) == 0 {
		tx.Fee = 0
		return
	}

	// В реальной реализации здесь нужно получить значение UTXO
	// Пока что используем заглушку - предполагаем, что входы покрывают выходы
	tx.Fee = 0
}

// IsValid проверяет валидность транзакции
func (tx *Transaction) IsValid() bool {
	// Проверяем хеш транзакции
	calculatedHash := tx.calculateHash()
	if !bytes.Equal(tx.ID, calculatedHash) {
		slog.Error("Invalid transaction hash", "expected", calculatedHash, "actual", tx.ID)
		return false
	}

	// Проверяем, что есть хотя бы один выход
	if len(tx.Outputs) == 0 {
		slog.Error("Transaction has no outputs")
		return false
	}

	// Проверяем, что сумма выходов не превышает сумму входов (кроме coinbase)
	// В реальной реализации здесь нужно проверить баланс UTXO
	// Пока что пропускаем эту проверку

	// Проверяем подписи входов (кроме coinbase)
	if len(tx.Inputs) > 0 {
		for i, input := range tx.Inputs {
			if len(input.Signature) == 0 {
				slog.Error("Transaction input has no signature", "input_index", i)
				return false
			}
			if len(input.PublicKey) == 0 {
				slog.Error("Transaction input has no public key", "input_index", i)
				return false
			}
			_ = input // Используем переменную, чтобы избежать ошибки компилятора
		}
	}

	return true
}

// IsCoinbase проверяет, является ли транзакция coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 0
}

// Serialize сериализует транзакцию в байты
func (tx *Transaction) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeTransaction десериализует транзакцию из байтов
func DeserializeTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}
	return &tx, nil
}

// String возвращает строковое представление транзакции
func (tx *Transaction) String() string {
	return fmt.Sprintf("Transaction{ID: %x, Inputs: %d, Outputs: %d, Fee: %d}",
		tx.ID, len(tx.Inputs), len(tx.Outputs), tx.Fee)
}
