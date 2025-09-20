package network

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"mirochain/internal/blockchain"
)

// Client представляет P2P клиент для подключения к другим узлам
type Client struct {
	Server *Server
}

// NewClient создает новый P2P клиент
func NewClient(server *Server) *Client {
	return &Client{
		Server: server,
	}
}

// ConnectToPeer подключается к другому узлу
func (c *Client) ConnectToPeer(address string) error {
	slog.Info("Connecting to peer", "address", address)

	// Устанавливаем соединение
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", address, err)
	}

	// Создаем peer
	peer := NewPeer(generateNodeID(), address, conn)

	// Отправляем handshake
	err = c.sendHandshake(peer)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Добавляем peer в сервер
	c.Server.PeerChan <- peer

	// Запускаем обработчик для peer'а
	go c.Server.handlePeer(peer)

	slog.Info("Successfully connected to peer", "address", address)
	return nil
}

// sendHandshake отправляет handshake сообщение
func (c *Client) sendHandshake(peer *Peer) error {
	handshakeData := &HandshakeData{
		Version:     "1.0.0",
		NodeID:      c.Server.ID,
		BestHeight:  c.Server.Blockchain.GetHeight(),
		GenesisHash: c.Server.Blockchain.GetGenesisHash(),
	}

	msg := NewMessage(MessageTypeHandshake, handshakeData, c.Server.ID, peer.ID)
	return peer.SendMessage(msg)
}

// ConnectToPeers подключается к списку узлов
func (c *Client) ConnectToPeers(addresses []string) {
	for _, address := range addresses {
		go func(addr string) {
			err := c.ConnectToPeer(addr)
			if err != nil {
				slog.Error("Failed to connect to peer", "address", addr, "error", err)
			}
		}(address)
	}
}

// DiscoverPeers пытается обнаружить другие узлы
func (c *Client) DiscoverPeers() {
	// В реальной реализации здесь может быть:
	// - DNS seed nodes
	// - Hardcoded bootstrap nodes
	// - DHT discovery
	// - mDNS discovery

	slog.Info("Peer discovery not implemented yet")
}

// PingPeer отправляет ping сообщение peer'у
func (c *Client) PingPeer(peerID string) error {
	msg := NewMessage(MessageTypePing, nil, c.Server.ID, peerID)
	c.Server.broadcastToPeer(peerID, msg)
	return nil
}

// RequestSync запрашивает синхронизацию у peer'а
func (c *Client) RequestSync(peerID string, startHeight, endHeight int64) error {
	syncData := &SyncData{
		StartHeight: startHeight,
		EndHeight:   endHeight,
	}

	msg := NewMessage(MessageTypeSyncRequest, syncData, c.Server.ID, peerID)
	c.Server.broadcastToPeer(peerID, msg)
	return nil
}

// BroadcastBlock распространяет блок по сети
func (c *Client) BroadcastBlock(block *blockchain.Block) {
	c.Server.BroadcastNewBlock(block)
}

// BroadcastTransaction распространяет транзакцию по сети
func (c *Client) BroadcastTransaction(tx *blockchain.Transaction) {
	c.Server.BroadcastNewTransaction(tx)
}

// GetConnectedPeers возвращает список подключенных peer'ов
func (c *Client) GetConnectedPeers() []*Peer {
	return c.Server.GetPeers()
}

// DisconnectFromPeer отключается от peer'а
func (c *Client) DisconnectFromPeer(peerID string) error {
	c.Server.mutex.Lock()
	defer c.Server.mutex.Unlock()

	peer, exists := c.Server.Peers[peerID]
	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	return peer.Close()
}
