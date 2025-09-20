package security

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// AttackProtection представляет систему защиты от атак
type AttackProtection struct {
	blockchain     *blockchain.Blockchain
	hashRateMap    map[string]float64 // nodeID -> hash rate
	blockTimeMap   map[string][]time.Time // nodeID -> block times
	lastBlockTime  time.Time
	alertThreshold float64 // порог для предупреждений
	blockThreshold int     // количество блоков для анализа
	mutex          sync.RWMutex
	alerts         chan SecurityAlert
	ctx            context.Context
	cancel         context.CancelFunc
}

// SecurityAlert представляет предупреждение о безопасности
type SecurityAlert struct {
	Type        AlertType    `json:"type"`
	Severity    Severity     `json:"severity"`
	Message     string       `json:"message"`
	Timestamp   time.Time    `json:"timestamp"`
	Data        interface{}  `json:"data,omitempty"`
	NodeID      string       `json:"node_id,omitempty"`
}

// AlertType определяет тип предупреждения
type AlertType string

const (
	AlertType51Attack    AlertType = "51_attack"
	AlertTypeHashRate    AlertType = "hash_rate_anomaly"
	AlertTypeBlockTime   AlertType = "block_time_anomaly"
	AlertTypeFork        AlertType = "fork_detected"
	AlertTypeSpam        AlertType = "spam_detected"
)

// Severity определяет серьезность предупреждения
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// HashRateData содержит данные о хеш-рейте
type HashRateData struct {
	NodeID    string    `json:"node_id"`
	HashRate  float64   `json:"hash_rate"`
	Timestamp time.Time `json:"timestamp"`
}

