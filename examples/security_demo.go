//go:build security_demo
// +build security_demo

package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/consensus"
	"mirochain/internal/crypto"
	"mirochain/internal/persistent"
	"mirochain/internal/security"
)

func main() {
	fmt.Println("=== MiroChain Security Demo ===")

	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain(
		"data/security_demo",
		"security_demo",
		[]byte("security_demo_public_key"),
		4,
	)
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}
	defer bc.Close()

	// 1. Тестируем защиту от атак 51%
	fmt.Println("\n1. Testing 51% Attack Protection:")
	fmt.Println("==================================")
	
	attackProtection := security.NewAttackProtection(&blockchain.Blockchain{})
	attackProtection.Start()
	defer attackProtection.Stop()
	
	// Симулируем хеш-рейт разных узлов
	nodes := []string{"node1", "node2", "node3", "node4", "node5"}
	hashRates := []float64{1000, 800, 600, 400, 200} // node1 имеет 33% хеш-рейта
	
	for i, node := range nodes {
		attackProtection.ReportHashRate(node, hashRates[i])
		fmt.Printf("Node %s: %.0f H/s\n", node, hashRates[i])
	}
	
	// Проверяем статистику
	stats := attackProtection.GetStats()
	fmt.Printf("Total Hash Rate: %.0f H/s\n", stats["total_hash_rate"])
	fmt.Printf("Active Nodes: %v\n", stats["active_nodes"])
	
	// 2. Тестируем валидацию входных данных
	fmt.Println("\n2. Testing Input Validation:")
	fmt.Println("============================")
	
	validator := security.NewInputValidator()
	
	// Тестируем валидацию блока
	block := &blockchain.Block{
		Height:       1,
		PreviousHash: []byte("0000000000000000000000000000000000000000000000000000000000000000"),
		Timestamp:    time.Now().Unix(),
		Nonce:        12345,
		Difficulty:   4,
		Transactions: []*blockchain.Transaction{},
	}
	
	blockResult := validator.ValidateBlock(block)
	fmt.Printf("Block validation: Valid=%t, Errors=%d\n", blockResult.Valid, len(blockResult.Errors))
	
	// Тестируем валидацию транзакции
	tx := &blockchain.Transaction{
		ID:    []byte("test_transaction_id"),
		Inputs: []*blockchain.TransactionInput{},
		Outputs: []*blockchain.TransactionOutput{
			{
				Address: "test_address",
				Value:   100,
			},
		},
		Timestamp: time.Now().Unix(),
		Fee:       1,
	}
	
	txResult := validator.ValidateTransaction(tx)
	fmt.Printf("Transaction validation: Valid=%t, Errors=%d\n", txResult.Valid, len(txResult.Errors))
	
	// Тестируем валидацию адреса
	addrResult := validator.ValidateAddress("test_address_123")
	fmt.Printf("Address validation: Valid=%t, Errors=%d\n", addrResult.Valid, len(addrResult.Errors))
	
	// 3. Тестируем улучшенный Rate Limiting
	fmt.Println("\n3. Testing Enhanced Rate Limiting:")
	fmt.Println("==================================")
	
	rateLimiter := security.NewAPIRateLimiter()
	
	// Тестируем разные типы запросов
	requestTypes := []string{"api", "mining", "wallet", "admin"}
	
	for _, reqType := range requestTypes {
		fmt.Printf("\nTesting %s rate limiting:\n", reqType)
		
		allowed := 0
		blocked := 0
		
		for i := 0; i < 10; i++ {
			result := rateLimiter.CheckRateLimit(reqType, "test_client")
			if result.Allowed {
				allowed++
			} else {
				blocked++
			}
		}
		
		fmt.Printf("  Allowed: %d, Blocked: %d\n", allowed, blocked)
	}
	
	// 4. Тестируем Proof of Stake
	fmt.Println("\n4. Testing Proof of Stake:")
	fmt.Println("=========================")
	
	pos := consensus.NewProofOfStake(&blockchain.Blockchain{})
	
	// Добавляем стейки
	stakers := []string{"staker1", "staker2", "staker3", "staker4", "staker5"}
	amounts := []*big.Int{
		big.NewInt(1000),
		big.NewInt(2000),
		big.NewInt(1500),
		big.NewInt(3000),
		big.NewInt(500),
	}
	
	for i, staker := range stakers {
		err := pos.Stake(staker, amounts[i], time.Now().Unix()+3600)
		if err != nil {
			log.Printf("Error staking for %s: %v", staker, err)
		} else {
			fmt.Printf("Staked %s tokens for %s\n", amounts[i].String(), staker)
		}
	}
	
	// Выбираем валидатора
	posValidator, err := pos.SelectValidator(1)
	if err != nil {
		log.Printf("Error selecting validator: %v", err)
	} else {
		fmt.Printf("Selected validator: %s\n", posValidator.Address)
	}
	
	// Получаем статистику PoS
	posStats := pos.GetStats()
	fmt.Printf("Total Stake: %s\n", posStats["total_stake"])
	fmt.Printf("Active Stakes: %v\n", posStats["active_stakes"])
	fmt.Printf("Validators: %v\n", posStats["validators"])
	
	// 5. Тестируем Delegated Proof of Stake
	fmt.Println("\n5. Testing Delegated Proof of Stake:")
	fmt.Println("====================================")
	
	dpos := consensus.NewDelegatedProofOfStake(&blockchain.Blockchain{})
	
	// Регистрируем делегатов
	delegates := []string{"delegate1", "delegate2", "delegate3", "delegate4", "delegate5"}
	for _, delegate := range delegates {
		err := dpos.RegisterDelegate(delegate)
		if err != nil {
			log.Printf("Error registering delegate %s: %v", delegate, err)
		} else {
			fmt.Printf("Registered delegate: %s\n", delegate)
		}
	}
	
	// Создаем голоса
	voters := []string{"voter1", "voter2", "voter3", "voter4", "voter5"}
	votePowers := []*big.Int{
		big.NewInt(100),
		big.NewInt(200),
		big.NewInt(150),
		big.NewInt(300),
		big.NewInt(50),
	}
	
	for i, voter := range voters {
		delegate := delegates[i%len(delegates)]
		err := dpos.Vote(voter, delegate, votePowers[i])
		if err != nil {
			log.Printf("Error voting from %s to %s: %v", voter, delegate, err)
		} else {
			fmt.Printf("Voted %s for %s\n", votePowers[i].String(), delegate)
		}
	}
	
	// Выбираем делегата
	delegate, err := dpos.SelectDelegate(1)
	if err != nil {
		log.Printf("Error selecting delegate: %v", err)
	} else {
		fmt.Printf("Selected delegate: %s\n", delegate.Address)
	}
	
	// Получаем статистику DPoS
	dposStats := dpos.GetStats()
	fmt.Printf("Total Votes: %s\n", dposStats["total_votes"])
	fmt.Printf("Total Delegates: %v\n", dposStats["total_delegates"])
	fmt.Printf("Active Delegates: %v\n", dposStats["active_delegates"])
	
	// 6. Тестируем сравнение алгоритмов консенсуса
	fmt.Println("\n6. Testing Consensus Algorithm Comparison:")
	fmt.Println("==========================================")
	
	comparison := consensus.NewConsensusComparison(&blockchain.Blockchain{})
	
	// Запускаем сравнение
	results := comparison.RunComparison()
	
	fmt.Println("\nComparison Results:")
	for algorithm, metrics := range results {
		fmt.Printf("\n%s:\n", algorithm)
		fmt.Printf("  Block Time: %v\n", metrics.BlockTime)
		fmt.Printf("  Throughput: %.2f TPS\n", metrics.Throughput)
		fmt.Printf("  Energy Usage: %.2f units\n", metrics.EnergyUsage)
		fmt.Printf("  Security: %.2f\n", metrics.Security)
		fmt.Printf("  Decentralization: %.2f\n", metrics.Decentralization)
		fmt.Printf("  Scalability: %.2f\n", metrics.Scalability)
	}
	
	// 7. Тестируем разные алгоритмы подписи
	fmt.Println("\n7. Testing Signature Algorithms:")
	fmt.Println("===============================")
	
	signatureManager := crypto.NewSignatureManager()
	
	algorithms := []crypto.SignatureAlgorithm{
		crypto.AlgorithmECDSA,
		crypto.AlgorithmEd25519,
		crypto.AlgorithmRSA,
		crypto.AlgorithmSchnorr,
	}
	
	for _, algorithm := range algorithms {
		fmt.Printf("\nTesting %s:\n", algorithm)
		
		// Генерируем пару ключей
		keyPair, err := signatureManager.GenerateKeyPair(algorithm)
		if err != nil {
			log.Printf("Error generating key pair for %s: %v", algorithm, err)
			continue
		}
		
		fmt.Printf("  Address: %s\n", keyPair.Address)
		
		// Подписываем данные
		data := []byte("test data for signing")
		signature, err := signatureManager.Sign(algorithm, keyPair.PrivateKey, data)
		if err != nil {
			log.Printf("Error signing with %s: %v", algorithm, err)
			continue
		}
		
		// Проверяем подпись
		valid := signatureManager.Verify(signature, data)
		fmt.Printf("  Signature valid: %t\n", valid)
	}
	
	// 8. Тестируем мультиподписи
	fmt.Println("\n8. Testing Multi-Signatures:")
	fmt.Println("============================")
	
	multisigManager := crypto.NewMultiSigManager()
	
	// Генерируем ключи для мультиподписи
	keyPairs := make([]*crypto.SignatureKeyPair, 3)
	for i := 0; i < 3; i++ {
		keyPair, err := signatureManager.GenerateKeyPair(crypto.AlgorithmECDSA)
		if err != nil {
			log.Printf("Error generating key pair %d: %v", i, err)
			continue
		}
		keyPairs[i] = keyPair
	}
	
	// Создаем мультиподпись
	publicKeys := make([][]byte, len(keyPairs))
	for i, keyPair := range keyPairs {
		publicKeys[i] = keyPair.PublicKey
	}
	
	multisig, err := multisigManager.CreateMultiSig(publicKeys, 2, crypto.AlgorithmECDSA)
	if err != nil {
		log.Printf("Error creating multisig: %v", err)
	} else {
		fmt.Printf("Created multisig: %s\n", multisig.Address)
		fmt.Printf("Threshold: %d/%d\n", multisig.Threshold, len(multisig.PublicKeys))
	}
	
	// Создаем транзакцию с мультиподписью
	transaction, err := multisigManager.CreateTransaction(
		multisig.Address,
		"recipient_address",
		big.NewInt(1000),
		[]byte("multisig transaction data"),
	)
	if err != nil {
		log.Printf("Error creating multisig transaction: %v", err)
	} else {
		fmt.Printf("Created transaction: %s\n", transaction.ID)
		fmt.Printf("Status: %s\n", transaction.Status)
	}
	
	// Подписываем транзакцию
	for i, keyPair := range keyPairs[:2] { // Подписываем первыми двумя ключами
		err := multisigManager.SignTransaction(
			transaction.ID,
			keyPair.Address,
			keyPair.PrivateKey,
		)
		if err != nil {
			log.Printf("Error signing transaction with key %d: %v", i, err)
		} else {
			fmt.Printf("Signed transaction with key %d\n", i)
		}
	}
	
	// Проверяем транзакцию
	valid, err := multisigManager.VerifyTransaction(transaction.ID)
	if err != nil {
		log.Printf("Error verifying transaction: %v", err)
	} else {
		fmt.Printf("Transaction valid: %t\n", valid)
	}
	
	// Получаем статистику мультиподписей
	multisigStats := multisigManager.GetStats()
	fmt.Printf("\nMultiSig Stats:\n")
	fmt.Printf("  Total MultiSigs: %v\n", multisigStats["total_multisigs"])
	fmt.Printf("  Total Transactions: %v\n", multisigStats["total_transactions"])
	fmt.Printf("  Pending Transactions: %v\n", multisigStats["pending_transactions"])
	fmt.Printf("  Signed Transactions: %v\n", multisigStats["signed_transactions"])
	
	fmt.Println("\nSecurity Demo completed!")
}
