package network

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// GossipMessageType определяет тип сообщения в gossip протоколе
type GossipMessageType string

const (
	GossipBlock       GossipMessageType = "block"
	GossipTransaction GossipMessageType = "transaction"
	GossipRequest     GossipMessageType = "request"
	GossipResponse    GossipMessageType = "response"
	GossipHeartbeat   GossipMessageType = "heartbeat"
)

// GossipMessage представляет сообщение в gossip протоколе
type GossipMessage struct {
	Type      GossipMessageType `json:"type"`
	Data      interface{}       `json:"data"`
	Sender    string            `json:"sender"`
	Timestamp int64             `json:"timestamp"`
	TTL       int               `json:"ttl"` // Time To Live
	ID        string            `json:"id"`
}

// GossipNode представляет узел в gossip сети
type GossipNode struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	LastSeen time.Time `json:"last_seen"`
	Score    float64   `json:"score"`     // Репутация узла
	IsActive bool      `json:"is_active"` // Активен ли узел
}

// GossipConfig конфигурация для gossip протокола
type GossipConfig struct {
	Fanout            int           `json:"fanout"`             // Количество узлов для отправки
	MaxTTL            int           `json:"max_ttl"`            // Максимальный TTL
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // Интервал heartbeat
	MessageTimeout    time.Duration `json:"message_timeout"`    // Таймаут сообщения
	MaxRetries        int           `json:"max_retries"`        // Максимальное количество повторов
}

// GossipProtocol реализует gossip протокол
type GossipProtocol struct {
	nodeID     string
	config     GossipConfig
	nodes      map[string]*GossipNode
	nodesMux   sync.RWMutex
	messageID  string
	messageMux sync.RWMutex
	server     *Server
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewGossipProtocol создает новый gossip протокол
func NewGossipProtocol(nodeID string, config GossipConfig, server *Server) *GossipProtocol {
	ctx, cancel := context.WithCancel(context.Background())

	return &GossipProtocol{
		nodeID: nodeID,
		config: config,
		nodes:  make(map[string]*GossipNode),
		server: server,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start запускает gossip протокол
func (g *GossipProtocol) Start() {
	log.Printf("Starting Gossip protocol for node %s", g.nodeID)

	// Запускаем heartbeat
	go g.heartbeatLoop()

	// Запускаем очистку неактивных узлов
	go g.cleanupLoop()
}

// Stop останавливает gossip протокол
func (g *GossipProtocol) Stop() {
	log.Printf("Stopping Gossip protocol for node %s", g.nodeID)
	g.cancel()
}

// AddNode добавляет узел в gossip сеть
func (g *GossipProtocol) AddNode(nodeID, address string) {
	g.nodesMux.Lock()
	defer g.nodesMux.Unlock()

	g.nodes[nodeID] = &GossipNode{
		ID:       nodeID,
		Address:  address,
		LastSeen: time.Now(),
		Score:    1.0,
		IsActive: true,
	}

	log.Printf("Added node %s (%s) to gossip network", nodeID, address)
}

// RemoveNode удаляет узел из gossip сети
func (g *GossipProtocol) RemoveNode(nodeID string) {
	g.nodesMux.Lock()
	defer g.nodesMux.Unlock()

	if node, exists := g.nodes[nodeID]; exists {
		node.IsActive = false
		log.Printf("Removed node %s from gossip network", nodeID)
	}
}

// BroadcastBlock распространяет блок через gossip
func (g *GossipProtocol) BroadcastBlock(block *blockchain.Block) error {
	message := &GossipMessage{
		Type:      GossipBlock,
		Data:      block,
		Sender:    g.nodeID,
		Timestamp: time.Now().Unix(),
		TTL:       g.config.MaxTTL,
		ID:        g.generateMessageID(),
	}

	return g.broadcastMessage(message)
}

// BroadcastTransaction распространяет транзакцию через gossip
func (g *GossipProtocol) BroadcastTransaction(tx *blockchain.Transaction) error {
	message := &GossipMessage{
		Type:      GossipTransaction,
		Data:      tx,
		Sender:    g.nodeID,
		Timestamp: time.Now().Unix(),
		TTL:       g.config.MaxTTL,
		ID:        g.generateMessageID(),
	}

	return g.broadcastMessage(message)
}

// broadcastMessage распространяет сообщение через gossip
func (g *GossipProtocol) broadcastMessage(message *GossipMessage) error {
	if message.TTL <= 0 {
		return fmt.Errorf("message TTL expired")
	}

	// Выбираем случайные узлы для отправки
	targets := g.selectRandomNodes(g.config.Fanout)

	if len(targets) == 0 {
		log.Printf("No active nodes available for gossip broadcast")
		return nil
	}

	log.Printf("Broadcasting %s message %s to %d nodes", message.Type, message.ID, len(targets))

	// Отправляем сообщение выбранным узлам
	for _, target := range targets {
		go g.sendMessageToNode(target, message)
	}

	return nil
}

// selectRandomNodes выбирает случайные узлы для отправки
func (g *GossipProtocol) selectRandomNodes(count int) []*GossipNode {
	g.nodesMux.RLock()
	defer g.nodesMux.RUnlock()

	activeNodes := make([]*GossipNode, 0)
	for _, node := range g.nodes {
		if node.IsActive && node.ID != g.nodeID {
			activeNodes = append(activeNodes, node)
		}
	}

	if len(activeNodes) <= count {
		return activeNodes
	}

	// Выбираем случайные узлы
	selected := make([]*GossipNode, 0, count)
	used := make(map[string]bool)

	for len(selected) < count && len(selected) < len(activeNodes) {
		// Генерируем случайный индекс
		randIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(activeNodes))))
		index := int(randIndex.Int64())

		node := activeNodes[index]
		if !used[node.ID] {
			selected = append(selected, node)
			used[node.ID] = true
		}
	}

	return selected
}

