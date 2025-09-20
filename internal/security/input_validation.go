package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"mirochain/internal/blockchain"
)

// InputValidator представляет валидатор входных данных
type InputValidator struct {
	maxBlockSize      int
	maxTransactionSize int
	maxAddressLength  int
	minTransactionFee int64
	allowedChars      *regexp.Regexp
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error реализует интерфейс error
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", ve.Field, ve.Message)
}

// ValidationResult представляет результат валидации
type ValidationResult struct {
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
}

// NewInputValidator создает новый валидатор входных данных
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxBlockSize:      1024 * 1024,  // 1MB
		maxTransactionSize: 64 * 1024,   // 64KB
		maxAddressLength:  64,           // максимальная длина адреса
		minTransactionFee: 1,            // минимальная комиссия
		allowedChars:      regexp.MustCompile(`^[a-zA-Z0-9+/=]+$`), // base64 символы
	}
}

// ValidateBlock валидирует блок
func (iv *InputValidator) ValidateBlock(block *blockchain.Block) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// Проверяем размер блока
	blockData, err := block.Serialize()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "block",
			Message: "Failed to serialize block",
			Code:    "SERIALIZATION_ERROR",
		})
		return result
	}
	
	if len(blockData) > iv.maxBlockSize {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "block_size",
			Message: fmt.Sprintf("Block size %d exceeds maximum %d", len(blockData), iv.maxBlockSize),
			Code:    "BLOCK_SIZE_EXCEEDED",
		})
	}
	
	// Проверяем заголовок блока
	iv.validateBlockHeader(block, result)
	
	// Проверяем транзакции
	transactions := make([]blockchain.Transaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		transactions[i] = *tx
	}
	iv.validateTransactions(transactions, result)
	
	// Проверяем временную метку
	iv.validateTimestamp(block.Timestamp, result)
	
	return result
}

// ValidateTransaction валидирует транзакцию
func (iv *InputValidator) ValidateTransaction(tx *blockchain.Transaction) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// Проверяем размер транзакции
	txData, err := tx.Serialize()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "transaction",
			Message: "Failed to serialize transaction",
			Code:    "SERIALIZATION_ERROR",
		})
		return result
	}
	
	if len(txData) > iv.maxTransactionSize {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "transaction_size",
			Message: fmt.Sprintf("Transaction size %d exceeds maximum %d", len(txData), iv.maxTransactionSize),
			Code:    "TRANSACTION_SIZE_EXCEEDED",
		})
	}
	
	// Проверяем ID транзакции
	iv.validateTransactionID(string(tx.ID), result)
	
	// Проверяем входы
	inputs := make([]blockchain.TransactionInput, len(tx.Inputs))
	for i, input := range tx.Inputs {
		inputs[i] = *input
	}
	iv.validateTransactionInputs(inputs, result)
	
	// Проверяем выходы
	outputs := make([]blockchain.TransactionOutput, len(tx.Outputs))
	for i, output := range tx.Outputs {
		outputs[i] = *output
	}
	iv.validateTransactionOutputs(outputs, result)
	
	// Проверяем комиссию
	iv.validateTransactionFee(tx.Fee, result)
	
	// Проверяем подпись
	iv.validateTransactionSignature(tx, result)
	
	return result
}

// ValidateAddress валидирует адрес
func (iv *InputValidator) ValidateAddress(address string) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	// Проверяем длину
	if len(address) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "address",
			Message: "Address cannot be empty",
			Code:    "EMPTY_ADDRESS",
		})
		return result
	}
	
	if len(address) > iv.maxAddressLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "address",
			Message: fmt.Sprintf("Address length %d exceeds maximum %d", len(address), iv.maxAddressLength),
			Code:    "ADDRESS_TOO_LONG",
		})
	}
	
	// Проверяем формат (должен быть hex)
	if !iv.isValidHex(address) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "address",
			Message: "Address must be valid hexadecimal",
			Code:    "INVALID_ADDRESS_FORMAT",
		})
	}
	
	// Проверяем контрольную сумму (если есть)
	if len(address) >= 4 {
		// Простая проверка контрольной суммы
		if !iv.validateChecksum(address) {
			result.Warnings = append(result.Warnings, "Address checksum validation failed")
		}
	}
	
	return result
}

