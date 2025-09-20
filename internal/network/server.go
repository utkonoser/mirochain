package network

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// Server представляет P2P сервер
type Server struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Port         int                    `json:"port"`
	Peers        map[string]*Peer       `json:"peers"`
	Blockchain   *blockchain.Blockchain `json:"-"`
	Listener     net.Listener           `json:"-"`
	IsRunning    bool                   `json:"is_running"`
	MessageChan  chan *Message          `json:"-"`
	PeerChan     chan *Peer             `json:"-"`
	StopChan     chan bool              `json:"-"`
	WebSocket    *WebSocketServer       `json:"-"`
	DHT          *DHT                   `json:"-"`
	DHTServer    *DHTServer             `json:"-"`
	DHTClient    *DHTClient             `json:"-"`
	Gossip       *GossipProtocol        `json:"-"`
	RateLimiter  *RateLimiterManager    `json:"-"`
	NATTraversal *NATTraversal          `json:"-"`
	mutex        sync.RWMutex
}

// NewServer создает новый P2P сервер
func NewServer(address string, port int, bc *blockchain.Blockchain) *Server {
	server := &Server{
		ID:          generateNodeID(),
		Address:     address,
		Port:        port,
		Peers:       make(map[string]*Peer),
		Blockchain:  bc,
		MessageChan: make(chan *Message, 1000),
		PeerChan:    make(chan *Peer, 100),
		StopChan:    make(chan bool),
	}

	// Инициализируем WebSocket сервер
	server.WebSocket = NewWebSocketServer(server, bc)

	// Инициализируем DHT
	nodeID := GenerateDHTNodeID()
	server.DHT = NewDHT(nodeID, 8, server) // Kademlia K=8
	server.DHTServer = NewDHTServer(server.DHT, server)
	server.DHTClient = NewDHTClient(server.DHT, server)

	// Инициализируем Gossip протокол
	gossipConfig := GossipConfig{
		Fanout:            3,
		MaxTTL:            7,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    5 * time.Second,
		MaxRetries:        3,
	}
	server.Gossip = NewGossipProtocol(server.ID, gossipConfig, server)

	// Инициализируем Rate Limiter
	server.RateLimiter = NewRateLimiterManager()

	// Добавляем rate limiter'ы
	server.RateLimiter.AddRateLimiter("api", RateLimiterConfig{
		Type:        TokenBucket,
		MaxRequests: 100,
		WindowSize:  time.Minute,
		BurstSize:   20,
		RefillRate:  10,
	})

	server.RateLimiter.AddRateLimiter("p2p", RateLimiterConfig{
		Type:        SlidingWindow,
		MaxRequests: 50,
		WindowSize:  time.Minute,
	})

	// Инициализируем NAT Traversal
	natConfig := NATTraversalConfig{
		STUNServers:       []string{"stun.l.google.com:19302", "stun1.l.google.com:19302"},
		STUNTimeout:       5 * time.Second,
		HolePunchTimeout:  10 * time.Second,
		KeepAliveInterval: 30 * time.Second,
		MaxRetries:        5,
	}
	server.NATTraversal = NewNATTraversal(natConfig)

	return server
}

// Start запускает P2P сервер
func (s *Server) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.IsRunning {
		return fmt.Errorf("server is already running")
	}

	// Регистрируем типы для gob
	RegisterTypes()

	// Запускаем TCP сервер
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Address, s.Port))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.Listener = listener
	s.IsRunning = true

	// Запускаем обработчики
	go s.handleMessages()
	go s.handlePeers()
	go s.acceptConnections()

	// Запускаем WebSocket сервер в отдельной горутине
	go func() {
		wsPort := s.Port + 1000 // WebSocket на порту +1000
		if err := s.WebSocket.Start(wsPort); err != nil {
			slog.Error("Failed to start WebSocket server", "error", err)
		}
	}()

	// Запускаем DHT сервер в отдельной горутине
	go func() {
		dhtPort := s.Port + 2000 // DHT на порту +2000
		if err := s.DHTServer.Start(dhtPort); err != nil {
			slog.Error("Failed to start DHT server", "error", err)
		}
	}()

	// Запускаем DHT
	if err := s.DHT.Start(); err != nil {
		slog.Error("Failed to start DHT", "error", err)
	}

	// Запускаем DHT bootstrap
	go func() {
		if err := s.DHTClient.Bootstrap(); err != nil {
			slog.Error("Failed to bootstrap DHT", "error", err)
		}
	}()

	// Запускаем Gossip протокол
	go s.Gossip.Start()

	// Запускаем NAT Traversal
	go func() {
		if err := s.NATTraversal.Start(s.Port); err != nil {
			slog.Error("Failed to start NAT Traversal", "error", err)
		}
	}()

	slog.Info("P2P server started", "address", s.Address, "port", s.Port, "node_id", s.ID)
	slog.Info("WebSocket server started", "port", s.Port+1000)
	slog.Info("DHT server started", "port", s.Port+2000, "dht_node_id", fmt.Sprintf("%x", s.DHT.nodeID))
	return nil
}