// sendMessageToNode отправляет сообщение конкретному узлу
func (g *GossipProtocol) sendMessageToNode(node *GossipNode, message *GossipMessage) {
	// Здесь должна быть реализация отправки через P2P сеть
	// Пока что просто логируем
	log.Printf("Sending %s message %s to node %s (%s)",
		message.Type, message.ID, node.ID, node.Address)

	// Обновляем время последнего контакта
	g.updateNodeLastSeen(node.ID)
}

// updateNodeLastSeen обновляет время последнего контакта с узлом
func (g *GossipProtocol) updateNodeLastSeen(nodeID string) {
	g.nodesMux.Lock()
	defer g.nodesMux.Unlock()

	if node, exists := g.nodes[nodeID]; exists {
		node.LastSeen = time.Now()
		node.Score = g.calculateNodeScore(node)
	}
}

// calculateNodeScore рассчитывает репутацию узла
func (g *GossipProtocol) calculateNodeScore(node *GossipNode) float64 {
	// Простая формула: чем чаще узел отвечает, тем выше его репутация
	timeSinceLastSeen := time.Since(node.LastSeen)
	if timeSinceLastSeen > time.Hour {
		return 0.1 // Низкая репутация для неактивных узлов
	}

	// Базовый скор с учетом времени последнего контакта
	baseScore := 1.0
	timePenalty := float64(timeSinceLastSeen.Minutes()) / 60.0
	return baseScore - (timePenalty * 0.1)
}

// generateMessageID генерирует уникальный ID сообщения
func (g *GossipProtocol) generateMessageID() string {
	g.messageMux.Lock()
	defer g.messageMux.Unlock()

	// Простая генерация ID на основе времени и случайности
	timestamp := time.Now().UnixNano()
	random, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return fmt.Sprintf("%d_%d", timestamp, random.Int64())
}

// heartbeatLoop отправляет heartbeat сообщения
func (g *GossipProtocol) heartbeatLoop() {
	ticker := time.NewTicker(g.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.sendHeartbeat()
		}
	}
}

// sendHeartbeat отправляет heartbeat сообщение
func (g *GossipProtocol) sendHeartbeat() {
	message := &GossipMessage{
		Type:      GossipHeartbeat,
		Data:      map[string]interface{}{"node_id": g.nodeID},
		Sender:    g.nodeID,
		Timestamp: time.Now().Unix(),
		TTL:       1,
		ID:        g.generateMessageID(),
	}

	// Отправляем heartbeat всем активным узлам
	targets := g.selectRandomNodes(g.config.Fanout)
	for _, target := range targets {
		go g.sendMessageToNode(target, message)
	}
}

// cleanupLoop очищает неактивные узлы
func (g *GossipProtocol) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.cleanupInactiveNodes()
		}
	}
}

// cleanupInactiveNodes удаляет неактивные узлы
func (g *GossipProtocol) cleanupInactiveNodes() {
	g.nodesMux.Lock()
	defer g.nodesMux.Unlock()

	cutoff := time.Now().Add(-30 * time.Minute)

	for nodeID, node := range g.nodes {
		if node.LastSeen.Before(cutoff) {
			node.IsActive = false
			log.Printf("Marked node %s as inactive (last seen: %v)", nodeID, node.LastSeen)
		}
	}
}

// GetActiveNodes возвращает список активных узлов
func (g *GossipProtocol) GetActiveNodes() []*GossipNode {
	g.nodesMux.RLock()
	defer g.nodesMux.RUnlock()

	activeNodes := make([]*GossipNode, 0)
	for _, node := range g.nodes {
		if node.IsActive {
			activeNodes = append(activeNodes, node)
		}
	}

	return activeNodes
}

// GetNodeCount возвращает количество узлов в gossip сети
func (g *GossipProtocol) GetNodeCount() int {
	g.nodesMux.RLock()
	defer g.nodesMux.RUnlock()

	return len(g.nodes)
}

// GetStats возвращает статистику gossip протокола
func (g *GossipProtocol) GetStats() map[string]interface{} {
	g.nodesMux.RLock()
	defer g.nodesMux.RUnlock()

	activeCount := 0
	totalScore := 0.0

	for _, node := range g.nodes {
		if node.IsActive {
			activeCount++
			totalScore += node.Score
		}
	}

	avgScore := 0.0
	if activeCount > 0 {
		avgScore = totalScore / float64(activeCount)
	}

	return map[string]interface{}{
		"total_nodes":   len(g.nodes),
		"active_nodes":  activeCount,
		"average_score": avgScore,
		"fanout":        g.config.Fanout,
		"max_ttl":       g.config.MaxTTL,
	}
}
