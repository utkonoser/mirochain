//go:build quantum_demo

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/crypto"
)

func main() {
	fmt.Println("=== Quantum-Resistant Cryptography Demo ===")
	fmt.Println()

	// Создаем менеджер квантово-устойчивых алгоритмов
	manager := crypto.NewQuantumResistantManager()
	
	// Получаем список поддерживаемых алгоритмов
	algorithms := manager.GetSupportedAlgorithms()
	fmt.Printf("Supported quantum-resistant algorithms: %v\n", algorithms)
	fmt.Println()

	// Тестовые данные
	testData := []byte("Hello, Quantum-Resistant World!")
	fmt.Printf("Test data: %s\n", string(testData))
	fmt.Println()

	// Демонстрируем каждый алгоритм
	for _, algo := range algorithms {
		fmt.Printf("--- Testing %s ---\n", algo)
		
		// Генерируем пару ключей
		keyPair, err := manager.GenerateKeyPair(algo)
		if err != nil {
			log.Printf("Error generating key pair for %s: %v", algo, err)
			continue
		}
		
		fmt.Printf("Address: %s\n", keyPair.Address)
		fmt.Printf("Public key size: %d bytes\n", len(keyPair.PublicKey))
		fmt.Printf("Private key size: %d bytes\n", len(keyPair.PrivateKey))
		
		// Показываем параметры алгоритма
		if keyPair.Params != nil {
			fmt.Printf("Parameters: %v\n", keyPair.Params)
		}
		
		// Подписываем данные
		start := time.Now()
		signature, err := manager.Sign(algo, keyPair.PrivateKey, testData)
		signTime := time.Since(start)
		
		if err != nil {
			log.Printf("Error signing with %s: %v", algo, err)
			continue
		}
		
		fmt.Printf("Signature size: %d bytes\n", len(signature))
		fmt.Printf("Sign time: %v\n", signTime)
		
		// Проверяем подпись
		start = time.Now()
		valid, err := manager.Verify(algo, keyPair.PublicKey, signature, testData)
		verifyTime := time.Since(start)
		
		if err != nil {
			log.Printf("Error verifying with %s: %v", algo, err)
			continue
		}
		
		fmt.Printf("Verify time: %v\n", verifyTime)
		fmt.Printf("Signature valid: %t\n", valid)
		fmt.Println()
	}

	// Сравниваем производительность алгоритмов
	fmt.Println("=== Performance Comparison ===")
	comparison := crypto.NewQuantumResistantComparison()
	metrics := comparison.CompareAlgorithms()
	
	fmt.Printf("%-15s %-10s %-12s %-12s %-12s %-15s\n", 
		"Algorithm", "Key Size", "Sig Size", "Sign Time", "Verify Time", "Security Level")
	fmt.Println("--------------------------------------------------------------------------------")
	
	for algo, metric := range metrics {
		fmt.Printf("%-15s %-10d %-12d %-12d %-12d %-15d\n",
			algo, metric.KeySize, metric.SignatureSize, 
			metric.SignTime, metric.VerifyTime, metric.SecurityLevel)
	}
	
	fmt.Println()
	
	// Получаем рекомендации
	recommendations := comparison.GetRecommendations()
	fmt.Println("=== Recommendations ===")
	for key, value := range recommendations {
		fmt.Printf("%s: %v\n", key, value)
	}
	
	fmt.Println()
	fmt.Println("=== Quantum-Resistant Features ===")
	fmt.Println("✅ SPHINCS+ - Stateless hash-based signatures")
	fmt.Println("✅ Dilithium - Lattice-based signatures")
	fmt.Println("✅ Falcon - Lattice-based signatures")
	fmt.Println("✅ XMSS - Stateful hash-based signatures")
	fmt.Println("✅ LMS - Leighton-Micali signatures")
	fmt.Println()
	fmt.Println("All algorithms are quantum-resistant and suitable for post-quantum cryptography!")
}
