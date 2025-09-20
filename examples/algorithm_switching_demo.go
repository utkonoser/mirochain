//go:build algorithm_demo

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/crypto"
)

func main() {
	fmt.Println("=== Algorithm Switching Demo ===")
	fmt.Println()

	// Создаем менеджеры для разных типов алгоритмов
	classicManager := crypto.NewSignatureManager()
	quantumManager := crypto.NewQuantumResistantManager()

	// Тестовые данные
	testData := []byte("Hello, Algorithm Switching World!")
	fmt.Printf("Test data: %s\n", string(testData))
	fmt.Println()

	// 1. Классические алгоритмы
	fmt.Println("=== Classic Algorithms ===")
	classicAlgorithms := classicManager.GetSupportedAlgorithms()
	for _, algo := range classicAlgorithms {
		fmt.Printf("--- Testing %s ---\n", algo)
		
		keyPair, err := classicManager.GenerateKeyPair(algo)
		if err != nil {
			log.Printf("Error generating key pair for %s: %v", algo, err)
			continue
		}
		
		fmt.Printf("Address: %s\n", keyPair.Address)
		
		// Подписываем
		start := time.Now()
		signature, err := classicManager.Sign(algo, keyPair.PrivateKey, testData)
		signTime := time.Since(start)
		
		if err != nil {
			log.Printf("Error signing with %s: %v", algo, err)
			continue
		}
		
		// Проверяем
		start = time.Now()
		valid := classicManager.Verify(signature, testData)
		verifyTime := time.Since(start)
		
		fmt.Printf("Sign time: %v, Verify time: %v, Valid: %t\n", signTime, verifyTime, valid)
		fmt.Println()
	}

	// 2. Квантово-устойчивые алгоритмы
	fmt.Println("=== Quantum-Resistant Algorithms ===")
	quantumAlgorithms := quantumManager.GetSupportedAlgorithms()
	for _, algo := range quantumAlgorithms {
		fmt.Printf("--- Testing %s ---\n", algo)
		
		keyPair, err := quantumManager.GenerateKeyPair(algo)
		if err != nil {
			log.Printf("Error generating key pair for %s: %v", algo, err)
			continue
		}
		
		fmt.Printf("Address: %s\n", keyPair.Address)
		
		// Подписываем
		start := time.Now()
		signature, err := quantumManager.Sign(algo, keyPair.PrivateKey, testData)
		signTime := time.Since(start)
		
		if err != nil {
			log.Printf("Error signing with %s: %v", algo, err)
			continue
		}
		
		// Проверяем
		start = time.Now()
		valid, err := quantumManager.Verify(algo, keyPair.PublicKey, signature, testData)
		verifyTime := time.Since(start)
		
		if err != nil {
			log.Printf("Error verifying with %s: %v", algo, err)
			continue
		}
		
		fmt.Printf("Sign time: %v, Verify time: %v, Valid: %t\n", signTime, verifyTime, valid)
		fmt.Println()
	}

	// 3. Демонстрация переключения алгоритмов
	fmt.Println("=== Algorithm Switching in Practice ===")
	
	// Создаем кошелек с разными алгоритмами
	wallet := createMultiAlgorithmWallet()
	
	// Демонстрируем использование разных алгоритмов
	wallet.demonstrateAlgorithmUsage(testData)
	
	// 4. Рекомендации по выбору алгоритма
	fmt.Println("=== Algorithm Selection Guidelines ===")
	printAlgorithmGuidelines()
}

// MultiAlgorithmWallet демонстрирует использование разных алгоритмов
type MultiAlgorithmWallet struct {
	classicKeys    map[string]*crypto.SignatureKeyPair
	quantumKeys    map[string]*crypto.QuantumResistantKeyPair
	classicManager *crypto.SignatureManager
	quantumManager *crypto.QuantumResistantManager
}

func createMultiAlgorithmWallet() *MultiAlgorithmWallet {
	return &MultiAlgorithmWallet{
		classicKeys:    make(map[string]*crypto.SignatureKeyPair),
		quantumKeys:    make(map[string]*crypto.QuantumResistantKeyPair),
		classicManager: crypto.NewSignatureManager(),
		quantumManager: crypto.NewQuantumResistantManager(),
	}
}

