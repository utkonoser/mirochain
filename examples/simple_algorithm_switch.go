//go:build simple_demo

package main

import (
	"fmt"
	"log"

	"mirochain/internal/crypto"
)

func main() {
	fmt.Println("=== Simple Algorithm Switching Demo ===")
	fmt.Println()

	// Создаем менеджеры
	classicManager := crypto.NewSignatureManager()
	quantumManager := crypto.NewQuantumResistantManager()

	// Тестовые данные
	data := []byte("Hello, MiroChain!")

	// 1. Показываем, как переключиться с одного классического алгоритма на другой
	fmt.Println("1. Switching between classic algorithms:")
	
	// Используем ECDSA
	fmt.Println("Using ECDSA...")
	ecdsaKey, err := classicManager.GenerateKeyPair("ecdsa")
	if err != nil {
		log.Fatal(err)
	}
	ecdsaSig, err := classicManager.Sign("ecdsa", ecdsaKey.PrivateKey, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ECDSA signature created: %s\n", ecdsaKey.Address)

	// Переключаемся на Ed25519
	fmt.Println("Switching to Ed25519...")
	ed25519Key, err := classicManager.GenerateKeyPair("ed25519")
	if err != nil {
		log.Fatal(err)
	}
	ed25519Sig, err := classicManager.Sign("ed25519", ed25519Key.PrivateKey, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Ed25519 signature created: %s\n", ed25519Key.Address)

	// Проверяем обе подписи
	ecdsaValid := classicManager.Verify(ecdsaSig, data)
	ed25519Valid := classicManager.Verify(ed25519Sig, data)
	fmt.Printf("ECDSA valid: %t, Ed25519 valid: %t\n", ecdsaValid, ed25519Valid)
	fmt.Println()

	// 2. Показываем, как переключиться с классического на квантово-устойчивый
	fmt.Println("2. Switching from classic to quantum-resistant:")
	
	// Начинаем с ECDSA
	fmt.Println("Starting with ECDSA...")
	fmt.Printf("ECDSA address: %s\n", ecdsaKey.Address)

	// Переключаемся на Dilithium
	fmt.Println("Switching to Dilithium (quantum-resistant)...")
	dilithiumKey, err := quantumManager.GenerateKeyPair("dilithium")
	if err != nil {
		log.Fatal(err)
	}
	dilithiumSig, err := quantumManager.Sign("dilithium", dilithiumKey.PrivateKey, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Dilithium address: %s\n", dilithiumKey.Address)

	// Проверяем обе подписи
	ecdsaValid = classicManager.Verify(ecdsaSig, data)
	dilithiumValid, err := quantumManager.Verify("dilithium", dilithiumKey.PublicKey, dilithiumSig, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ECDSA valid: %t, Dilithium valid: %t\n", ecdsaValid, dilithiumValid)
	fmt.Println()

	// 3. Показываем, как выбрать алгоритм на основе требований
	fmt.Println("3. Algorithm selection based on requirements:")
	
	// Для высокой производительности
	fmt.Println("For high performance (choose Ed25519):")
	ed25519Key2, _ := classicManager.GenerateKeyPair("ed25519")
	_, _ = classicManager.Sign("ed25519", ed25519Key2.PrivateKey, data)
	fmt.Printf("Ed25519 address: %s\n", ed25519Key2.Address)

	// Для максимальной безопасности
	fmt.Println("For maximum security (choose SPHINCS+):")
	sphincsKey, _ := quantumManager.GenerateKeyPair("sphincs+")
	_, _ = quantumManager.Sign("sphincs+", sphincsKey.PrivateKey, data)
	fmt.Printf("SPHINCS+ address: %s\n", sphincsKey.Address)

	// Для баланса производительности и безопасности
	fmt.Println("For balanced performance/security (choose Dilithium):")
	fmt.Printf("Dilithium address: %s\n", dilithiumKey.Address)
	fmt.Println()

	// 4. Показываем, как можно поддерживать несколько алгоритмов одновременно
	fmt.Println("4. Supporting multiple algorithms simultaneously:")
	
	// Создаем мульти-алгоритмический кошелек
	wallet := &MultiAlgorithmWallet{
		classicManager: classicManager,
		quantumManager: quantumManager,
	}
	
	// Добавляем ключи для разных алгоритмов
	wallet.addKey("ecdsa", ecdsaKey)
	wallet.addKey("ed25519", ed25519Key)
	wallet.addQuantumKey("dilithium", dilithiumKey)
	wallet.addQuantumKey("sphincs+", sphincsKey)
	
	// Демонстрируем использование
	wallet.demonstrateUsage(data)
}

// MultiAlgorithmWallet поддерживает несколько алгоритмов
type MultiAlgorithmWallet struct {
	classicKeys    map[string]*crypto.SignatureKeyPair
	quantumKeys    map[string]*crypto.QuantumResistantKeyPair
	classicManager *crypto.SignatureManager
	quantumManager *crypto.QuantumResistantManager
}

func (w *MultiAlgorithmWallet) addKey(algorithm string, keyPair *crypto.SignatureKeyPair) {
	if w.classicKeys == nil {
		w.classicKeys = make(map[string]*crypto.SignatureKeyPair)
	}
	w.classicKeys[algorithm] = keyPair
}

func (w *MultiAlgorithmWallet) addQuantumKey(algorithm string, keyPair *crypto.QuantumResistantKeyPair) {
	if w.quantumKeys == nil {
		w.quantumKeys = make(map[string]*crypto.QuantumResistantKeyPair)
	}
	w.quantumKeys[algorithm] = keyPair
}

func (w *MultiAlgorithmWallet) demonstrateUsage(data []byte) {
	fmt.Println("Available algorithms in wallet:")
	
	// Показываем классические алгоритмы
	for algo, key := range w.classicKeys {
		fmt.Printf("  Classic: %s -> %s\n", algo, key.Address)
	}
	
	// Показываем квантово-устойчивые алгоритмы
	for algo, key := range w.quantumKeys {
		fmt.Printf("  Quantum: %s -> %s\n", algo, key.Address)
	}
	
	fmt.Println()
	fmt.Println("You can use any of these algorithms for signing!")
	fmt.Println("Example: Use ECDSA for fast signing, Dilithium for quantum resistance")
}
