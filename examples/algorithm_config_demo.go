//go:build config_demo

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"mirochain/internal/crypto"
)

// AlgorithmConfig представляет конфигурацию алгоритмов
type AlgorithmConfig struct {
	DefaultClassicAlgorithm    string   `json:"default_classic_algorithm"`
	DefaultQuantumAlgorithm    string   `json:"default_quantum_algorithm"`
	EnabledClassicAlgorithms   []string `json:"enabled_classic_algorithms"`
	EnabledQuantumAlgorithms   []string `json:"enabled_quantum_algorithms"`
	AlgorithmSelectionStrategy string   `json:"algorithm_selection_strategy"`
	MigrationSettings          struct {
		EnableMigration      bool   `json:"enable_migration"`
		MigrationStartDate   string `json:"migration_start_date"`
		MigrationEndDate     string `json:"migration_end_date"`
		ClassicAlgorithm     string `json:"classic_algorithm"`
		QuantumAlgorithm     string `json:"quantum_algorithm"`
	} `json:"migration_settings"`
}

func main() {
	fmt.Println("=== Algorithm Configuration Demo ===")
	fmt.Println()

	// 1. Создаем конфигурацию по умолчанию
	config := createDefaultConfig()
	
	// 2. Сохраняем конфигурацию в файл
	saveConfig(config, "algorithm_config.json")
	
	// 3. Загружаем конфигурацию из файла
	loadedConfig := loadConfig("algorithm_config.json")
	
	// 4. Демонстрируем использование конфигурации
	demonstrateConfigUsage(loadedConfig)
	
	// 5. Показываем примеры разных стратегий
	showAlgorithmStrategies()
}

func createDefaultConfig() *AlgorithmConfig {
	config := &AlgorithmConfig{
		DefaultClassicAlgorithm: "ecdsa",
		DefaultQuantumAlgorithm: "dilithium",
		EnabledClassicAlgorithms: []string{"ecdsa", "ed25519", "rsa", "schnorr"},
		EnabledQuantumAlgorithms: []string{"dilithium", "falcon", "sphincs+", "xmss", "lms"},
		AlgorithmSelectionStrategy: "hybrid", // classic, quantum, hybrid, migration
	}
	
	config.MigrationSettings.EnableMigration = true
	config.MigrationSettings.MigrationStartDate = "2025-01-01"
	config.MigrationSettings.MigrationEndDate = "2025-12-31"
	config.MigrationSettings.ClassicAlgorithm = "ecdsa"
	config.MigrationSettings.QuantumAlgorithm = "dilithium"
	
	return config
}

func saveConfig(config *AlgorithmConfig, filename string) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling config: %v", err)
	}
	
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Error writing config file: %v", err)
	}
	
	fmt.Printf("✅ Configuration saved to %s\n", filename)
	fmt.Println()
}

func loadConfig(filename string) *AlgorithmConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	
	var config AlgorithmConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}
	
	fmt.Printf("✅ Configuration loaded from %s\n", filename)
	return &config
}

