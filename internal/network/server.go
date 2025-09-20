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
	ID          string                 `json:"id"`
	Address     string                 `json:"address"`
	Port        int                    `json:"port"`
	Peers       map[string]*Peer       `json:"peers"`
	Blockchain  *blockchain.Blockchain `json:"-"`
	Listener    net.Listener           `json:"-"`
	IsRunning   bool                   `json:"is_running"`
	MessageChan chan *Message          `json:"-"`
	PeerChan    chan *Peer             `json:"-"`
	StopChan    chan bool              `json:"-"`
	mutex       sync.RWMutex
}

// NewServer создает новый P2P сервер
func NewServer(address string, port int, bc *blockchain.Blockchain) *Server {
	return &Server{
		ID:          generateNodeID(),
		Address:     address,
		Port:        port,
		Peers:       make(map[string]*Peer),
		Blockchain:  bc,
		MessageChan: make(chan *Message, 1000),
		PeerChan:    make(chan *Peer, 100),
		StopChan:    make(chan bool),
	}
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

	slog.Info("P2P server started", "address", s.Address, "port", s.Port, "node_id", s.ID)
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
}

// removePeer удаляет peer из списка
func (s *Server) removePeer(peerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.Peers, peerID)
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
