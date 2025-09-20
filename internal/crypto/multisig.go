package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"
)

// MultiSig представляет систему мультиподписей
type MultiSig struct {
	Address      string   `json:"address"`
	PublicKeys   [][]byte `json:"public_keys"`
	Threshold    int      `json:"threshold"`
	Algorithm    SignatureAlgorithm `json:"algorithm"`
	CreatedAt    int64    `json:"created_at"`
	LastUsed     int64    `json:"last_used"`
}

// MultiSigTransaction представляет транзакцию с мультиподписью
type MultiSigTransaction struct {
	ID          string            `json:"id"`
	From        string            `json:"from"`
	To          string            `json:"to"`
	Amount      *big.Int          `json:"amount"`
	Data        []byte            `json:"data"`
	Signatures  map[string][]byte `json:"signatures"` // address -> signature
	Threshold   int               `json:"threshold"`
	Status      string            `json:"status"` // pending, signed, executed
	CreatedAt   int64             `json:"created_at"`
	ExecutedAt  int64             `json:"executed_at"`
}

// MultiSigManager управляет мультиподписями
type MultiSigManager struct {
	multisigs    map[string]*MultiSig
	transactions map[string]*MultiSigTransaction
	signatureManager *SignatureManager
	mutex        sync.RWMutex
}

// NewMultiSigManager создает новый менеджер мультиподписей
func NewMultiSigManager() *MultiSigManager {
	return &MultiSigManager{
		multisigs:        make(map[string]*MultiSig),
		transactions:      make(map[string]*MultiSigTransaction),
		signatureManager: NewSignatureManager(),
	}
}

// CreateMultiSig создает новую мультиподпись
func (msm *MultiSigManager) CreateMultiSig(publicKeys [][]byte, threshold int, algorithm SignatureAlgorithm) (*MultiSig, error) {
	if len(publicKeys) < 2 {
		return nil, fmt.Errorf("multisig requires at least 2 public keys")
	}
	
	if threshold < 2 {
		return nil, fmt.Errorf("threshold must be at least 2")
	}
	
	if threshold > len(publicKeys) {
		return nil, fmt.Errorf("threshold cannot exceed number of public keys")
	}
	
	// Сортируем публичные ключи для детерминированного адреса
	sortedKeys := make([][]byte, len(publicKeys))
	copy(sortedKeys, publicKeys)
	sort.Slice(sortedKeys, func(i, j int) bool {
		return string(sortedKeys[i]) < string(sortedKeys[j])
	})
	
	// Генерируем адрес мультиподписи
	address, err := msm.generateMultiSigAddress(sortedKeys, threshold, algorithm)
	if err != nil {
		return nil, err
	}
	
	multisig := &MultiSig{
		Address:    address,
		PublicKeys: sortedKeys,
		Threshold:  threshold,
		Algorithm:  algorithm,
		CreatedAt:  time.Now().Unix(),
		LastUsed:   0,
	}
	
	msm.mutex.Lock()
	msm.multisigs[address] = multisig
	msm.mutex.Unlock()
	
	return multisig, nil
}

// generateMultiSigAddress генерирует адрес для мультиподписи
func (msm *MultiSigManager) generateMultiSigAddress(publicKeys [][]byte, threshold int, algorithm SignatureAlgorithm) (string, error) {
	// Создаем детерминированный хеш из всех параметров
	data := fmt.Sprintf("%d:%s:", threshold, algorithm)
	for _, key := range publicKeys {
		data += hex.EncodeToString(key) + ":"
	}
	
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// CreateTransaction создает транзакцию с мультиподписью
func (msm *MultiSigManager) CreateTransaction(from, to string, amount *big.Int, data []byte) (*MultiSigTransaction, error) {
	msm.mutex.RLock()
	multisig, exists := msm.multisigs[from]
	msm.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("multisig address %s not found", from)
	}
	
	// Генерируем ID транзакции
	txID := msm.generateTransactionID(from, to, amount, data)
	
	transaction := &MultiSigTransaction{
		ID:         txID,
		From:       from,
		To:         to,
		Amount:     new(big.Int).Set(amount),
		Data:       data,
		Signatures: make(map[string][]byte),
		Threshold:  multisig.Threshold,
		Status:     "pending",
		CreatedAt:  time.Now().Unix(),
		ExecutedAt: 0,
	}
	
	msm.mutex.Lock()
	msm.transactions[txID] = transaction
	msm.mutex.Unlock()
	
	return transaction, nil
}

// generateTransactionID генерирует ID транзакции
func (msm *MultiSigManager) generateTransactionID(from, to string, amount *big.Int, data []byte) string {
	dataStr := fmt.Sprintf("%s:%s:%s:%s", from, to, amount.String(), hex.EncodeToString(data))
	hash := sha256.Sum256([]byte(dataStr))
	return hex.EncodeToString(hash[:])
}