func demonstrateConfigUsage(config *AlgorithmConfig) {
	fmt.Println("=== Configuration Usage Demo ===")
	
	// Создаем менеджеры
	classicManager := crypto.NewSignatureManager()
	quantumManager := crypto.NewQuantumResistantManager()
	
	// Проверяем, что выбранные алгоритмы поддерживаются
	fmt.Printf("Default classic algorithm: %s\n", config.DefaultClassicAlgorithm)
	fmt.Printf("Default quantum algorithm: %s\n", config.DefaultQuantumAlgorithm)
	fmt.Printf("Strategy: %s\n", config.AlgorithmSelectionStrategy)
	fmt.Println()
	
	// Демонстрируем генерацию ключей с выбранными алгоритмами
	testData := []byte("Configuration test data")
	
	// Классический алгоритм
	fmt.Printf("Generating key with %s...\n", config.DefaultClassicAlgorithm)
	classicKeyPair, err := classicManager.GenerateKeyPair(crypto.SignatureAlgorithm(config.DefaultClassicAlgorithm))
	if err != nil {
		log.Printf("Error generating classic key: %v", err)
	} else {
		fmt.Printf("Classic key address: %s\n", classicKeyPair.Address)
		
		// Подписываем и проверяем
		signature, err := classicManager.Sign(crypto.SignatureAlgorithm(config.DefaultClassicAlgorithm), classicKeyPair.PrivateKey, testData)
		if err == nil {
			valid := classicManager.Verify(signature, testData)
			if valid {
				fmt.Printf("✅ Classic signature verified\n")
			}
		}
	}
	
	// Квантово-устойчивый алгоритм
	fmt.Printf("Generating key with %s...\n", config.DefaultQuantumAlgorithm)
	quantumKeyPair, err := quantumManager.GenerateKeyPair(crypto.QuantumResistantAlgorithm(config.DefaultQuantumAlgorithm))
	if err != nil {
		log.Printf("Error generating quantum key: %v", err)
	} else {
		fmt.Printf("Quantum key address: %s\n", quantumKeyPair.Address)
		
		// Подписываем и проверяем
		signature, err := quantumManager.Sign(crypto.QuantumResistantAlgorithm(config.DefaultQuantumAlgorithm), quantumKeyPair.PrivateKey, testData)
		if err == nil {
			valid, err := quantumManager.Verify(crypto.QuantumResistantAlgorithm(config.DefaultQuantumAlgorithm), quantumKeyPair.PublicKey, signature, testData)
			if err == nil && valid {
				fmt.Printf("✅ Quantum signature verified\n")
			}
		}
	}
	
	fmt.Println()
}

func showAlgorithmStrategies() {
	fmt.Println("=== Algorithm Selection Strategies ===")
	
	strategies := map[string]string{
		"classic": "Use only classic algorithms (fast, but vulnerable to quantum attacks)",
		"quantum": "Use only quantum-resistant algorithms (secure, but larger and slower)",
		"hybrid":  "Support both classic and quantum-resistant algorithms",
		"migration": "Gradually migrate from classic to quantum-resistant algorithms",
		"negotiation": "Let clients choose their preferred algorithm",
		"time-based": "Use classic algorithms now, quantum-resistant in the future",
	}
	
	for strategy, description := range strategies {
		fmt.Printf("🔧 %s: %s\n", strategy, description)
	}
	
	fmt.Println()
	
	// Показываем примеры конфигураций для разных стратегий
	fmt.Println("=== Example Configurations ===")
	
	// Стратегия "classic"
	classicConfig := &AlgorithmConfig{
		DefaultClassicAlgorithm: "ed25519",
		EnabledClassicAlgorithms: []string{"ed25519", "ecdsa"},
		AlgorithmSelectionStrategy: "classic",
	}
	fmt.Println("Classic-only configuration:")
	printConfig(classicConfig)
	
	// Стратегия "quantum"
	quantumConfig := &AlgorithmConfig{
		DefaultQuantumAlgorithm: "dilithium",
		EnabledQuantumAlgorithms: []string{"dilithium", "falcon"},
		AlgorithmSelectionStrategy: "quantum",
	}
	fmt.Println("Quantum-only configuration:")
	printConfig(quantumConfig)
	
	// Стратегия "hybrid"
	hybridConfig := &AlgorithmConfig{
		DefaultClassicAlgorithm: "ed25519",
		DefaultQuantumAlgorithm: "dilithium",
		EnabledClassicAlgorithms: []string{"ed25519", "ecdsa"},
		EnabledQuantumAlgorithms: []string{"dilithium", "falcon"},
		AlgorithmSelectionStrategy: "hybrid",
	}
	fmt.Println("Hybrid configuration:")
	printConfig(hybridConfig)
}

func printConfig(config *AlgorithmConfig) {
	fmt.Printf("  Strategy: %s\n", config.AlgorithmSelectionStrategy)
	if config.DefaultClassicAlgorithm != "" {
		fmt.Printf("  Default classic: %s\n", config.DefaultClassicAlgorithm)
	}
	if config.DefaultQuantumAlgorithm != "" {
		fmt.Printf("  Default quantum: %s\n", config.DefaultQuantumAlgorithm)
	}
	if len(config.EnabledClassicAlgorithms) > 0 {
		fmt.Printf("  Enabled classic: %v\n", config.EnabledClassicAlgorithms)
	}
	if len(config.EnabledQuantumAlgorithms) > 0 {
		fmt.Printf("  Enabled quantum: %v\n", config.EnabledQuantumAlgorithms)
	}
	fmt.Println()
}
