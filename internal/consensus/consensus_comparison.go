package consensus

import (
	"log"
	"math/big"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// ConsensusComparison представляет систему сравнения алгоритмов консенсуса
type ConsensusComparison struct {
	blockchain *blockchain.Blockchain
	pos        *ProofOfStake
	dpos       *DelegatedProofOfStake
	metrics    *ConsensusMetrics
	mutex      sync.RWMutex
}

// ConsensusMetrics содержит метрики для сравнения
type ConsensusMetrics struct {
	Algorithm        string        `json:"algorithm"`
	BlockTime        time.Duration `json:"block_time"`
	Throughput       float64       `json:"throughput"`       // транзакций в секунду
	EnergyUsage      float64       `json:"energy_usage"`     // условные единицы энергии
	Security         float64       `json:"security"`         // оценка безопасности (0-1)
	Decentralization float64       `json:"decentralization"` // оценка децентрализации (0-1)
	Finality         time.Duration `json:"finality"`         // время финальности
	Cost             float64       `json:"cost"`             // стоимость участия
	Scalability      float64       `json:"scalability"`      // оценка масштабируемости (0-1)
}

// ConsensusTest представляет тест консенсуса
type ConsensusTest struct {
	Algorithm        string        `json:"algorithm"`
	Duration         time.Duration `json:"duration"`
	BlockCount       int           `json:"block_count"`
	TxCount          int           `json:"tx_count"`
	EnergyUsed       float64       `json:"energy_used"`
	Participants     int           `json:"participants"`
	SuccessRate      float64       `json:"success_rate"`
	AverageBlockTime time.Duration `json:"average_block_time"`
}

// NewConsensusComparison создает новую систему сравнения консенсуса
func NewConsensusComparison(bc *blockchain.Blockchain) *ConsensusComparison {
	return &ConsensusComparison{
		blockchain: bc,
		pos:        NewProofOfStake(bc),
		dpos:       NewDelegatedProofOfStake(bc),
		metrics:    &ConsensusMetrics{},
	}
}

// RunComparison запускает сравнение алгоритмов консенсуса
func (cc *ConsensusComparison) RunComparison() map[string]*ConsensusMetrics {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	results := make(map[string]*ConsensusMetrics)

	// Тестируем PoW
	log.Println("Testing Proof of Work...")
	powMetrics := cc.testProofOfWork()
	results["pow"] = powMetrics

	// Тестируем PoS
	log.Println("Testing Proof of Stake...")
	posMetrics := cc.testProofOfStake()
	results["pos"] = posMetrics

	// Тестируем DPoS
	log.Println("Testing Delegated Proof of Stake...")
	dposMetrics := cc.testDelegatedProofOfStake()
	results["dpos"] = dposMetrics

	// Анализируем результаты
	cc.analyzeResults(results)

	return results
}

// testProofOfWork тестирует Proof of Work
func (cc *ConsensusComparison) testProofOfWork() *ConsensusMetrics {
	startTime := time.Now()

	// Симулируем майнинг блоков
	blockCount := 10
	txCount := 0
	energyUsed := 0.0

	for i := 0; i < blockCount; i++ {
		// Симулируем создание блока
		_ = &blockchain.Block{
			Height:     int64(i + 1),
			Timestamp:  time.Now().Unix(),
			Difficulty: 4,
			Nonce:      0,
		}

		// Симулируем транзакции
		txCount += 5 + i%10

		// Симулируем энергопотребление
		energyUsed += float64(4) * 1000 // Сложность 4

		// Симулируем время майнинга
		time.Sleep(time.Millisecond * 100)
	}

	duration := time.Since(startTime)

	return &ConsensusMetrics{
		Algorithm:        "Proof of Work",
		BlockTime:        duration / time.Duration(blockCount),
		Throughput:       float64(txCount) / duration.Seconds(),
		EnergyUsage:      energyUsed,
		Security:         0.9,              // Высокая безопасность
		Decentralization: 0.8,              // Хорошая децентрализация
		Finality:         time.Minute * 6,  // 6 подтверждений
		Cost:             energyUsed * 0.1, // Стоимость энергии
		Scalability:      0.3,              // Низкая масштабируемость
	}
}

// testProofOfStake тестирует Proof of Stake
func (cc *ConsensusComparison) testProofOfStake() *ConsensusMetrics {
	startTime := time.Now()

	// Настраиваем стейки
	cc.setupStakes()

	// Симулируем создание блоков
	blockCount := 10
	txCount := 0
	energyUsed := 0.0

	for i := 0; i < blockCount; i++ {
		// Выбираем валидатора
		_, err := cc.pos.SelectValidator(int64(i + 1))
		if err != nil {
			log.Printf("Error selecting validator: %v", err)
			continue
		}

		// Симулируем создание блока
		_ = &blockchain.Block{
			Height:    int64(i + 1),
			Timestamp: time.Now().Unix(),
		}

		// Симулируем транзакции
		txCount += 10 + i%20

		// Симулируем энергопотребление (намного меньше чем PoW)
		energyUsed += 10

		// Симулируем время создания блока
		time.Sleep(time.Millisecond * 50)
	}

	duration := time.Since(startTime)

	return &ConsensusMetrics{
		Algorithm:        "Proof of Stake",
		BlockTime:        duration / time.Duration(blockCount),
		Throughput:       float64(txCount) / duration.Seconds(),
		EnergyUsage:      energyUsed,
		Security:         0.8,               // Хорошая безопасность
		Decentralization: 0.7,               // Средняя децентрализация
		Finality:         time.Minute * 2,   // 2 подтверждения
		Cost:             energyUsed * 0.01, // Низкая стоимость
		Scalability:      0.7,               // Хорошая масштабируемость
	}
}

// testDelegatedProofOfStake тестирует Delegated Proof of Stake
func (cc *ConsensusComparison) testDelegatedProofOfStake() *ConsensusMetrics {
	startTime := time.Now()

	// Настраиваем делегатов и голоса
	cc.setupDelegates()

	// Симулируем создание блоков
	blockCount := 10
	txCount := 0
	energyUsed := 0.0

	for i := 0; i < blockCount; i++ {
		// Выбираем делегата
		_, err := cc.dpos.SelectDelegate(int64(i + 1))
		if err != nil {
			log.Printf("Error selecting delegate: %v", err)
			continue
		}

		// Симулируем создание блока
		_ = &blockchain.Block{
			Height:    int64(i + 1),
			Timestamp: time.Now().Unix(),
		}

		// Симулируем транзакции
		txCount += 15 + i%25

		// Симулируем энергопотребление (очень низкое)
		energyUsed += 5

		// Симулируем время создания блока
		time.Sleep(time.Millisecond * 30)
	}

	duration := time.Since(startTime)

	return &ConsensusMetrics{
		Algorithm:        "Delegated Proof of Stake",
		BlockTime:        duration / time.Duration(blockCount),
		Throughput:       float64(txCount) / duration.Seconds(),
		EnergyUsage:      energyUsed,
		Security:         0.7,                // Средняя безопасность
		Decentralization: 0.5,                // Низкая децентрализация
		Finality:         time.Second * 30,   // 30 секунд
		Cost:             energyUsed * 0.005, // Очень низкая стоимость
		Scalability:      0.9,                // Высокая масштабируемость
	}
}

// setupStakes настраивает стейки для тестирования
func (cc *ConsensusComparison) setupStakes() {
	// Создаем несколько стейков
	addresses := []string{
		"staker_1", "staker_2", "staker_3", "staker_4", "staker_5",
	}

	for i, address := range addresses {
		amount := big.NewInt(int64(1000 + i*500))
		lockTime := time.Now().Unix() + 3600 // 1 час

		err := cc.pos.Stake(address, amount, lockTime)
		if err != nil {
			log.Printf("Error setting up stake for %s: %v", address, err)
		}
	}
}

// setupDelegates настраивает делегатов для тестирования
func (cc *ConsensusComparison) setupDelegates() {
	// Регистрируем делегатов
	delegates := []string{
		"delegate_1", "delegate_2", "delegate_3", "delegate_4", "delegate_5",
	}

	for _, delegate := range delegates {
		err := cc.dpos.RegisterDelegate(delegate)
		if err != nil {
			log.Printf("Error registering delegate %s: %v", delegate, err)
		}
	}

	// Создаем голоса
	voters := []string{
		"voter_1", "voter_2", "voter_3", "voter_4", "voter_5",
	}

	for i, voter := range voters {
		delegate := delegates[i%len(delegates)]
		votePower := big.NewInt(int64(100 + i*50))

		err := cc.dpos.Vote(voter, delegate, votePower)
		if err != nil {
			log.Printf("Error casting vote from %s to %s: %v", voter, delegate, err)
		}
	}
}

// analyzeResults анализирует результаты сравнения
func (cc *ConsensusComparison) analyzeResults(results map[string]*ConsensusMetrics) {
	log.Println("\n=== Consensus Algorithm Comparison Results ===")

	// Находим лучшие алгоритмы по разным критериям
	bestThroughput := cc.findBestAlgorithm(results, "throughput")
	bestEnergy := cc.findBestAlgorithm(results, "energy")
	bestSecurity := cc.findBestAlgorithm(results, "security")
	bestDecentralization := cc.findBestAlgorithm(results, "decentralization")
	bestScalability := cc.findBestAlgorithm(results, "scalability")

	log.Printf("Best Throughput: %s (%.2f TPS)", bestThroughput, results[bestThroughput].Throughput)
	log.Printf("Best Energy Efficiency: %s (%.2f units)", bestEnergy, results[bestEnergy].EnergyUsage)
	log.Printf("Best Security: %s (%.2f)", bestSecurity, results[bestSecurity].Security)
	log.Printf("Best Decentralization: %s (%.2f)", bestDecentralization, results[bestDecentralization].Decentralization)
	log.Printf("Best Scalability: %s (%.2f)", bestScalability, results[bestScalability].Scalability)

	// Вычисляем общий рейтинг
	cc.calculateOverallRating(results)
}

// findBestAlgorithm находит лучший алгоритм по критерию
func (cc *ConsensusComparison) findBestAlgorithm(results map[string]*ConsensusMetrics, criterion string) string {
	bestAlgorithm := ""
	bestValue := 0.0

	for algorithm, metrics := range results {
		var value float64
		switch criterion {
		case "throughput":
			value = metrics.Throughput
		case "energy":
			value = 1.0 / metrics.EnergyUsage // Инвертируем для поиска минимума
		case "security":
			value = metrics.Security
		case "decentralization":
			value = metrics.Decentralization
		case "scalability":
			value = metrics.Scalability
		default:
			continue
		}

		if value > bestValue {
			bestValue = value
			bestAlgorithm = algorithm
		}
	}

	return bestAlgorithm
}

// calculateOverallRating вычисляет общий рейтинг алгоритмов
func (cc *ConsensusComparison) calculateOverallRating(results map[string]*ConsensusMetrics) {
	log.Println("\n=== Overall Rating ===")

	// Веса для разных критериев
	weights := map[string]float64{
		"throughput":       0.2,
		"energy":           0.15,
		"security":         0.25,
		"decentralization": 0.2,
		"scalability":      0.2,
	}

	ratings := make(map[string]float64)

	for algorithm, metrics := range results {
		rating := 0.0

		// Нормализуем значения (0-1)
		throughput := cc.normalize(metrics.Throughput, 0, 1000)
		energy := cc.normalize(1.0/metrics.EnergyUsage, 0, 1)
		security := metrics.Security
		decentralization := metrics.Decentralization
		scalability := metrics.Scalability

		// Вычисляем взвешенную сумму
		rating += throughput * weights["throughput"]
		rating += energy * weights["energy"]
		rating += security * weights["security"]
		rating += decentralization * weights["decentralization"]
		rating += scalability * weights["scalability"]

		ratings[algorithm] = rating
	}

	// Сортируем по рейтингу
	sortedAlgorithms := cc.sortByRating(ratings)

	log.Println("Overall Ranking:")
	for i, algorithm := range sortedAlgorithms {
		log.Printf("%d. %s: %.3f", i+1, algorithm, ratings[algorithm])
	}
}

// normalize нормализует значение в диапазоне [0, 1]
func (cc *ConsensusComparison) normalize(value, min, max float64) float64 {
	if max == min {
		return 0.5
	}

	normalized := (value - min) / (max - min)
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

// sortByRating сортирует алгоритмы по рейтингу
func (cc *ConsensusComparison) sortByRating(ratings map[string]float64) []string {
	algorithms := make([]string, 0, len(ratings))
	for algorithm := range ratings {
		algorithms = append(algorithms, algorithm)
	}

	// Простая сортировка пузырьком
	for i := 0; i < len(algorithms)-1; i++ {
		for j := 0; j < len(algorithms)-i-1; j++ {
			if ratings[algorithms[j]] < ratings[algorithms[j+1]] {
				algorithms[j], algorithms[j+1] = algorithms[j+1], algorithms[j]
			}
		}
	}

	return algorithms
}

// GetDetailedComparison возвращает детальное сравнение
func (cc *ConsensusComparison) GetDetailedComparison() map[string]interface{} {
	results := cc.RunComparison()

	comparison := map[string]interface{}{
		"timestamp":  time.Now().Unix(),
		"algorithms": results,
		"summary": map[string]interface{}{
			"best_throughput":       cc.findBestAlgorithm(results, "throughput"),
			"best_energy":           cc.findBestAlgorithm(results, "energy"),
			"best_security":         cc.findBestAlgorithm(results, "security"),
			"best_decentralization": cc.findBestAlgorithm(results, "decentralization"),
			"best_scalability":      cc.findBestAlgorithm(results, "scalability"),
		},
	}

	return comparison
}

// RunStressTest запускает стресс-тест для алгоритмов консенсуса
func (cc *ConsensusComparison) RunStressTest(duration time.Duration) map[string]*ConsensusTest {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	tests := make(map[string]*ConsensusTest)

	// Стресс-тест для PoW
	log.Println("Running stress test for Proof of Work...")
	powTest := cc.stressTestPoW(duration)
	tests["pow"] = powTest

	// Стресс-тест для PoS
	log.Println("Running stress test for Proof of Stake...")
	posTest := cc.stressTestPoS(duration)
	tests["pos"] = posTest

	// Стресс-тест для DPoS
	log.Println("Running stress test for Delegated Proof of Stake...")
	dposTest := cc.stressTestDPoS(duration)
	tests["dpos"] = dposTest

	return tests
}

// stressTestPoW запускает стресс-тест для PoW
func (cc *ConsensusComparison) stressTestPoW(duration time.Duration) *ConsensusTest {
	startTime := time.Now()
	blockCount := 0
	txCount := 0
	energyUsed := 0.0

	for time.Since(startTime) < duration {
		blockCount++
		txCount += 5 + blockCount%10
		energyUsed += float64(4) * 1000 // Сложность 4
		time.Sleep(time.Millisecond * 100)
	}

	actualDuration := time.Since(startTime)

	return &ConsensusTest{
		Algorithm:        "Proof of Work",
		Duration:         actualDuration,
		BlockCount:       blockCount,
		TxCount:          txCount,
		EnergyUsed:       energyUsed,
		Participants:     5,
		SuccessRate:      0.95,
		AverageBlockTime: actualDuration / time.Duration(blockCount),
	}
}

// stressTestPoS запускает стресс-тест для PoS
func (cc *ConsensusComparison) stressTestPoS(duration time.Duration) *ConsensusTest {
	startTime := time.Now()
	blockCount := 0
	txCount := 0
	energyUsed := 0.0

	// Настраиваем стейки
	cc.setupStakes()

	for time.Since(startTime) < duration {
		blockCount++
		txCount += 10 + blockCount%20
		energyUsed += 10
		time.Sleep(time.Millisecond * 50)
	}

	actualDuration := time.Since(startTime)

	return &ConsensusTest{
		Algorithm:        "Proof of Stake",
		Duration:         actualDuration,
		BlockCount:       blockCount,
		TxCount:          txCount,
		EnergyUsed:       energyUsed,
		Participants:     5,
		SuccessRate:      0.98,
		AverageBlockTime: actualDuration / time.Duration(blockCount),
	}
}

// stressTestDPoS запускает стресс-тест для DPoS
func (cc *ConsensusComparison) stressTestDPoS(duration time.Duration) *ConsensusTest {
	startTime := time.Now()
	blockCount := 0
	txCount := 0
	energyUsed := 0.0

	// Настраиваем делегатов
	cc.setupDelegates()

	for time.Since(startTime) < duration {
		blockCount++
		txCount += 15 + blockCount%25
		energyUsed += 5
		time.Sleep(time.Millisecond * 30)
	}

	actualDuration := time.Since(startTime)

	return &ConsensusTest{
		Algorithm:        "Delegated Proof of Stake",
		Duration:         actualDuration,
		BlockCount:       blockCount,
		TxCount:          txCount,
		EnergyUsed:       energyUsed,
		Participants:     5,
		SuccessRate:      0.99,
		AverageBlockTime: actualDuration / time.Duration(blockCount),
	}
}