// SignTransaction подписывает транзакцию
func (msm *MultiSigManager) SignTransaction(txID string, signerAddress string, privateKey []byte) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()
	
	transaction, exists := msm.transactions[txID]
	if !exists {
		return fmt.Errorf("transaction %s not found", txID)
	}
	
	if transaction.Status != "pending" {
		return fmt.Errorf("transaction %s is not pending", txID)
	}
	
	// Получаем мультиподпись
	multisig, exists := msm.multisigs[transaction.From]
	if !exists {
		return fmt.Errorf("multisig address %s not found", transaction.From)
	}
	
	// Проверяем, что подписывающий является участником мультиподписи
	if !msm.isParticipant(multisig, signerAddress) {
		return fmt.Errorf("address %s is not a participant in multisig %s", signerAddress, transaction.From)
	}
	
	// Проверяем, что подписывающий еще не подписал
	if _, alreadySigned := transaction.Signatures[signerAddress]; alreadySigned {
		return fmt.Errorf("address %s has already signed transaction %s", signerAddress, txID)
	}
	
	// Создаем данные для подписи
	signData := msm.createSignData(transaction)
	
	// Подписываем данные
	signature, err := msm.signatureManager.Sign(multisig.Algorithm, privateKey, signData)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}
	
	// Добавляем подпись
	transaction.Signatures[signerAddress] = signature.Data
	
	// Проверяем, достигнут ли порог подписей
	if len(transaction.Signatures) >= transaction.Threshold {
		transaction.Status = "signed"
		transaction.ExecutedAt = time.Now().Unix()
		
		// Обновляем время последнего использования мультиподписи
		multisig.LastUsed = transaction.ExecutedAt
	}
	
	return nil
}

// isParticipant проверяет, является ли адрес участником мультиподписи
func (msm *MultiSigManager) isParticipant(multisig *MultiSig, address string) bool {
	for _, publicKey := range multisig.PublicKeys {
		// В реальной реализации здесь должна быть проверка адреса
		// Пока что используем простую проверку
		if hex.EncodeToString(publicKey) == address {
			return true
		}
	}
	return false
}

// createSignData создает данные для подписи
func (msm *MultiSigManager) createSignData(transaction *MultiSigTransaction) []byte {
	data := fmt.Sprintf("%s:%s:%s:%d:%d", 
		transaction.ID, 
		transaction.From, 
		transaction.To, 
		transaction.Amount.Int64(),
		transaction.CreatedAt)
	return []byte(data)
}

// VerifyTransaction проверяет транзакцию с мультиподписью
func (msm *MultiSigManager) VerifyTransaction(txID string) (bool, error) {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	transaction, exists := msm.transactions[txID]
	if !exists {
		return false, fmt.Errorf("transaction %s not found", txID)
	}
	
	if transaction.Status != "signed" {
		return false, fmt.Errorf("transaction %s is not signed", txID)
	}
	
	// Получаем мультиподпись
	multisig, exists := msm.multisigs[transaction.From]
	if !exists {
		return false, fmt.Errorf("multisig address %s not found", transaction.From)
	}
	
	// Проверяем количество подписей
	if len(transaction.Signatures) < transaction.Threshold {
		return false, fmt.Errorf("insufficient signatures: %d/%d", len(transaction.Signatures), transaction.Threshold)
	}
	
	// Проверяем каждую подпись
	signData := msm.createSignData(transaction)
	validSignatures := 0
	
	for signerAddress, signatureData := range transaction.Signatures {
		// Находим публичный ключ подписывающего
		publicKey, err := msm.findPublicKey(multisig, signerAddress)
		if err != nil {
			continue
		}
		
		// Проверяем подпись
		signature := &Signature{
			Algorithm: multisig.Algorithm,
			Data:      signatureData,
			PublicKey: publicKey,
		}
		
		if msm.signatureManager.Verify(signature, signData) {
			validSignatures++
		}
	}
	
	return validSignatures >= transaction.Threshold, nil
}

// findPublicKey находит публичный ключ по адресу
func (msm *MultiSigManager) findPublicKey(multisig *MultiSig, address string) ([]byte, error) {
	for _, publicKey := range multisig.PublicKeys {
		// В реальной реализации здесь должна быть проверка адреса
		// Пока что используем простую проверку
		if hex.EncodeToString(publicKey) == address {
			return publicKey, nil
		}
	}
	return nil, fmt.Errorf("public key not found for address %s", address)
}

// GetMultiSig возвращает мультиподпись по адресу
func (msm *MultiSigManager) GetMultiSig(address string) (*MultiSig, error) {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	multisig, exists := msm.multisigs[address]
	if !exists {
		return nil, fmt.Errorf("multisig address %s not found", address)
	}
	
	return multisig, nil
}

