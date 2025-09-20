package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

// SignatureAlgorithm представляет алгоритм подписи
type SignatureAlgorithm string

const (
	AlgorithmECDSA    SignatureAlgorithm = "ecdsa"
	AlgorithmEd25519  SignatureAlgorithm = "ed25519"
	AlgorithmRSA      SignatureAlgorithm = "rsa"
	AlgorithmSchnorr  SignatureAlgorithm = "schnorr"
)

// SignatureKeyPair представляет пару ключей для подписи
type SignatureKeyPair struct {
	Algorithm SignatureAlgorithm `json:"algorithm"`
	PublicKey []byte            `json:"public_key"`
	PrivateKey []byte           `json:"private_key"`
	Address   string            `json:"address"`
}

// Signature представляет подпись
type Signature struct {
	Algorithm SignatureAlgorithm `json:"algorithm"`
	Data      []byte            `json:"data"`
	PublicKey []byte            `json:"public_key"`
}

// SignatureManager управляет алгоритмами подписи
type SignatureManager struct {
	algorithms map[SignatureAlgorithm]SignatureAlgorithmInterface
}

// SignatureAlgorithmInterface определяет интерфейс для алгоритмов подписи
type SignatureAlgorithmInterface interface {
	GenerateKeyPair() (*SignatureKeyPair, error)
	Sign(privateKey []byte, data []byte) ([]byte, error)
	Verify(publicKey []byte, data []byte, signature []byte) bool
	GetAddress(publicKey []byte) (string, error)
	GetAlgorithm() SignatureAlgorithm
}

// NewSignatureManager создает новый менеджер алгоритмов подписи
func NewSignatureManager() *SignatureManager {
	manager := &SignatureManager{
		algorithms: make(map[SignatureAlgorithm]SignatureAlgorithmInterface),
	}
	
	// Регистрируем алгоритмы
	manager.RegisterAlgorithm(AlgorithmECDSA, &ECDSAAlgorithm{})
	manager.RegisterAlgorithm(AlgorithmEd25519, &Ed25519Algorithm{})
	manager.RegisterAlgorithm(AlgorithmRSA, &RSAAlgorithm{})
	manager.RegisterAlgorithm(AlgorithmSchnorr, &SchnorrAlgorithm{})
	
	return manager
}

// RegisterAlgorithm регистрирует алгоритм подписи
func (sm *SignatureManager) RegisterAlgorithm(algorithm SignatureAlgorithm, impl SignatureAlgorithmInterface) {
	sm.algorithms[algorithm] = impl
}

// GenerateKeyPair генерирует пару ключей для указанного алгоритма
func (sm *SignatureManager) GenerateKeyPair(algorithm SignatureAlgorithm) (*SignatureKeyPair, error) {
	impl, exists := sm.algorithms[algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm %s not supported", algorithm)
	}
	
	return impl.GenerateKeyPair()
}

// Sign подписывает данные
func (sm *SignatureManager) Sign(algorithm SignatureAlgorithm, privateKey []byte, data []byte) (*Signature, error) {
	impl, exists := sm.algorithms[algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm %s not supported", algorithm)
	}
	
	signatureData, err := impl.Sign(privateKey, data)
	if err != nil {
		return nil, err
	}
	
	// Получаем публичный ключ из приватного
	publicKey, err := sm.getPublicKeyFromPrivate(algorithm, privateKey)
	if err != nil {
		return nil, err
	}
	
	return &Signature{
		Algorithm: algorithm,
		Data:      signatureData,
		PublicKey: publicKey,
	}, nil
}

// Verify проверяет подпись
func (sm *SignatureManager) Verify(signature *Signature, data []byte) bool {
	impl, exists := sm.algorithms[signature.Algorithm]
	if !exists {
		return false
	}
	
	return impl.Verify(signature.PublicKey, data, signature.Data)
}

// GetAddress получает адрес из публичного ключа
func (sm *SignatureManager) GetAddress(algorithm SignatureAlgorithm, publicKey []byte) (string, error) {
	impl, exists := sm.algorithms[algorithm]
	if !exists {
		return "", fmt.Errorf("algorithm %s not supported", algorithm)
	}
	
	return impl.GetAddress(publicKey)
}