// Stop останавливает P2P сервер
func (s *Server) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.IsRunning {
		return fmt.Errorf("server is not running")
	}

	s.IsRunning = false
	s.StopChan <- true

	// Закрываем все соединения
	for _, peer := range s.Peers {
		peer.Close()
	}

	// Закрываем listener
	if s.Listener != nil {
		s.Listener.Close()
	}

	slog.Info("P2P server stopped")
	return nil
}

// acceptConnections принимает входящие соединения
func (s *Server) acceptConnections() {
	for s.IsRunning {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.IsRunning {
				slog.Error("Failed to accept connection", "error", err)
			}
			continue
		}

		// Создаем peer для нового соединения
		peer := NewPeer(generateNodeID(), conn.RemoteAddr().String(), conn)
		s.PeerChan <- peer

		// Запускаем обработчик для peer'а
		go s.handlePeer(peer)
	}
}

// handlePeer обрабатывает соединение с peer'ом
func (s *Server) handlePeer(peer *Peer) {
	defer peer.Close()

	slog.Info("New peer connected", "peer", peer.Address)

	// Добавляем peer в список
	s.addPeer(peer)
	defer s.removePeer(peer.ID)

	// Обрабатываем сообщения от peer'а
	for s.IsRunning && peer.IsConnected() {
		msg, err := peer.ReadMessage()
		if err != nil {
			slog.Error("Failed to read message from peer", "peer", peer.Address, "error", err)
			break
		}

		// Отправляем сообщение на обработку
		s.MessageChan <- msg
	}

	slog.Info("Peer disconnected", "peer", peer.Address)
}

// handleMessages обрабатывает входящие сообщения
func (s *Server) handleMessages() {
	for s.IsRunning {
		select {
		case msg := <-s.MessageChan:
			s.processMessage(msg)
		case <-s.StopChan:
			return
		}
	}
}

// handlePeers обрабатывает события с peer'ами
func (s *Server) handlePeers() {
	for s.IsRunning {
		select {
		case peer := <-s.PeerChan:
			s.addPeer(peer)
		case <-s.StopChan:
			return
		}
	}
}

// processMessage обрабатывает сообщение
func (s *Server) processMessage(msg *Message) {
	slog.Debug("Processing message", "type", msg.Type, "from", msg.From)

	switch msg.Type {
	case MessageTypeHandshake:
		s.handleHandshake(msg)
	case MessageTypeHandshakeAck:
		s.handleHandshakeAck(msg)
	case MessageTypeNewBlock:
		s.handleNewBlock(msg)
	case MessageTypeNewTransaction:
		s.handleNewTransaction(msg)
	case MessageTypeSyncRequest:
		s.handleSyncRequest(msg)
	case MessageTypePing:
		s.handlePing(msg)
	default:
		slog.Warn("Unknown message type", "type", msg.Type)
	}
}

// addPeer добавляет peer в список
func (s *Server) addPeer(peer *Peer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Peers[peer.ID] = peer

	// Отправляем WebSocket уведомление
	if s.WebSocket != nil {
		s.WebSocket.BroadcastPeerConnected(peer)
	}
}

// removePeer удаляет peer из списка
func (s *Server) removePeer(peerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Получаем peer перед удалением для уведомления
	peer, exists := s.Peers[peerID]
	if exists {
		// Отправляем WebSocket уведомление
		if s.WebSocket != nil {
			s.WebSocket.BroadcastPeerDisconnected(peer)
		}
		delete(s.Peers, peerID)
	}
}

// GetPeers возвращает список активных peer'ов
func (s *Server) GetPeers() []*Peer {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var peers []*Peer
	for _, peer := range s.Peers {
		if peer.IsConnected() {
			peers = append(peers, peer)
		}
	}
	return peers
}

// GetPeerCount возвращает количество активных peer'ов
func (s *Server) GetPeerCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.Peers)
}

// generateNodeID генерирует уникальный ID узла
func generateNodeID() string {
	return fmt.Sprintf("node_%d", time.Now().UnixNano())
}