// GetTransaction возвращает транзакцию по ID
func (msm *MultiSigManager) GetTransaction(txID string) (*MultiSigTransaction, error) {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	transaction, exists := msm.transactions[txID]
	if !exists {
		return nil, fmt.Errorf("transaction %s not found", txID)
	}
	
	return transaction, nil
}

// GetMultiSigTransactions возвращает транзакции мультиподписи
func (msm *MultiSigManager) GetMultiSigTransactions(address string) ([]*MultiSigTransaction, error) {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	var transactions []*MultiSigTransaction
	for _, transaction := range msm.transactions {
		if transaction.From == address {
			transactions = append(transactions, transaction)
		}
	}
	
	return transactions, nil
}

// GetStats возвращает статистику мультиподписей
func (msm *MultiSigManager) GetStats() map[string]interface{} {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_multisigs":    len(msm.multisigs),
		"total_transactions": len(msm.transactions),
		"pending_transactions": 0,
		"signed_transactions": 0,
		"executed_transactions": 0,
	}
	
	// Подсчитываем транзакции по статусам
	for _, transaction := range msm.transactions {
		switch transaction.Status {
		case "pending":
			stats["pending_transactions"] = stats["pending_transactions"].(int) + 1
		case "signed":
			stats["signed_transactions"] = stats["signed_transactions"].(int) + 1
		case "executed":
			stats["executed_transactions"] = stats["executed_transactions"].(int) + 1
		}
	}
	
	return stats
}

// ListMultiSigs возвращает список всех мультиподписей
func (msm *MultiSigManager) ListMultiSigs() []*MultiSig {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	multisigs := make([]*MultiSig, 0, len(msm.multisigs))
	for _, multisig := range msm.multisigs {
		multisigs = append(multisigs, multisig)
	}
	
	return multisigs
}

// ListTransactions возвращает список всех транзакций
func (msm *MultiSigManager) ListTransactions() []*MultiSigTransaction {
	msm.mutex.RLock()
	defer msm.mutex.RUnlock()
	
	transactions := make([]*MultiSigTransaction, 0, len(msm.transactions))
	for _, transaction := range msm.transactions {
		transactions = append(transactions, transaction)
	}
	
	return transactions
}

// DeleteMultiSig удаляет мультиподпись
func (msm *MultiSigManager) DeleteMultiSig(address string) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()
	
	_, exists := msm.multisigs[address]
	if !exists {
		return fmt.Errorf("multisig address %s not found", address)
	}
	
	// Проверяем, есть ли активные транзакции
	for _, transaction := range msm.transactions {
		if transaction.From == address && transaction.Status == "pending" {
			return fmt.Errorf("cannot delete multisig with pending transactions")
		}
	}
	
	delete(msm.multisigs, address)
	return nil
}

// UpdateThreshold обновляет порог подписей
func (msm *MultiSigManager) UpdateThreshold(address string, newThreshold int) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()
	
	multisig, exists := msm.multisigs[address]
	if !exists {
		return fmt.Errorf("multisig address %s not found", address)
	}
	
	if newThreshold < 2 {
		return fmt.Errorf("threshold must be at least 2")
	}
	
	if newThreshold > len(multisig.PublicKeys) {
		return fmt.Errorf("threshold cannot exceed number of public keys")
	}
	
	multisig.Threshold = newThreshold
	return nil
}

// AddParticipant добавляет участника в мультиподпись
func (msm *MultiSigManager) AddParticipant(address string, publicKey []byte) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()
	
	multisig, exists := msm.multisigs[address]
	if !exists {
		return fmt.Errorf("multisig address %s not found", address)
	}
	
	// Проверяем, что участник еще не добавлен
	for _, existingKey := range multisig.PublicKeys {
		if string(existingKey) == string(publicKey) {
			return fmt.Errorf("participant already exists")
		}
	}
	
	// Добавляем участника
	multisig.PublicKeys = append(multisig.PublicKeys, publicKey)
	
	// Сортируем ключи для детерминированности
	sort.Slice(multisig.PublicKeys, func(i, j int) bool {
		return string(multisig.PublicKeys[i]) < string(multisig.PublicKeys[j])
	})
	
	return nil
}

// RemoveParticipant удаляет участника из мультиподписи
func (msm *MultiSigManager) RemoveParticipant(address string, publicKey []byte) error {
	msm.mutex.Lock()
	defer msm.mutex.Unlock()
	
	multisig, exists := msm.multisigs[address]
	if !exists {
		return fmt.Errorf("multisig address %s not found", address)
	}
	
	// Находим и удаляем участника
	for i, existingKey := range multisig.PublicKeys {
		if string(existingKey) == string(publicKey) {
			multisig.PublicKeys = append(multisig.PublicKeys[:i], multisig.PublicKeys[i+1:]...)
			
			// Проверяем, что порог не превышает количество участников
			if multisig.Threshold > len(multisig.PublicKeys) {
				multisig.Threshold = len(multisig.PublicKeys)
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("participant not found")
}
