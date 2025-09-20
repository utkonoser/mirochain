package network

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"mirochain/internal/blockchain"
)

// WebSocketServer представляет WebSocket сервер для real-time уведомлений
type WebSocketServer struct {
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]*WSClient
	clientsMux  sync.RWMutex
	register    chan *WSClient
	unregister  chan *WSClient
	broadcast   chan []byte
	blockchain  *blockchain.Blockchain
	server      *Server
}

// WSClient представляет WebSocket клиента
type WSClient struct {
	conn     *websocket.Conn
	send     chan []byte
	server   *WebSocketServer
	userID   string
	subscribed map[string]bool // Подписки на события
}

// NotificationType определяет тип уведомления
type NotificationType string

const (
	NotificationTypeNewBlock       NotificationType = "new_block"
	NotificationTypeNewTransaction NotificationType = "new_transaction"
	NotificationTypeBalanceUpdate  NotificationType = "balance_update"
	NotificationTypePeerConnected  NotificationType = "peer_connected"
	NotificationTypePeerDisconnected NotificationType = "peer_disconnected"
	NotificationTypeNetworkStatus  NotificationType = "network_status"
)

// Notification представляет уведомление
type Notification struct {
	Type      NotificationType `json:"type"`
	Data      interface{}      `json:"data"`
	Timestamp int64            `json:"timestamp"`
	NodeID    string           `json:"node_id"`
}

// BlockNotification содержит данные о новом блоке
type BlockNotification struct {
	Height    int64  `json:"height"`
	Hash      string `json:"hash"`
	Timestamp int64  `json:"timestamp"`
	TxCount   int    `json:"tx_count"`
	Miner     string `json:"miner"`
}

// TransactionNotification содержит данные о новой транзакции
type TransactionNotification struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Timestamp int64  `json:"timestamp"`
}

// BalanceNotification содержит данные об обновлении баланса
type BalanceNotification struct {
	Address string `json:"address"`
	Balance int64  `json:"balance"`
	Change  int64  `json:"change"`
}