// getPublicKeyFromPrivate получает публичный ключ из приватного
func (sm *SignatureManager) getPublicKeyFromPrivate(algorithm SignatureAlgorithm, privateKey []byte) ([]byte, error) {
	impl, exists := sm.algorithms[algorithm]
	if !exists {
		return nil, fmt.Errorf("algorithm %s not supported", algorithm)
	}
	
	// Для получения публичного ключа создаем временную пару ключей
	keyPair, err := impl.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	
	// В реальной реализации здесь должна быть логика извлечения публичного ключа
	// из приватного ключа для каждого алгоритма
	return keyPair.PublicKey, nil
}

// ECDSAAlgorithm реализует ECDSA алгоритм
type ECDSAAlgorithm struct{}

func (e *ECDSAAlgorithm) GenerateKeyPair() (*SignatureKeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	
	publicKey := &privateKey.PublicKey
	publicKeyBytes := elliptic.Marshal(elliptic.P256(), publicKey.X, publicKey.Y)
	privateKeyBytes := privateKey.D.Bytes()
	
	address, err := e.GetAddress(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	
	return &SignatureKeyPair{
		Algorithm:  AlgorithmECDSA,
		PublicKey:  publicKeyBytes,
		PrivateKey: privateKeyBytes,
		Address:    address,
	}, nil
}

func (e *ECDSAAlgorithm) Sign(privateKey []byte, data []byte) ([]byte, error) {
	// В реальной реализации здесь должна быть логика подписи ECDSA
	// Пока что возвращаем заглушку
	hash := sha256.Sum256(data)
	return hash[:], nil
}

func (e *ECDSAAlgorithm) Verify(publicKey []byte, data []byte, signature []byte) bool {
	// В реальной реализации здесь должна быть логика проверки ECDSA
	// Пока что возвращаем заглушку
	hash := sha256.Sum256(data)
	return len(signature) == len(hash)
}

func (e *ECDSAAlgorithm) GetAddress(publicKey []byte) (string, error) {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:]), nil
}

func (e *ECDSAAlgorithm) GetAlgorithm() SignatureAlgorithm {
	return AlgorithmECDSA
}

// Ed25519Algorithm реализует Ed25519 алгоритм
type Ed25519Algorithm struct{}

func (e *Ed25519Algorithm) GenerateKeyPair() (*SignatureKeyPair, error) {
	// В реальной реализации здесь должна быть генерация Ed25519 ключей
	// Пока что используем заглушку
	publicKey := make([]byte, 32)
	privateKey := make([]byte, 64)
	
	address, err := e.GetAddress(publicKey)
	if err != nil {
		return nil, err
	}
	
	return &SignatureKeyPair{
		Algorithm:  AlgorithmEd25519,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

func (e *Ed25519Algorithm) Sign(privateKey []byte, data []byte) ([]byte, error) {
	// В реальной реализации здесь должна быть логика подписи Ed25519
	// Пока что возвращаем заглушку
	hash := sha512.Sum512(data)
	return hash[:], nil
}

func (e *Ed25519Algorithm) Verify(publicKey []byte, data []byte, signature []byte) bool {
	// В реальной реализации здесь должна быть логика проверки Ed25519
	// Пока что возвращаем заглушку
	hash := sha512.Sum512(data)
	return len(signature) == len(hash)
}

func (e *Ed25519Algorithm) GetAddress(publicKey []byte) (string, error) {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:]), nil
}

func (e *Ed25519Algorithm) GetAlgorithm() SignatureAlgorithm {
	return AlgorithmEd25519
}

// RSAAlgorithm реализует RSA алгоритм
type RSAAlgorithm struct{}