// ValidateAmount валидирует сумму
func (iv *InputValidator) ValidateAmount(amount int64) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if amount < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "amount",
			Message: "Amount cannot be negative",
			Code:    "NEGATIVE_AMOUNT",
		})
	}
	
	if amount == 0 {
		result.Warnings = append(result.Warnings, "Amount is zero")
	}
	
	// Проверяем на переполнение
	if amount > 1000000000000 { // 1 триллион
		result.Warnings = append(result.Warnings, "Amount is very large")
	}
	
	return result
}

// validateBlockHeader валидирует заголовок блока
func (iv *InputValidator) validateBlockHeader(block *blockchain.Block, result *ValidationResult) {
	// Проверяем хеш предыдущего блока
	if len(block.PreviousHash) != 64 { // SHA-256 hex length
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "previous_hash",
			Message: "Previous hash must be 64 characters (SHA-256)",
			Code:    "INVALID_PREVIOUS_HASH_LENGTH",
		})
	}
	
	// Проверяем nonce
	if block.Nonce < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "nonce",
			Message: "Nonce cannot be negative",
			Code:    "NEGATIVE_NONCE",
		})
	}
	
	// Проверяем сложность
	if block.Difficulty < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "difficulty",
			Message: "Difficulty cannot be negative",
			Code:    "NEGATIVE_DIFFICULTY",
		})
	}
	
	// Проверяем высоту
	if block.Height < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "height",
			Message: "Height cannot be negative",
			Code:    "NEGATIVE_HEIGHT",
		})
	}
}

// validateTransactions валидирует транзакции в блоке
func (iv *InputValidator) validateTransactions(transactions []blockchain.Transaction, result *ValidationResult) {
	if len(transactions) == 0 {
		result.Warnings = append(result.Warnings, "Block contains no transactions")
		return
	}
	
	// Проверяем дубликаты транзакций
	txIDs := make(map[string]bool)
	for i, tx := range transactions {
		txID := string(tx.ID)
		if txIDs[txID] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("transaction_%d", i),
				Message: "Duplicate transaction ID",
				Code:    "DUPLICATE_TRANSACTION",
			})
		}
		txIDs[txID] = true
		
		// Валидируем каждую транзакцию
		txResult := iv.ValidateTransaction(&tx)
		if !txResult.Valid {
			result.Valid = false
			for _, err := range txResult.Errors {
				err.Field = fmt.Sprintf("transaction_%d.%s", i, err.Field)
				result.Errors = append(result.Errors, err)
			}
		}
	}
}

// validateTimestamp валидирует временную метку
func (iv *InputValidator) validateTimestamp(timestamp int64, result *ValidationResult) {
	now := time.Now().Unix()
	
	// Проверяем, что временная метка не в будущем (с запасом в 1 час)
	if timestamp > now+3600 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "timestamp",
			Message: "Timestamp is too far in the future",
			Code:    "FUTURE_TIMESTAMP",
		})
	}
	
	// Проверяем, что временная метка не слишком старый (более 2 часов)
	if timestamp < now-7200 {
		result.Warnings = append(result.Warnings, "Timestamp is very old")
	}
}

// validateTransactionID валидирует ID транзакции
func (iv *InputValidator) validateTransactionID(id string, result *ValidationResult) {
	if len(id) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "transaction_id",
			Message: "Transaction ID cannot be empty",
			Code:    "EMPTY_TRANSACTION_ID",
		})
		return
	}
	
	if len(id) != 64 { // SHA-256 hex length
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "transaction_id",
			Message: "Transaction ID must be 64 characters (SHA-256)",
			Code:    "INVALID_TRANSACTION_ID_LENGTH",
		})
	}
	
	if !iv.isValidHex(id) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "transaction_id",
			Message: "Transaction ID must be valid hexadecimal",
			Code:    "INVALID_TRANSACTION_ID_FORMAT",
		})
	}
}

