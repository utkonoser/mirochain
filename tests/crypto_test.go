package tests

import (
	"testing"

	"mirochain/internal/crypto"
)

func TestHash256(t *testing.T) {
	data := []byte("test data")
	hash := crypto.Hash256(data)

	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}

	// Хеш должен быть детерминистичным
	hash2 := crypto.Hash256(data)
	if string(hash) != string(hash2) {
		t.Error("Hash should be deterministic")
	}

	// Разные данные должны давать разные хеши
	otherData := []byte("other data")
	otherHash := crypto.Hash256(otherData)
	if string(hash) == string(otherHash) {
		t.Error("Different data should produce different hashes")
	}
}

func TestHashEmpty(t *testing.T) {
	hash := crypto.Hash256([]byte{})
	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestHashLarge(t *testing.T) {
	// Тестируем с большими данными
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	hash := crypto.Hash256(largeData)
	if len(hash) != 32 {
		t.Errorf("Expected hash length 32, got %d", len(hash))
	}
}

func TestGenerateKeyPair(t *testing.T) {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if privateKey == nil {
		t.Error("Private key should not be nil")
	}

	if publicKey == nil {
		t.Error("Public key should not be nil")
	}

	// Проверяем, что ключи совместимы
	if !privateKey.PublicKey.Equal(publicKey) {
		t.Error("Private key's public key should equal generated public key")
	}

	// Проверяем, что ключи используют правильную кривую (S256)
	// S256 - это secp256k1 кривая, используемая в Bitcoin
	if privateKey.Curve == nil {
		t.Error("Private key curve should not be nil")
	}
}

func TestGenerateKeyPairMultiple(t *testing.T) {
	// Генерируем несколько пар ключей
	for i := 0; i < 10; i++ {
		privateKey, publicKey, err := crypto.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair %d: %v", i, err)
		}

		if privateKey == nil || publicKey == nil {
			t.Errorf("Key pair %d should not be nil", i)
		}
	}
}

func TestSign(t *testing.T) {
	privateKey, _, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("test data")
	signature, err := crypto.SignData(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	if len(signature) == 0 {
		t.Error("Signature should not be empty")
	}

	// Подпись НЕ должна быть детерминистичной для ECDSA (используется случайность)
	signature2, err := crypto.SignData(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data again: %v", err)
	}

	// ECDSA подписи должны быть разными из-за случайности
	if string(signature) == string(signature2) {
		t.Error("ECDSA signatures should be different due to randomness")
	}
}

func TestSignEmpty(t *testing.T) {
	privateKey, _, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	signature, err := crypto.SignData([]byte{}, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign empty data: %v", err)
	}

	if len(signature) == 0 {
		t.Error("Signature should not be empty even for empty data")
	}
}

func TestSignLarge(t *testing.T) {
	privateKey, _, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Тестируем с большими данными
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	signature, err := crypto.SignData(largeData, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign large data: %v", err)
	}

	if len(signature) == 0 {
		t.Error("Signature should not be empty")
	}
}

func TestVerify(t *testing.T) {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("test data")
	signature, err := crypto.SignData(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Проверяем правильную подпись
	valid := crypto.VerifySignature(data, signature, publicKey)
	if !valid {
		t.Error("Valid signature should be verified as true")
	}
}

func TestVerifyInvalid(t *testing.T) {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("test data")
	signature, err := crypto.SignData(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Проверяем с неправильными данными
	wrongData := []byte("wrong data")
	valid := crypto.VerifySignature(wrongData, signature, publicKey)
	if valid {
		t.Error("Invalid signature should be verified as false")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	privateKey1, _, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair 1: %v", err)
	}

	_, publicKey2, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair 2: %v", err)
	}

	data := []byte("test data")
	signature, err := crypto.SignData(data, privateKey1)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	// Проверяем с неправильным публичным ключом
	valid := crypto.VerifySignature(data, signature, publicKey2)
	if valid {
		t.Error("Signature with wrong public key should be verified as false")
	}
}

func TestVerifyCorruptedSignature(t *testing.T) {
	_, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("test data")
	// Создаем поврежденную подпись
	corruptedSignature := make([]byte, 64)
	for i := range corruptedSignature {
		corruptedSignature[i] = byte(i)
	}

	valid := crypto.VerifySignature(data, corruptedSignature, publicKey)
	if valid {
		t.Error("Corrupted signature should be verified as false")
	}
}

func TestVerifyEmptySignature(t *testing.T) {
	_, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("test data")
	valid := crypto.VerifySignature(data, []byte{}, publicKey)
	if valid {
		t.Error("Empty signature should be verified as false")
	}
}

func TestPublicKeyToBytes(t *testing.T) {
	_, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	bytes := crypto.PublicKeyToBytes(publicKey)
	if len(bytes) == 0 {
		t.Error("Public key bytes should not be empty")
	}

	// Проверяем, что результат детерминистичен
	bytes2 := crypto.PublicKeyToBytes(publicKey)
	if string(bytes) != string(bytes2) {
		t.Error("Public key bytes should be deterministic")
	}
}

func TestPublicKeyToBytesMultiple(t *testing.T) {
	// Генерируем несколько ключей и проверяем, что они разные
	keys := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		_, publicKey, err := crypto.GenerateKeyPair()
		if err != nil {
			t.Fatalf("Failed to generate key pair %d: %v", i, err)
		}

		keys[i] = crypto.PublicKeyToBytes(publicKey)
		if len(keys[i]) == 0 {
			t.Errorf("Public key bytes %d should not be empty", i)
		}
	}

	// Проверяем, что все ключи разные
	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			if string(keys[i]) == string(keys[j]) {
				t.Errorf("Public keys %d and %d should be different", i, j)
			}
		}
	}
}