func (r *RSAAlgorithm) GenerateKeyPair() (*SignatureKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	
	publicKey := &privateKey.PublicKey
	publicKeyBytes := []byte(fmt.Sprintf("%x", publicKey.N))
	privateKeyBytes := []byte(fmt.Sprintf("%x", privateKey.D))
	
	address, err := r.GetAddress(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	
	return &SignatureKeyPair{
		Algorithm:  AlgorithmRSA,
		PublicKey:  publicKeyBytes,
		PrivateKey: privateKeyBytes,
		Address:    address,
	}, nil
}

func (r *RSAAlgorithm) Sign(privateKey []byte, data []byte) ([]byte, error) {
	// В реальной реализации здесь должна быть логика подписи RSA
	// Пока что возвращаем заглушку
	hash := sha256.Sum256(data)
	return hash[:], nil
}

func (r *RSAAlgorithm) Verify(publicKey []byte, data []byte, signature []byte) bool {
	// В реальной реализации здесь должна быть логика проверки RSA
	// Пока что возвращаем заглушку
	hash := sha256.Sum256(data)
	return len(signature) == len(hash)
}

func (r *RSAAlgorithm) GetAddress(publicKey []byte) (string, error) {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:]), nil
}

func (r *RSAAlgorithm) GetAlgorithm() SignatureAlgorithm {
	return AlgorithmRSA
}

// SchnorrAlgorithm реализует Schnorr алгоритм
type SchnorrAlgorithm struct{}

func (s *SchnorrAlgorithm) GenerateKeyPair() (*SignatureKeyPair, error) {
	// В реальной реализации здесь должна быть логика генерации ключей Schnorr
	// Пока что используем ECDSA как основу
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	
	publicKey := &privateKey.PublicKey
	publicKeyBytes := elliptic.Marshal(elliptic.P256(), publicKey.X, publicKey.Y)
	privateKeyBytes := privateKey.D.Bytes()
	
	address, err := s.GetAddress(publicKeyBytes)
	if err != nil {
		return nil, err
	}
	
	return &SignatureKeyPair{
		Algorithm:  AlgorithmSchnorr,
		PublicKey:  publicKeyBytes,
		PrivateKey: privateKeyBytes,
		Address:    address,
	}, nil
}

func (s *SchnorrAlgorithm) Sign(privateKey []byte, data []byte) ([]byte, error) {
	// В реальной реализации здесь должна быть логика подписи Schnorr
	// Пока что возвращаем заглушку
	hash := sha256.Sum256(data)
	return hash[:], nil
}

func (s *SchnorrAlgorithm) Verify(publicKey []byte, data []byte, signature []byte) bool {
	// В реальной реализации здесь должна быть логика проверки Schnorr
	// Пока что возвращаем заглушку
	hash := sha256.Sum256(data)
	return len(signature) == len(hash)
}

func (s *SchnorrAlgorithm) GetAddress(publicKey []byte) (string, error) {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:]), nil
}

func (s *SchnorrAlgorithm) GetAlgorithm() SignatureAlgorithm {
	return AlgorithmSchnorr
}

// GetSupportedAlgorithms возвращает список поддерживаемых алгоритмов
func (sm *SignatureManager) GetSupportedAlgorithms() []SignatureAlgorithm {
	algorithms := make([]SignatureAlgorithm, 0, len(sm.algorithms))
	for algorithm := range sm.algorithms {
		algorithms = append(algorithms, algorithm)
	}
	return algorithms
}

// GetAlgorithmInfo возвращает информацию об алгоритме
func (sm *SignatureManager) GetAlgorithmInfo(algorithm SignatureAlgorithm) map[string]interface{} {
	info := map[string]interface{}{
		"name":        string(algorithm),
		"supported":   false,
		"description": "",
		"key_size":    0,
		"signature_size": 0,
	}
	
	switch algorithm {
	case AlgorithmECDSA:
		info["supported"] = true
		info["description"] = "Elliptic Curve Digital Signature Algorithm"
		info["key_size"] = 256
		info["signature_size"] = 64
	case AlgorithmEd25519:
		info["supported"] = true
		info["description"] = "Edwards Curve Digital Signature Algorithm"
		info["key_size"] = 256
		info["signature_size"] = 64
	case AlgorithmRSA:
		info["supported"] = true
		info["description"] = "Rivest-Shamir-Adleman Digital Signature Algorithm"
		info["key_size"] = 2048
		info["signature_size"] = 256
	case AlgorithmSchnorr:
		info["supported"] = true
		info["description"] = "Schnorr Digital Signature Algorithm"
		info["key_size"] = 256
		info["signature_size"] = 64
	}
	
	return info
}
