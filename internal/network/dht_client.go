package network

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"
)

// DHTClient представляет клиент для DHT операций
type DHTClient struct {
	dht    *DHT
	server *Server
}

// NewDHTClient создает новый DHT клиент
func NewDHTClient(dht *DHT, server *Server) *DHTClient {
	return &DHTClient{
		dht:    dht,
		server: server,
	}
}

// Ping отправляет ping сообщение peer'у
func (c *DHTClient) Ping(peerAddr string) error {
	conn, err := net.DialTimeout("tcp", peerAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerAddr, err)
	}
	defer conn.Close()

	// Создаем ping сообщение
	pingMsg := DHTMessage{
		Type:      string(DHTMessageTypePing),
		SenderID:  string(c.dht.nodeID),
		TargetID:  "",
		Data:      nil,
		Timestamp: time.Now().Unix(),
	}

	// Отправляем сообщение
	if err := c.sendMessage(conn, pingMsg); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	// Ждем ответ
	var pongMsg DHTMessage
	if err := c.receiveMessage(conn, &pongMsg); err != nil {
		return fmt.Errorf("failed to receive pong: %w", err)
	}

	if pongMsg.Type != string(DHTMessageTypePong) {
		return fmt.Errorf("expected pong, got %s", pongMsg.Type)
	}

	slog.Debug("Ping successful", "peer", peerAddr)
	return nil
}

// FindNode отправляет запрос на поиск узла
func (c *DHTClient) FindNode(peerAddr, targetID string) ([]*PeerInfo, error) {
	conn, err := net.DialTimeout("tcp", peerAddr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %w", peerAddr, err)
	}
	defer conn.Close()

	// Создаем find_node сообщение
	findMsg := DHTMessage{
		Type:     string(DHTMessageTypeFindNode),
		SenderID: string(c.dht.nodeID),
		TargetID: targetID,
		Data: FindNodeData{
			TargetID: targetID,
			Count:    8,
		},
		Timestamp: time.Now().Unix(),
	}

	// Отправляем сообщение
	if err := c.sendMessage(conn, findMsg); err != nil {
		return nil, fmt.Errorf("failed to send find_node: %w", err)
	}

	// Ждем ответ
	var response DHTMessage
	if err := c.receiveMessage(conn, &response); err != nil {
		return nil, fmt.Errorf("failed to receive find_node response: %w", err)
	}

	if response.Type != string(DHTMessageTypeFindNodeRes) {
		return nil, fmt.Errorf("expected find_node_res, got %s", response.Type)
	}

	// Парсим ответ
	var findRes FindNodeResponse
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", response.Data)), &findRes); err != nil {
		return nil, fmt.Errorf("failed to parse find_node response: %w", err)
	}

	return findRes.Peers, nil
}

// Store отправляет запрос на хранение значения
func (c *DHTClient) Store(peerAddr, key string, value interface{}, ttl int64) error {
	conn, err := net.DialTimeout("tcp", peerAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", peerAddr, err)
	}
	defer conn.Close()

	// Создаем store сообщение
	storeMsg := DHTMessage{
		Type:     string(DHTMessageTypeStore),
		SenderID: string(c.dht.nodeID),
		TargetID: "",
		Data: StoreData{
			Key:   key,
			Value: value,
			TTL:   ttl,
		},
		Timestamp: time.Now().Unix(),
	}

	// Отправляем сообщение
	if err := c.sendMessage(conn, storeMsg); err != nil {
		return fmt.Errorf("failed to send store: %w", err)
	}

	// Ждем ответ
	var response DHTMessage
	if err := c.receiveMessage(conn, &response); err != nil {
		return fmt.Errorf("failed to receive store response: %w", err)
	}

	if response.Type != string(DHTMessageTypeStoreRes) {
		return fmt.Errorf("expected store_res, got %s", response.Type)
	}

	slog.Debug("Store successful", "peer", peerAddr, "key", key)
	return nil
}

