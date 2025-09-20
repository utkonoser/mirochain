package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
)

// Hash256 вычисляет двойной SHA-256 хеш
func Hash256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

// Hash160 вычисляет RIPEMD160(SHA256(data))
func Hash160(data []byte) []byte {
	hash := sha256.Sum256(data)
	// В реальной реализации здесь должен быть RIPEMD160
	// Пока что используем первые 20 байт SHA256
	return hash[:20]
}

// GenerateKeyPair генерирует пару ключей ECDSA
func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

// SignData подписывает данные приватным ключом
func SignData(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	// Вычисляем хеш данных
	hash := sha256.Sum256(data)

	// Подписываем хеш
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Кодируем подпись
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature проверяет подпись
func VerifySignature(data []byte, signature []byte, publicKey *ecdsa.PublicKey) bool {
	// Вычисляем хеш данных
	hash := sha256.Sum256(data)

	// Декодируем подпись
	if len(signature) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	// Проверяем подпись
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// PublicKeyToAddress конвертирует публичный ключ в адрес
func PublicKeyToAddress(publicKey *ecdsa.PublicKey) (string, error) {
	// Конвертируем публичный ключ в байты
	publicKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Вычисляем хеш публичного ключа
	hash := Hash160(publicKeyBytes)

	// Создаем адрес
	address := hex.EncodeToString(hash)
	return address, nil
}

// PublicKeyToBytes конвертирует публичный ключ в байты
func PublicKeyToBytes(publicKey *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

// BytesToPublicKey конвертирует байты в публичный ключ
func BytesToPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(btcec.S256(), data)
	if x == nil || y == nil {
		return nil, fmt.Errorf("invalid public key data")
	}

	publicKey := &ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     x,
		Y:     y,
	}

	return publicKey, nil
}

// PrivateKeyToBytes конвертирует приватный ключ в байты
func PrivateKeyToBytes(privateKey *ecdsa.PrivateKey) []byte {
	return privateKey.D.Bytes()
}

// BytesToPrivateKey конвертирует байты в приватный ключ
func BytesToPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	d := new(big.Int).SetBytes(data)

	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: btcec.S256(),
		},
		D: d,
	}

	// Вычисляем публичный ключ
	privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.Curve.ScalarBaseMult(d.Bytes())

	return privateKey, nil
}

// MerkleRoot вычисляет Merkle root для списка хешей
func MerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return []byte{}
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// Создаем следующий уровень
	var nextLevel [][]byte
	for i := 0; i < len(hashes); i += 2 {
		var hash []byte
		if i+1 < len(hashes) {
			// Объединяем два хеша
			combined := append(hashes[i], hashes[i+1]...)
			hash = Hash256(combined)
		} else {
			// Нечетное количество - дублируем последний хеш
			combined := append(hashes[i], hashes[i]...)
			hash = Hash256(combined)
		}
		nextLevel = append(nextLevel, hash)
	}

	return MerkleRoot(nextLevel)
}

// DoubleHash вычисляет двойной хеш (как в Bitcoin)
func DoubleHash(data []byte) []byte {
	return Hash256(data)
}

// HashFromString конвертирует hex строку в хеш
func HashFromString(hashStr string) ([]byte, error) {
	return hex.DecodeString(hashStr)
}

// HashToString конвертирует хеш в hex строку
func HashToString(hash []byte) string {
	return hex.EncodeToString(hash)
}

// IsValidAddress проверяет валидность адреса
func IsValidAddress(address string) bool {
	// Проверяем, что адрес является hex строкой
	_, err := hex.DecodeString(address)
	if err != nil {
		return false
	}

	// Проверяем длину (20 байт = 40 hex символов)
	return len(address) == 40
}

// GenerateRandomBytes генерирует случайные байты
func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// ChainHash вычисляет хеш цепочки (как в Bitcoin)
func ChainHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	// В Bitcoin используется двойной хеш, но с обратным порядком байтов
	// Для простоты используем обычный двойной хеш
	return Hash256(hash[:])
}