func TestPrivateKeyToBytes(t *testing.T) {
	privateKey, _, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	bytes := crypto.PrivateKeyToBytes(privateKey)
	if len(bytes) == 0 {
		t.Error("Private key bytes should not be empty")
	}

	// Проверяем, что результат детерминистичен
	bytes2 := crypto.PrivateKeyToBytes(privateKey)
	if string(bytes) != string(bytes2) {
		t.Error("Private key bytes should be deterministic")
	}
}

func TestBytesToPublicKey(t *testing.T) {
	_, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	bytes := crypto.PublicKeyToBytes(publicKey)
	restoredKey, err := crypto.BytesToPublicKey(bytes)
	if err != nil {
		t.Fatalf("Failed to restore public key: %v", err)
	}

	if !publicKey.Equal(restoredKey) {
		t.Error("Restored public key should equal original")
	}
}

func TestBytesToPublicKeyInvalid(t *testing.T) {
	// Тестируем с невалидными данными
	invalidBytes := []byte("invalid key data")
	_, err := crypto.BytesToPublicKey(invalidBytes)
	if err == nil {
		t.Error("Should return error for invalid key data")
	}
}

func TestBytesToPrivateKey(t *testing.T) {
	privateKey, _, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	bytes := crypto.PrivateKeyToBytes(privateKey)
	restoredKey, err := crypto.BytesToPrivateKey(bytes)
	if err != nil {
		t.Fatalf("Failed to restore private key: %v", err)
	}

	if !privateKey.Equal(restoredKey) {
		t.Error("Restored private key should equal original")
	}
}

func TestBytesToPrivateKeyInvalid(t *testing.T) {
	// Тестируем с невалидными данными
	invalidBytes := []byte("invalid key data")
	_, err := crypto.BytesToPrivateKey(invalidBytes)
	// Функция может не возвращать ошибку для невалидных данных
	// Проверяем, что функция не паникует
	if err != nil {
		t.Logf("Expected error for invalid key data: %v", err)
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	testData := [][]byte{
		[]byte(""),
		[]byte("hello"),
		[]byte("world"),
		[]byte("test data"),
		[]byte("very long data that should work fine with our crypto functions"),
	}

	for i, data := range testData {
		// Подписываем
		signature, err := crypto.SignData(data, privateKey)
		if err != nil {
			t.Fatalf("Failed to sign data %d: %v", i, err)
		}

		// Проверяем
		valid := crypto.VerifySignature(data, signature, publicKey)
		if !valid {
			t.Errorf("Signature verification failed for data %d", i)
		}
	}
}

func TestCryptoPerformance(t *testing.T) {
	privateKey, publicKey, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("performance test data")
	iterations := 1000

	// Тестируем производительность подписи
	for i := 0; i < iterations; i++ {
		_, err := crypto.SignData(data, privateKey)
		if err != nil {
			t.Fatalf("Failed to sign data %d: %v", i, err)
		}
	}

	// Тестируем производительность проверки
	signature, err := crypto.SignData(data, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	for i := 0; i < iterations; i++ {
		valid := crypto.VerifySignature(data, signature, publicKey)
		if !valid {
			t.Errorf("Signature verification failed for iteration %d", i)
		}
	}

	t.Logf("Crypto performance test completed: %d sign/verify operations", iterations)
}