// GetValue отправляет запрос на получение значения
func (c *DHTClient) GetValue(peerAddr, key string) (interface{}, bool, error) {
	conn, err := net.DialTimeout("tcp", peerAddr, 5*time.Second)
	if err != nil {
		return nil, false, fmt.Errorf("failed to connect to peer %s: %w", peerAddr, err)
	}
	defer conn.Close()

	// Создаем get_value сообщение
	getMsg := DHTMessage{
		Type:     string(DHTMessageTypeGetValue),
		SenderID: string(c.dht.nodeID),
		TargetID: "",
		Data: GetValueData{
			Key: key,
		},
		Timestamp: time.Now().Unix(),
	}

	// Отправляем сообщение
	if err := c.sendMessage(conn, getMsg); err != nil {
		return nil, false, fmt.Errorf("failed to send get_value: %w", err)
	}

	// Ждем ответ
	var response DHTMessage
	if err := c.receiveMessage(conn, &response); err != nil {
		return nil, false, fmt.Errorf("failed to receive get_value response: %w", err)
	}

	if response.Type != string(DHTMessageTypeGetValueRes) {
		return nil, false, fmt.Errorf("expected get_value_res, got %s", response.Type)
	}

	// Парсим ответ
	var getRes GetValueResponse
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", response.Data)), &getRes); err != nil {
		return nil, false, fmt.Errorf("failed to parse get_value response: %w", err)
	}

	return getRes.Value, getRes.Found, nil
}

// sendMessage отправляет сообщение через соединение
func (c *DHTClient) sendMessage(conn net.Conn, msg DHTMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Добавляем длину сообщения в начало
	length := make([]byte, 4)
	length[0] = byte(len(data) >> 24)
	length[1] = byte(len(data) >> 16)
	length[2] = byte(len(data) >> 8)
	length[3] = byte(len(data))

	// Отправляем длину + данные
	if _, err := conn.Write(length); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// receiveMessage получает сообщение через соединение
func (c *DHTClient) receiveMessage(conn net.Conn, msg *DHTMessage) error {
	// Читаем длину сообщения
	lengthBytes := make([]byte, 4)
	if _, err := conn.Read(lengthBytes); err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}

	length := int(lengthBytes[0])<<24 | int(lengthBytes[1])<<16 | int(lengthBytes[2])<<8 | int(lengthBytes[3])

	// Читаем данные
	data := make([]byte, length)
	if _, err := conn.Read(data); err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	// Парсим JSON
	if err := json.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

// DiscoverPeers использует DHT для поиска новых peer'ов
func (c *DHTClient) DiscoverPeers() error {
	// Получаем случайный target ID для поиска
	targetID := string(GenerateDHTNodeID())

	// Ищем ближайших peer'ов
	closestPeers := c.dht.FindNode(targetID)

	slog.Info("Discovering peers via DHT", "target", targetID, "closest_peers", len(closestPeers))

	// Обращаемся к ближайшим peer'ам для поиска новых
	for _, peer := range closestPeers {
		if peer.Address == "" {
			continue
		}

		peerAddr := fmt.Sprintf("%s:%d", peer.Address, peer.Port)

		// Запрашиваем у peer'а его ближайших соседей
		newPeers, err := c.FindNode(peerAddr, targetID)
		if err != nil {
			slog.Debug("Failed to find nodes from peer", "peer", peerAddr, "error", err)
			continue
		}

		// Добавляем новых peer'ов в DHT
		for _, newPeer := range newPeers {
			if newPeer.ID != string(c.dht.nodeID) { // Не добавляем себя
				c.dht.AddPeer(newPeer)
			}
		}
	}

	return nil
}

// Bootstrap выполняет начальную загрузку DHT
func (c *DHTClient) Bootstrap() error {
	slog.Info("Bootstrapping DHT")

	// Подключаемся к bootstrap узлам
	for _, bootstrapAddr := range c.dht.bootstrap {
		// Пингуем bootstrap узел
		if err := c.Ping(bootstrapAddr); err != nil {
			slog.Debug("Failed to ping bootstrap node", "address", bootstrapAddr, "error", err)
			continue
		}

		// Запрашиваем список peer'ов у bootstrap узла
		peers, err := c.FindNode(bootstrapAddr, string(c.dht.nodeID))
		if err != nil {
			slog.Debug("Failed to get peers from bootstrap node", "address", bootstrapAddr, "error", err)
			continue
		}

		// Добавляем peer'ов в DHT
		for _, peer := range peers {
			c.dht.AddPeer(peer)
		}

		slog.Info("Bootstrap successful", "address", bootstrapAddr, "peers_added", len(peers))
	}

	// Запускаем поиск новых peer'ов
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := c.DiscoverPeers(); err != nil {
				slog.Debug("Peer discovery failed", "error", err)
			}
		}
	}()

	return nil
}