// BlockTimeData содержит данные о времени блоков
type BlockTimeData struct {
	NodeID     string    `json:"node_id"`
	BlockTime  time.Duration `json:"block_time"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewAttackProtection создает новую систему защиты от атак
func NewAttackProtection(bc *blockchain.Blockchain) *AttackProtection {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AttackProtection{
		blockchain:     bc,
		hashRateMap:    make(map[string]float64),
		blockTimeMap:   make(map[string][]time.Time),
		alertThreshold: 0.6, // 60% хеш-рейта от одного узла
		blockThreshold: 10,  // анализируем последние 10 блоков
		alerts:         make(chan SecurityAlert, 100),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start запускает систему мониторинга
func (ap *AttackProtection) Start() {
	go ap.monitorHashRate()
	go ap.monitorBlockTimes()
	go ap.monitorForks()
	go ap.processAlerts()
	
	log.Println("Security monitoring started")
}

// Stop останавливает систему мониторинга
func (ap *AttackProtection) Stop() {
	ap.cancel()
	close(ap.alerts)
	log.Println("Security monitoring stopped")
}

// ReportHashRate сообщает о хеш-рейте узла
func (ap *AttackProtection) ReportHashRate(nodeID string, hashRate float64) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	
	ap.hashRateMap[nodeID] = hashRate
	
	// Проверяем на аномалии
	ap.checkHashRateAnomaly(nodeID, hashRate)
}

// ReportBlockTime сообщает о времени создания блока
func (ap *AttackProtection) ReportBlockTime(nodeID string, blockTime time.Duration) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	
	now := time.Now()
	ap.blockTimeMap[nodeID] = append(ap.blockTimeMap[nodeID], now)
	
	// Ограничиваем количество записей
	if len(ap.blockTimeMap[nodeID]) > ap.blockThreshold {
		ap.blockTimeMap[nodeID] = ap.blockTimeMap[nodeID][1:]
	}
	
	ap.lastBlockTime = now
	
	// Проверяем на аномалии
	ap.checkBlockTimeAnomaly(nodeID, blockTime)
}

// checkHashRateAnomaly проверяет аномалии в хеш-рейте
func (ap *AttackProtection) checkHashRateAnomaly(nodeID string, hashRate float64) {
	totalHashRate := ap.calculateTotalHashRate()
	
	if totalHashRate > 0 {
		percentage := hashRate / totalHashRate
		
		if percentage > ap.alertThreshold {
			alert := SecurityAlert{
				Type:      AlertType51Attack,
				Severity:  ap.getSeverity(percentage),
				Message:   fmt.Sprintf("Node %s has %.2f%% of total hash rate", nodeID, percentage*100),
				Timestamp: time.Now(),
				Data: HashRateData{
					NodeID:    nodeID,
					HashRate:  hashRate,
					Timestamp: time.Now(),
				},
				NodeID: nodeID,
			}
			
			select {
			case ap.alerts <- alert:
			default:
				log.Printf("Alert channel full, dropping alert: %v", alert)
			}
		}
	}
}

// checkBlockTimeAnomaly проверяет аномалии во времени блоков
func (ap *AttackProtection) checkBlockTimeAnomaly(nodeID string, blockTime time.Duration) {
	times := ap.blockTimeMap[nodeID]
	if len(times) < 3 {
		return
	}
	
	// Вычисляем среднее время между блоками для этого узла
	var totalDuration time.Duration
	for i := 1; i < len(times); i++ {
		totalDuration += times[i].Sub(times[i-1])
	}
	avgDuration := totalDuration / time.Duration(len(times)-1)
	
	// Проверяем, не слишком ли быстро создаются блоки
	expectedBlockTime := 10 * time.Second // ожидаемое время блока
	if avgDuration < expectedBlockTime/2 {
		alert := SecurityAlert{
			Type:      AlertTypeBlockTime,
			Severity:  SeverityMedium,
			Message:   fmt.Sprintf("Node %s creating blocks too fast: avg %.2fs", nodeID, avgDuration.Seconds()),
			Timestamp: time.Now(),
			Data: BlockTimeData{
				NodeID:    nodeID,
				BlockTime: avgDuration,
				Timestamp: time.Now(),
			},
			NodeID: nodeID,
		}
		
		select {
		case ap.alerts <- alert:
		default:
			log.Printf("Alert channel full, dropping alert: %v", alert)
		}
	}
}

// calculateTotalHashRate вычисляет общий хеш-рейт сети
func (ap *AttackProtection) calculateTotalHashRate() float64 {
	var total float64
	for _, rate := range ap.hashRateMap {
		total += rate
	}
	return total
}

// getSeverity определяет серьезность на основе процента хеш-рейта
func (ap *AttackProtection) getSeverity(percentage float64) Severity {
	switch {
	case percentage >= 0.8:
		return SeverityCritical
	case percentage >= 0.6:
		return SeverityHigh
	case percentage >= 0.4:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// monitorHashRate мониторит хеш-рейт в фоне
func (ap *AttackProtection) monitorHashRate() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ap.ctx.Done():
			return
		case <-ticker.C:
			ap.analyzeHashRateDistribution()
		}
	}
}

// monitorBlockTimes мониторит время блоков в фоне
func (ap *AttackProtection) monitorBlockTimes() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ap.ctx.Done():
			return
		case <-ticker.C:
			ap.analyzeBlockTimeDistribution()
		}
	}
}

// monitorForks мониторит форки в фоне
func (ap *AttackProtection) monitorForks() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ap.ctx.Done():
			return
		case <-ticker.C:
			ap.checkForForks()
		}
	}
}

// processAlerts обрабатывает предупреждения
func (ap *AttackProtection) processAlerts() {
	for {
		select {
		case <-ap.ctx.Done():
			return
		case alert := <-ap.alerts:
			ap.handleAlert(alert)
		}
	}
}

// analyzeHashRateDistribution анализирует распределение хеш-рейта
func (ap *AttackProtection) analyzeHashRateDistribution() {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	totalHashRate := ap.calculateTotalHashRate()
	if totalHashRate == 0 {
		return
	}
	
	// Проверяем концентрацию хеш-рейта
	var topNodes []string
	var topHashRate float64
	
	for nodeID, rate := range ap.hashRateMap {
		percentage := rate / totalHashRate
		if percentage > 0.3 { // более 30%
			topNodes = append(topNodes, nodeID)
			topHashRate += rate
		}
	}
	
	if len(topNodes) > 0 && topHashRate/totalHashRate > 0.7 {
		alert := SecurityAlert{
			Type:      AlertTypeHashRate,
			Severity:  SeverityHigh,
			Message:   fmt.Sprintf("High hash rate concentration: %d nodes control %.2f%%", len(topNodes), (topHashRate/totalHashRate)*100),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"top_nodes":     topNodes,
				"concentration": topHashRate / totalHashRate,
			},
		}
		
		select {
		case ap.alerts <- alert:
		default:
		}
	}
}

// analyzeBlockTimeDistribution анализирует распределение времени блоков
func (ap *AttackProtection) analyzeBlockTimeDistribution() {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	// Анализируем аномалии во времени блоков
	for nodeID, times := range ap.blockTimeMap {
		if len(times) < 5 {
			continue
		}
		
		// Вычисляем стандартное отклонение
		var total time.Duration
		for i := 1; i < len(times); i++ {
			total += times[i].Sub(times[i-1])
		}
		avg := total / time.Duration(len(times)-1)
		
		var variance time.Duration
		for i := 1; i < len(times); i++ {
			diff := times[i].Sub(times[i-1]) - avg
			variance += diff * diff
		}
		stdDev := time.Duration(variance / time.Duration(len(times)-1))
		
		// Если стандартное отклонение слишком большое, это подозрительно
		if stdDev > avg {
			alert := SecurityAlert{
				Type:      AlertTypeBlockTime,
				Severity:  SeverityMedium,
				Message:   fmt.Sprintf("Node %s has irregular block timing", nodeID),
				Timestamp: time.Now(),
				NodeID:    nodeID,
			}
			
			select {
			case ap.alerts <- alert:
			default:
			}
		}
	}
}

// checkForForks проверяет наличие форков
func (ap *AttackProtection) checkForForks() {
	// Здесь можно добавить логику проверки форков
	// Например, сравнивать с другими узлами в сети
}

// handleAlert обрабатывает предупреждение
func (ap *AttackProtection) handleAlert(alert SecurityAlert) {
	log.Printf("SECURITY ALERT [%s] %s: %s", alert.Severity, alert.Type, alert.Message)
	
	// Здесь можно добавить дополнительные действия:
	// - Отправка уведомлений
	// - Блокировка подозрительных узлов
	// - Уведомление администраторов
}

// GetStats возвращает статистику безопасности
func (ap *AttackProtection) GetStats() map[string]interface{} {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()
	
	totalHashRate := ap.calculateTotalHashRate()
	
	stats := map[string]interface{}{
		"total_hash_rate": totalHashRate,
		"active_nodes":    len(ap.hashRateMap),
		"last_block_time": ap.lastBlockTime,
		"alert_threshold": ap.alertThreshold,
	}
	
	// Добавляем топ узлы по хеш-рейту
	var topNodes []map[string]interface{}
	for nodeID, rate := range ap.hashRateMap {
		percentage := float64(0)
		if totalHashRate > 0 {
			percentage = rate / totalHashRate
		}
		
		topNodes = append(topNodes, map[string]interface{}{
			"node_id":    nodeID,
			"hash_rate":  rate,
			"percentage": percentage,
		})
	}
	
	stats["top_nodes"] = topNodes
	return stats
}

// GetAlerts возвращает канал предупреждений
func (ap *AttackProtection) GetAlerts() <-chan SecurityAlert {
	return ap.alerts
}