// PeerNotification содержит данные о peer'е
type PeerNotification struct {
	PeerID  string `json:"peer_id"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

// NetworkStatusNotification содержит статус сети
type NetworkStatusNotification struct {
	PeerCount    int    `json:"peer_count"`
	BlockHeight  int64  `json:"block_height"`
	NetworkHash  string `json:"network_hash"`
	Uptime       int64  `json:"uptime"`
}

// NewWebSocketServer создает новый WebSocket сервер
func NewWebSocketServer(server *Server, blockchain *blockchain.Blockchain) *WebSocketServer {
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // В продакшене нужно добавить проверку origin
			},
		},
		clients:    make(map[*websocket.Conn]*WSClient),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan []byte),
		blockchain: blockchain,
		server:     server,
	}
}

// Start запускает WebSocket сервер
func (ws *WebSocketServer) Start(port int) error {
	// Запускаем hub для управления клиентами
	go ws.run()

	// Настраиваем HTTP маршруты
	http.HandleFunc("/ws", ws.HandleWebSocket)
	http.HandleFunc("/ws/status", ws.HandleStatus)

	slog.Info("WebSocket server starting", "port", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// run управляет клиентами WebSocket
func (ws *WebSocketServer) run() {
	for {
		select {
		case client := <-ws.register:
			ws.registerClient(client)

		case client := <-ws.unregister:
			ws.unregisterClient(client)

		case message := <-ws.broadcast:
			ws.broadcastToClients(message)
		}
	}
}

// HandleWebSocket обрабатывает WebSocket соединения
func (ws *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}

	client := &WSClient{
		conn:       conn,
		send:       make(chan []byte, 256),
		server:     ws,
		userID:     r.URL.Query().Get("user_id"),
		subscribed: make(map[string]bool),
	}

	ws.register <- client

	// Запускаем горутины для чтения и записи
	go client.writePump()
	go client.readPump()
}

// HandleStatus возвращает статус WebSocket сервера
func (ws *WebSocketServer) HandleStatus(w http.ResponseWriter, r *http.Request) {
	ws.clientsMux.RLock()
	clientCount := len(ws.clients)
	ws.clientsMux.RUnlock()

	status := map[string]interface{}{
		"clients":     clientCount,
		"blockchain":  ws.blockchain.GetHeight(),
		"peers":       ws.server.GetPeerCount(),
		"server_time": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// registerClient регистрирует нового клиента
func (ws *WebSocketServer) registerClient(client *WSClient) {
	ws.clientsMux.Lock()
	defer ws.clientsMux.Unlock()

	ws.clients[client.conn] = client
	slog.Info("WebSocket client connected", "user_id", client.userID, "clients", len(ws.clients))

	// Отправляем приветственное сообщение
	welcome := Notification{
		Type:      NotificationTypeNetworkStatus,
		Data:      ws.getNetworkStatus(),
		Timestamp: time.Now().Unix(),
		NodeID:    ws.server.ID,
	}
	client.sendNotification(welcome)
}

// unregisterClient удаляет клиента
func (ws *WebSocketServer) unregisterClient(client *WSClient) {
	ws.clientsMux.Lock()
	defer ws.clientsMux.Unlock()

	if _, ok := ws.clients[client.conn]; ok {
		delete(ws.clients, client.conn)
		close(client.send)
		slog.Info("WebSocket client disconnected", "user_id", client.userID, "clients", len(ws.clients))
	}
}

// broadcastToClients отправляет сообщение всем клиентам
func (ws *WebSocketServer) broadcastToClients(message []byte) {
	ws.clientsMux.RLock()
	defer ws.clientsMux.RUnlock()

	for _, client := range ws.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(ws.clients, client.conn)
		}
	}
}

// BroadcastNewBlock уведомляет о новом блоке
func (ws *WebSocketServer) BroadcastNewBlock(block *blockchain.Block) {
	notification := Notification{
		Type: NotificationTypeNewBlock,
		Data: BlockNotification{
			Height:    block.Height,
			Hash:      fmt.Sprintf("%x", block.Hash),
			Timestamp: block.Timestamp,
			TxCount:   len(block.Transactions),
			Miner:     "unknown", // TODO: добавить информацию о майнере
		},
		Timestamp: time.Now().Unix(),
		NodeID:    ws.server.ID,
	}

	ws.broadcastNotification(notification)
}

// BroadcastNewTransaction уведомляет о новой транзакции
func (ws *WebSocketServer) BroadcastNewTransaction(tx *blockchain.Transaction) {
	notification := Notification{
		Type: NotificationTypeNewTransaction,
		Data: TransactionNotification{
			ID:        fmt.Sprintf("%x", tx.ID),
			From:      "unknown", // Упрощенно
			To:        "unknown", // Упрощенно
			Amount:    tx.Outputs[0].Value,
			Timestamp: time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
		NodeID:    ws.server.ID,
	}

	ws.broadcastNotification(notification)
}

// BroadcastBalanceUpdate уведомляет об обновлении баланса
func (ws *WebSocketServer) BroadcastBalanceUpdate(address string, balance, change int64) {
	notification := Notification{
		Type: NotificationTypeBalanceUpdate,
		Data: BalanceNotification{
			Address: address,
			Balance: balance,
			Change:  change,
		},
		Timestamp: time.Now().Unix(),
		NodeID:    ws.server.ID,
	}

	ws.broadcastNotification(notification)
}

// BroadcastPeerConnected уведомляет о подключении peer'а
func (ws *WebSocketServer) BroadcastPeerConnected(peer *Peer) {
	notification := Notification{
		Type: NotificationTypePeerConnected,
		Data: PeerNotification{
			PeerID:  peer.ID,
			Address: peer.Address,
			Status:  "connected",
		},
		Timestamp: time.Now().Unix(),
		NodeID:    ws.server.ID,
	}

	ws.broadcastNotification(notification)
}

// BroadcastPeerDisconnected уведомляет об отключении peer'а
func (ws *WebSocketServer) BroadcastPeerDisconnected(peer *Peer) {
	notification := Notification{
		Type: NotificationTypePeerDisconnected,
		Data: PeerNotification{
			PeerID:  peer.ID,
			Address: peer.Address,
			Status:  "disconnected",
		},
		Timestamp: time.Now().Unix(),
		NodeID:    ws.server.ID,
	}

	ws.broadcastNotification(notification)
}

// broadcastNotification отправляет уведомление всем клиентам
func (ws *WebSocketServer) broadcastNotification(notification Notification) {
	message, err := json.Marshal(notification)
	if err != nil {
		slog.Error("Failed to marshal notification", "error", err)
		return
	}

	ws.broadcast <- message
}

// getNetworkStatus возвращает текущий статус сети
func (ws *WebSocketServer) getNetworkStatus() NetworkStatusNotification {
	height := ws.blockchain.GetHeight()
	return NetworkStatusNotification{
		PeerCount:   ws.server.GetPeerCount(),
		BlockHeight: height,
		NetworkHash: fmt.Sprintf("%x", ws.blockchain.GetLastBlock().Hash),
		Uptime:      time.Now().Unix(),
	}
}

// readPump читает сообщения от клиента
func (c *WSClient) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket error", "error", err)
			}
			break
		}
	}
}

// writePump отправляет сообщения клиенту
func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Добавляем дополнительные сообщения из очереди
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendNotification отправляет уведомление клиенту
func (c *WSClient) sendNotification(notification Notification) {
	message, err := json.Marshal(notification)
	if err != nil {
		slog.Error("Failed to marshal notification", "error", err)
		return
	}

	select {
	case c.send <- message:
	default:
		close(c.send)
	}
}

// GetClientCount возвращает количество подключенных клиентов
func (ws *WebSocketServer) GetClientCount() int {
	ws.clientsMux.RLock()
	defer ws.clientsMux.RUnlock()
	return len(ws.clients)
}
