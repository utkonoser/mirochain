package network

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"
)

// DHTServer представляет сервер для обработки DHT сообщений
type DHTServer struct {
	dht      *DHT
	server   *Server
	listener net.Listener
	running  bool
}

// NewDHTServer создает новый DHT сервер
func NewDHTServer(dht *DHT, server *Server) *DHTServer {
	return &DHTServer{
		dht:    dht,
		server: server,
	}
}

// Start запускает DHT сервер
func (s *DHTServer) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start DHT server: %w", err)
	}

	s.listener = listener
	s.running = true

	slog.Info("DHT server started", "port", port)

	// Запускаем обработчик соединений
	go s.handleConnections()

	return nil
}

// Stop останавливает DHT сервер
func (s *DHTServer) Stop() error {
	s.running = false
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// handleConnections обрабатывает входящие соединения
func (s *DHTServer) handleConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				slog.Error("Failed to accept DHT connection", "error", err)
			}
			continue
		}

		// Обрабатываем соединение в отдельной горутине
		go s.handleConnection(conn)
	}
}

// handleConnection обрабатывает отдельное соединение
func (s *DHTServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Устанавливаем таймаут
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Читаем сообщение
	var msg DHTMessage
	if err := s.receiveMessage(conn, &msg); err != nil {
		slog.Error("Failed to receive DHT message", "error", err)
		return
	}

	// Обрабатываем сообщение
	response, err := s.handleMessage(msg)
	if err != nil {
		slog.Error("Failed to handle DHT message", "error", err)
		return
	}

	// Отправляем ответ
	if response != nil {
		if err := s.sendMessage(conn, *response); err != nil {
			slog.Error("Failed to send DHT response", "error", err)
		}
	}
}

// handleMessage обрабатывает DHT сообщение
func (s *DHTServer) handleMessage(msg DHTMessage) (*DHTMessage, error) {
	switch DHTMessageType(msg.Type) {
	case DHTMessageTypePing:
		return s.handlePing(msg)
	case DHTMessageTypeFindNode:
		return s.handleFindNode(msg)
	case DHTMessageTypeStore:
		return s.handleStore(msg)
	case DHTMessageTypeGetValue:
		return s.handleGetValue(msg)
	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handlePing обрабатывает ping сообщение
func (s *DHTServer) handlePing(msg DHTMessage) (*DHTMessage, error) {
	slog.Debug("Received ping", "from", msg.SenderID)

	// Создаем pong ответ
	pong := DHTMessage{
		Type:      string(DHTMessageTypePong),
		SenderID:  string(s.dht.nodeID),
		TargetID:  msg.SenderID,
		Data:      nil,
		Timestamp: time.Now().Unix(),
	}

	return &pong, nil
}

// handleFindNode обрабатывает find_node сообщение
func (s *DHTServer) handleFindNode(msg DHTMessage) (*DHTMessage, error) {
	slog.Debug("Received find_node", "from", msg.SenderID, "target", msg.TargetID)

	// Парсим данные запроса
	var findData FindNodeData
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Data)), &findData); err != nil {
		return nil, fmt.Errorf("failed to parse find_node data: %w", err)
	}

	// Ищем ближайших peer'ов
	peers := s.dht.FindNode(findData.TargetID)

	// Создаем ответ
	response := DHTMessage{
		Type:     string(DHTMessageTypeFindNodeRes),
		SenderID: string(s.dht.nodeID),
		TargetID: msg.SenderID,
		Data: FindNodeResponse{
			Peers: peers,
		},
		Timestamp: time.Now().Unix(),
	}

	return &response, nil
}

// handleStore обрабатывает store сообщение
func (s *DHTServer) handleStore(msg DHTMessage) (*DHTMessage, error) {
	slog.Debug("Received store", "from", msg.SenderID)

	// Парсим данные запроса
	var storeData StoreData
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Data)), &storeData); err != nil {
		return nil, fmt.Errorf("failed to parse store data: %w", err)
	}

	// Здесь должна быть логика хранения данных
	// Пока просто логируем
	slog.Debug("Storing data", "key", storeData.Key, "ttl", storeData.TTL)

	// Создаем ответ
	response := DHTMessage{
		Type:      string(DHTMessageTypeStoreRes),
		SenderID:  string(s.dht.nodeID),
		TargetID:  msg.SenderID,
		Data:      map[string]string{"status": "stored"},
		Timestamp: time.Now().Unix(),
	}

	return &response, nil
}

// handleGetValue обрабатывает get_value сообщение
func (s *DHTServer) handleGetValue(msg DHTMessage) (*DHTMessage, error) {
	slog.Debug("Received get_value", "from", msg.SenderID)

	// Парсим данные запроса
	var getData GetValueData
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Data)), &getData); err != nil {
		return nil, fmt.Errorf("failed to parse get_value data: %w", err)
	}

	// Здесь должна быть логика получения данных
	// Пока возвращаем "не найдено"
	found := false
	var value interface{} = nil

	// Создаем ответ
	response := DHTMessage{
		Type:     string(DHTMessageTypeGetValueRes),
		SenderID: string(s.dht.nodeID),
		TargetID: msg.SenderID,
		Data: GetValueResponse{
			Value: value,
			Found: found,
		},
		Timestamp: time.Now().Unix(),
	}

	return &response, nil
}

// sendMessage отправляет сообщение через соединение
func (s *DHTServer) sendMessage(conn net.Conn, msg DHTMessage) error {
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
func (s *DHTServer) receiveMessage(conn net.Conn, msg *DHTMessage) error {
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