func (w *MultiAlgorithmWallet) demonstrateAlgorithmUsage(data []byte) {
	fmt.Println("Creating wallet with multiple algorithms...")
	
	// Генерируем ключи для разных алгоритмов
	classicAlgos := w.classicManager.GetSupportedAlgorithms()
	for _, algo := range classicAlgos {
		keyPair, err := w.classicManager.GenerateKeyPair(algo)
		if err != nil {
			continue
		}
		w.classicKeys[string(algo)] = keyPair
		fmt.Printf("Generated %s key: %s\n", algo, keyPair.Address)
	}
	
	quantumAlgos := w.quantumManager.GetSupportedAlgorithms()
	for _, algo := range quantumAlgos {
		keyPair, err := w.quantumManager.GenerateKeyPair(algo)
		if err != nil {
			continue
		}
		w.quantumKeys[string(algo)] = keyPair
		fmt.Printf("Generated %s key: %s\n", algo, keyPair.Address)
	}
	
	fmt.Println()
	
	// Демонстрируем подписание одним алгоритмом и проверку другим
	fmt.Println("Cross-algorithm compatibility test:")
	
	// Используем ECDSA для подписи
	ecdsaKey := w.classicKeys["ecdsa"]
	if ecdsaKey != nil {
		signature, err := w.classicManager.Sign("ecdsa", ecdsaKey.PrivateKey, data)
		if err == nil {
			valid := w.classicManager.Verify(signature, data)
			if valid {
				fmt.Printf("✅ ECDSA signature verified successfully\n")
			}
		}
	}
	
	// Используем Dilithium для подписи
	dilithiumKey := w.quantumKeys["dilithium"]
	if dilithiumKey != nil {
		signature, err := w.quantumManager.Sign("dilithium", dilithiumKey.PrivateKey, data)
		if err == nil {
			valid, err := w.quantumManager.Verify("dilithium", dilithiumKey.PublicKey, signature, data)
			if err == nil && valid {
				fmt.Printf("✅ Dilithium signature verified successfully\n")
			}
		}
	}
	
	fmt.Println()
}

func printAlgorithmGuidelines() {
	fmt.Println("📋 Algorithm Selection Guidelines:")
	fmt.Println()
	
	fmt.Println("🔐 Classic Algorithms (Fast, but vulnerable to quantum attacks):")
	fmt.Println("  • ECDSA - Fast, widely supported, good for current use")
	fmt.Println("  • Ed25519 - Very fast, good for high-throughput systems")
	fmt.Println("  • RSA - Slow, large keys, but very well established")
	fmt.Println("  • Schnorr - Fast, good for multi-signatures")
	fmt.Println()
	
	fmt.Println("🛡️ Quantum-Resistant Algorithms (Future-proof, but larger):")
	fmt.Println("  • SPHINCS+ - Most secure, large signatures, stateless")
	fmt.Println("  • Dilithium - Good balance, moderate size, lattice-based")
	fmt.Println("  • Falcon - Compact signatures, lattice-based")
	fmt.Println("  • XMSS - Stateful, good for limited resources")
	fmt.Println("  • LMS - Stateful, good for IoT devices")
	fmt.Println()
	
	fmt.Println("💡 Recommendations:")
	fmt.Println("  • For current production: Use ECDSA or Ed25519")
	fmt.Println("  • For future-proofing: Use Dilithium or Falcon")
	fmt.Println("  • For maximum security: Use SPHINCS+")
	fmt.Println("  • For resource-constrained: Use XMSS or LMS")
	fmt.Println("  • For hybrid approach: Support both classic and quantum-resistant")
	fmt.Println()
	
	fmt.Println("🔄 Switching Strategies:")
	fmt.Println("  1. Gradual migration: Start with classic, add quantum-resistant")
	fmt.Println("  2. Dual support: Support both types simultaneously")
	fmt.Println("  3. Algorithm negotiation: Let clients choose their preferred algorithm")
	fmt.Println("  4. Time-based switching: Use classic now, quantum-resistant later")
}