// String возвращает строковое представление сервера
func (s *Server) String() string {
	return fmt.Sprintf("Server{ID: %s, Address: %s:%d, Peers: %d, Running: %t}",
		s.ID, s.Address, s.Port, s.GetPeerCount(), s.IsRunning)
}

// GetWebSocketClientCount возвращает количество WebSocket клиентов
func (s *Server) GetWebSocketClientCount() int {
	if s.WebSocket == nil {
		return 0
	}
	return s.WebSocket.GetClientCount()
}

// BroadcastBalanceUpdate отправляет уведомление об обновлении баланса
func (s *Server) BroadcastBalanceUpdate(address string, balance, change int64) {
	if s.WebSocket != nil {
		s.WebSocket.BroadcastBalanceUpdate(address, balance, change)
	}
}

// AddBootstrapNode добавляет bootstrap узел в DHT
func (s *Server) AddBootstrapNode(address string) {
	if s.DHT != nil {
		s.DHT.AddBootstrapNode(address)
	}
}

// GetDHTStats возвращает статистику DHT
func (s *Server) GetDHTStats() map[string]interface{} {
	if s.DHT == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"node_id":    fmt.Sprintf("%x", s.DHT.nodeID),
		"peer_count": s.DHT.GetPeerCount(),
		"bootstrap":  s.DHT.bootstrap,
	}
}

// DiscoverPeers использует DHT для поиска новых peer'ов
func (s *Server) DiscoverPeers() error {
	if s.DHTClient == nil {
		return fmt.Errorf("DHT client not initialized")
	}
	return s.DHTClient.DiscoverPeers()
}

// GetDHTPeers возвращает список peer'ов из DHT
func (s *Server) GetDHTPeers() []*PeerInfo {
	if s.DHT == nil {
		return []*PeerInfo{}
	}
	return s.DHT.GetAllPeers()
}

// GetGossipStats возвращает статистику Gossip протокола
func (s *Server) GetGossipStats() map[string]interface{} {
	if s.Gossip == nil {
		return map[string]interface{}{}
	}
	return s.Gossip.GetStats()
}

// GetRateLimiterStats возвращает статистику Rate Limiter'ов
func (s *Server) GetRateLimiterStats() map[string]interface{} {
	if s.RateLimiter == nil {
		return map[string]interface{}{}
	}
	return s.RateLimiter.GetAllStats()
}

// GetNATStats возвращает статистику NAT Traversal
func (s *Server) GetNATStats() map[string]interface{} {
	if s.NATTraversal == nil {
		return map[string]interface{}{}
	}
	return s.NATTraversal.GetStats()
}

// AddGossipNode добавляет узел в Gossip сеть
func (s *Server) AddGossipNode(nodeID, address string) {
	if s.Gossip != nil {
		s.Gossip.AddNode(nodeID, address)
	}
}

// RemoveGossipNode удаляет узел из Gossip сети
func (s *Server) RemoveGossipNode(nodeID string) {
	if s.Gossip != nil {
		s.Gossip.RemoveNode(nodeID)
	}
}

// BroadcastBlockGossip распространяет блок через Gossip
func (s *Server) BroadcastBlockGossip(block *blockchain.Block) error {
	if s.Gossip == nil {
		return fmt.Errorf("Gossip protocol not initialized")
	}
	return s.Gossip.BroadcastBlock(block)
}

// BroadcastTransactionGossip распространяет транзакцию через Gossip
func (s *Server) BroadcastTransactionGossip(tx *blockchain.Transaction) error {
	if s.Gossip == nil {
		return fmt.Errorf("Gossip protocol not initialized")
	}
	return s.Gossip.BroadcastTransaction(tx)
}

// CheckRateLimit проверяет rate limit для клиента
func (s *Server) CheckRateLimit(limiterName, clientID string) bool {
	if s.RateLimiter == nil {
		return true // Разрешаем, если rate limiter не инициализирован
	}
	return s.RateLimiter.Allow(limiterName, clientID)
}

// AddNATPeer добавляет peer для NAT Traversal
func (s *Server) AddNATPeer(peerID, internalAddr, externalAddr string, natType NATType) {
	if s.NATTraversal != nil {
		s.NATTraversal.AddPeer(peerID, internalAddr, externalAddr, natType)
	}
}

// EstablishNATConnection устанавливает соединение через NAT
func (s *Server) EstablishNATConnection(peerID string) error {
	if s.NATTraversal == nil {
		return fmt.Errorf("NAT Traversal not initialized")
	}
	return s.NATTraversal.EstablishConnection(peerID)
}
