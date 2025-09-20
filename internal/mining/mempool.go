package mining

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/parallel"
)

// Mempool представляет пул неподтвержденных транзакций
type Mempool struct {
	Transactions map[string]*blockchain.Transaction `json:"transactions"` // Ключ: transaction ID
	mutex        sync.RWMutex
	MaxSize      int `json:"max_size"`
	// TransactionProcessor для параллельной валидации
	processor *parallel.TransactionProcessor
}

// NewMempool создает новый mempool
func NewMempool(maxSize int) *Mempool {
	return &Mempool{
		Transactions: make(map[string]*blockchain.Transaction),
		MaxSize:      maxSize,
		processor:    parallel.NewTransactionProcessor(2, 50), // 2 воркера для mempool
	}
}

// AddTransaction добавляет транзакцию в mempool
func (mp *Mempool) AddTransaction(tx *blockchain.Transaction) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Проверяем, что mempool не переполнен
	if len(mp.Transactions) >= mp.MaxSize {
		return fmt.Errorf("mempool is full")
	}

	// Проверяем валидность транзакции
	if !tx.IsValid() {
		return fmt.Errorf("invalid transaction")
	}

	// Проверяем, что транзакция не дублируется
	txID := string(tx.ID)
	if _, exists := mp.Transactions[txID]; exists {
		return fmt.Errorf("transaction already exists in mempool")
	}

	// Добавляем транзакцию
	mp.Transactions[txID] = tx

	slog.Debug("Transaction added to mempool", "tx_id", txID)
	return nil
}

// RemoveTransaction удаляет транзакцию из mempool
func (mp *Mempool) RemoveTransaction(txID string) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	delete(mp.Transactions, txID)
	slog.Debug("Transaction removed from mempool", "tx_id", txID)
}

// GetTransaction возвращает транзакцию по ID
func (mp *Mempool) GetTransaction(txID string) (*blockchain.Transaction, bool) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	tx, exists := mp.Transactions[txID]
	return tx, exists
}

// GetTransactions возвращает все транзакции
func (mp *Mempool) GetTransactions() []*blockchain.Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	var transactions []*blockchain.Transaction
	for _, tx := range mp.Transactions {
		transactions = append(transactions, tx)
	}
	return transactions
}

// GetTransactionsForBlock возвращает транзакции для создания блока
func (mp *Mempool) GetTransactionsForBlock(maxCount int) []*blockchain.Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	var transactions []*blockchain.Transaction
	count := 0

	// Сортируем транзакции по комиссии (в реальной реализации)
	for _, tx := range mp.Transactions {
		if count >= maxCount {
			break
		}
		transactions = append(transactions, tx)
		count++
	}

	return transactions
}

// RemoveTransactions удаляет транзакции из mempool
func (mp *Mempool) RemoveTransactions(txIDs []string) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	for _, txID := range txIDs {
		delete(mp.Transactions, txID)
	}
}

// Size возвращает количество транзакций в mempool
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return len(mp.Transactions)
}

// IsEmpty проверяет, пуст ли mempool
func (mp *Mempool) IsEmpty() bool {
	return mp.Size() == 0
}

// Clear очищает mempool
func (mp *Mempool) Clear() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	mp.Transactions = make(map[string]*blockchain.Transaction)
	slog.Info("Mempool cleared")
}

// GetStats возвращает статистику mempool
func (mp *Mempool) GetStats() map[string]interface{} {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	var totalFees int64
	for _, tx := range mp.Transactions {
		totalFees += tx.Fee
	}

	return map[string]interface{}{
		"size":       len(mp.Transactions),
		"max_size":   mp.MaxSize,
		"total_fees": totalFees,
		"is_empty":   len(mp.Transactions) == 0,
	}
}

// String возвращает строковое представление mempool
func (mp *Mempool) String() string {
	return fmt.Sprintf("Mempool{Size: %d/%d, Empty: %t}",
		mp.Size(), mp.MaxSize, mp.IsEmpty())
}

// StartProcessor запускает TransactionProcessor для mempool
func (mp *Mempool) StartProcessor() error {
	utxoSet := blockchain.NewUTXOSet()
	return mp.processor.Start(utxoSet)
}

// StopProcessor останавливает TransactionProcessor для mempool
func (mp *Mempool) StopProcessor() {
	mp.processor.Stop()
}

// ValidateTransactionAsync валидирует транзакцию асинхронно
func (mp *Mempool) ValidateTransactionAsync(tx *blockchain.Transaction, utxoSet *blockchain.UTXOSet) <-chan *parallel.TransactionResult {
	resultChan := make(chan *parallel.TransactionResult, 1)

	go func() {
		result, err := mp.processor.ProcessTransaction(tx, utxoSet, 1)
		if err != nil {
			result = &parallel.TransactionResult{
				Transaction: tx,
				Valid:       false,
				Error:       err,
				ProcessTime: 0,
			}
		}
		resultChan <- result
	}()

	return resultChan
}

// ValidateTransactionsBatch валидирует несколько транзакций параллельно
func (mp *Mempool) ValidateTransactionsBatch(transactions []*blockchain.Transaction, utxoSet *blockchain.UTXOSet) []*parallel.TransactionResult {
	results := make([]*parallel.TransactionResult, len(transactions))

	// Создаем каналы для результатов
	channels := make([]<-chan *parallel.TransactionResult, len(transactions))
	for i, tx := range transactions {
		channels[i] = mp.ValidateTransactionAsync(tx, utxoSet)
	}

	// Собираем результаты
	for i, ch := range channels {
		select {
		case result := <-ch:
			results[i] = result
		case <-time.After(10 * time.Second):
			results[i] = &parallel.TransactionResult{
				Transaction: transactions[i],
				Valid:       false,
				Error:       fmt.Errorf("transaction validation timeout"),
				ProcessTime: 10 * time.Second,
			}
		}
	}

	return results
}

// GetProcessorStats возвращает статистику TransactionProcessor
func (mp *Mempool) GetProcessorStats() map[string]interface{} {
	return mp.processor.GetStats()
}