// validateTransactionInputs валидирует входы транзакции
func (iv *InputValidator) validateTransactionInputs(inputs []blockchain.TransactionInput, result *ValidationResult) {
	if len(inputs) == 0 {
		result.Warnings = append(result.Warnings, "Transaction has no inputs")
		return
	}
	
	for i, input := range inputs {
		// Проверяем ссылку на предыдущую транзакцию
		if len(input.TransactionID) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("input_%d.transaction_id", i),
				Message: "Transaction ID cannot be empty",
				Code:    "EMPTY_TRANSACTION_ID",
			})
		}
		
		// Проверяем индекс выхода
		if input.OutputIndex < 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("input_%d.output_index", i),
				Message: "Output index cannot be negative",
				Code:    "NEGATIVE_OUTPUT_INDEX",
			})
		}
		
		// Проверяем подпись
		if len(input.Signature) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("input_%d.signature", i),
				Message: "Signature cannot be empty",
				Code:    "EMPTY_SIGNATURE",
			})
		}
	}
}

// validateTransactionOutputs валидирует выходы транзакции
func (iv *InputValidator) validateTransactionOutputs(outputs []blockchain.TransactionOutput, result *ValidationResult) {
	if len(outputs) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "outputs",
			Message: "Transaction must have at least one output",
			Code:    "NO_OUTPUTS",
		})
		return
	}
	
	for i, output := range outputs {
		// Проверяем адрес получателя
		addrResult := iv.ValidateAddress(output.Address)
		if !addrResult.Valid {
			result.Valid = false
			for _, err := range addrResult.Errors {
				err.Field = fmt.Sprintf("output_%d.address", i)
				result.Errors = append(result.Errors, err)
			}
		}
		
		// Проверяем сумму
		amountResult := iv.ValidateAmount(output.Value)
		if !amountResult.Valid {
			result.Valid = false
			for _, err := range amountResult.Errors {
				err.Field = fmt.Sprintf("output_%d.value", i)
				result.Errors = append(result.Errors, err)
			}
		}
	}
}

// validateTransactionFee валидирует комиссию транзакции
func (iv *InputValidator) validateTransactionFee(fee int64, result *ValidationResult) {
	if fee < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "fee",
			Message: "Transaction fee cannot be negative",
			Code:    "NEGATIVE_FEE",
		})
	}
	
	if fee < iv.minTransactionFee {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Transaction fee %d is below minimum %d", fee, iv.minTransactionFee))
	}
}

// validateTransactionSignature валидирует подпись транзакции
func (iv *InputValidator) validateTransactionSignature(tx *blockchain.Transaction, result *ValidationResult) {
	// Здесь можно добавить проверку подписи
	// Пока что просто проверяем, что подпись не пустая
	for i, input := range tx.Inputs {
		if len(input.Signature) == 0 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("input_%d.signature", i),
				Message: "Transaction input signature cannot be empty",
				Code:    "EMPTY_SIGNATURE",
			})
		}
	}
}

// isValidHex проверяет, является ли строка валидным hex
func (iv *InputValidator) isValidHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// validateChecksum проверяет контрольную сумму адреса
func (iv *InputValidator) validateChecksum(address string) bool {
	if len(address) < 4 {
		return false
	}
	
	// Простая проверка контрольной суммы
	// В реальной реализации здесь должна быть более сложная логика
	hash := sha256.Sum256([]byte(address[:len(address)-4]))
	expectedChecksum := hex.EncodeToString(hash[:])[:4]
	actualChecksum := address[len(address)-4:]
	
	return expectedChecksum == actualChecksum
}

// SanitizeString очищает строку от потенциально опасных символов
func (iv *InputValidator) SanitizeString(input string) string {
	// Удаляем управляющие символы
	cleaned := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1 // удаляем
		}
		return r
	}, input)
	
	// Ограничиваем длину
	if len(cleaned) > 1000 {
		cleaned = cleaned[:1000]
	}
	
	return cleaned
}

// ValidateJSONSize проверяет размер JSON данных
func (iv *InputValidator) ValidateJSONSize(data []byte, maxSize int) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if len(data) > maxSize {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json_size",
			Message: fmt.Sprintf("JSON size %d exceeds maximum %d", len(data), maxSize),
			Code:    "JSON_SIZE_EXCEEDED",
		})
	}
	
	return result
}

// ValidateNumericRange проверяет, что число находится в допустимом диапазоне
func (iv *InputValidator) ValidateNumericRange(value int64, min, max int64, fieldName string) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if value < min || value > max {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Value %d is outside allowed range [%d, %d]", value, min, max),
			Code:    "VALUE_OUT_OF_RANGE",
		})
	}
	
	return result
}
