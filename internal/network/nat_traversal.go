package network

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// NATType определяет тип NAT
type NATType string

const (
	NATNone               NATType = "none"                 // Прямое подключение
	NATFullCone           NATType = "full_cone"            // Полный конус
	NATRestrictedCone     NATType = "restricted_cone"      // Ограниченный конус
	NATPortRestrictedCone NATType = "port_restricted_cone" // Порт-ограниченный конус
	NATSymmetric          NATType = "symmetric"            // Симметричный
	NATUnknown            NATType = "unknown"              // Неизвестный
)

// NATTraversalConfig конфигурация для NAT traversal
type NATTraversalConfig struct {
	STUNServers       []string      `json:"stun_servers"`        // Список STUN серверов
	STUNTimeout       time.Duration `json:"stun_timeout"`        // Таймаут STUN запросов
	HolePunchTimeout  time.Duration `json:"hole_punch_timeout"`  // Таймаут hole punching
	KeepAliveInterval time.Duration `json:"keep_alive_interval"` // Интервал keep-alive
	MaxRetries        int           `json:"max_retries"`         // Максимальное количество повторов
}

// NATInfo содержит информацию о NAT
type NATInfo struct {
	Type         NATType   `json:"type"`
	ExternalIP   string    `json:"external_ip"`
	ExternalPort int       `json:"external_port"`
	InternalIP   string    `json:"internal_ip"`
	InternalPort int       `json:"internal_port"`
	MappedPort   int       `json:"mapped_port"`
	IsBehindNAT  bool      `json:"is_behind_nat"`
	LastChecked  time.Time `json:"last_checked"`
}

// STUNClient реализует STUN клиент
type STUNClient struct {
	config NATTraversalConfig
	conn   *net.UDPConn
}

// NewSTUNClient создает новый STUN клиент
func NewSTUNClient(config NATTraversalConfig) *STUNClient {
	return &STUNClient{
		config: config,
	}
}

// DiscoverNATType определяет тип NAT
func (c *STUNClient) DiscoverNATType(localPort int) (*NATInfo, error) {
	// Создаем UDP соединение
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: localPort})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	c.conn = conn

	// Получаем локальный адрес
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// Пробуем определить тип NAT
	natInfo := &NATInfo{
		InternalIP:   localAddr.IP.String(),
		InternalPort: localAddr.Port,
		IsBehindNAT:  false,
		LastChecked:  time.Now(),
	}

	// Тестируем с разными STUN серверами
	for _, stunServer := range c.config.STUNServers {
		info, err := c.testSTUNServer(stunServer)
		if err != nil {
			log.Printf("STUN server %s failed: %v", stunServer, err)
			continue
		}

		// Если получили внешний IP, значит за NAT
		if info.ExternalIP != "" && info.ExternalIP != localAddr.IP.String() {
			natInfo.IsBehindNAT = true
			natInfo.ExternalIP = info.ExternalIP
			natInfo.ExternalPort = info.ExternalPort
			natInfo.MappedPort = info.MappedPort
			natInfo.Type = c.determineNATType(info)
			break
		}
	}

	// Если не за NAT, определяем как прямой доступ
	if !natInfo.IsBehindNAT {
		natInfo.Type = NATNone
		natInfo.ExternalIP = localAddr.IP.String()
		natInfo.ExternalPort = localAddr.Port
		natInfo.MappedPort = localAddr.Port
	}

	return natInfo, nil
}

// testSTUNServer тестирует STUN сервер
func (c *STUNClient) testSTUNServer(stunServer string) (*NATInfo, error) {
	// Упрощенная реализация STUN
	// В реальной реализации здесь должен быть полный STUN протокол

	// Парсим адрес STUN сервера
	addr, err := net.ResolveUDPAddr("udp", stunServer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve STUN server: %v", err)
	}

	// Отправляем тестовый пакет
	testData := []byte("STUN_TEST")
	_, err = c.conn.WriteToUDP(testData, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %v", err)
	}

	// Устанавливаем таймаут
	c.conn.SetReadDeadline(time.Now().Add(c.config.STUNTimeout))

	// Читаем ответ
	buffer := make([]byte, 1024)
	n, remoteAddr, err := c.conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %v", err)
	}

	// В реальной реализации здесь должен быть парсинг STUN ответа
	// Пока что просто возвращаем информацию о соединении
	log.Printf("STUN response from %s: %d bytes", remoteAddr, n)

	return &NATInfo{
		ExternalIP:   remoteAddr.IP.String(),
		ExternalPort: remoteAddr.Port,
		MappedPort:   c.conn.LocalAddr().(*net.UDPAddr).Port,
	}, nil
}

// determineNATType определяет тип NAT на основе тестов
func (c *STUNClient) determineNATType(info *NATInfo) NATType {
	// Упрощенная логика определения типа NAT
	// В реальной реализации здесь должны быть более сложные тесты

	if info.ExternalIP == "" {
		return NATUnknown
	}

	// Если внешний IP отличается от внутреннего, скорее всего за NAT
	if info.ExternalIP != info.InternalIP {
		// Простая эвристика: если порт изменился, скорее всего симметричный NAT
		if info.ExternalPort != info.InternalPort {
			return NATSymmetric
		}
		return NATFullCone
	}

	return NATNone
}

// NATTraversal реализует NAT traversal
type NATTraversal struct {
	config     NATTraversalConfig
	stunClient *STUNClient
	natInfo    *NATInfo
	peers      map[string]*NATPeer
	peersMux   sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NATPeer представляет peer для NAT traversal
type NATPeer struct {
	ID           string    `json:"id"`
	InternalAddr string    `json:"internal_addr"`
	ExternalAddr string    `json:"external_addr"`
	NATType      NATType   `json:"nat_type"`
	LastSeen     time.Time `json:"last_seen"`
	IsReachable  bool      `json:"is_reachable"`
}

// NewNATTraversal создает новый NAT traversal
func NewNATTraversal(config NATTraversalConfig) *NATTraversal {
	ctx, cancel := context.WithCancel(context.Background())

	return &NATTraversal{
		config:     config,
		stunClient: NewSTUNClient(config),
		peers:      make(map[string]*NATPeer),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start запускает NAT traversal
func (n *NATTraversal) Start(localPort int) error {
	log.Printf("Starting NAT traversal on port %d", localPort)

	// Определяем тип NAT
	natInfo, err := n.stunClient.DiscoverNATType(localPort)
	if err != nil {
		return fmt.Errorf("failed to discover NAT type: %v", err)
	}

	n.natInfo = natInfo
	log.Printf("NAT type: %s, External IP: %s:%d", natInfo.Type, natInfo.ExternalIP, natInfo.ExternalPort)

	// Запускаем keep-alive для поддержания соединений
	go n.keepAliveLoop()

	return nil
}

// Stop останавливает NAT traversal
func (n *NATTraversal) Stop() {
	log.Printf("Stopping NAT traversal")
	n.cancel()
}

// AddPeer добавляет peer для NAT traversal
func (n *NATTraversal) AddPeer(peerID, internalAddr, externalAddr string, natType NATType) {
	n.peersMux.Lock()
	defer n.peersMux.Unlock()

	n.peers[peerID] = &NATPeer{
		ID:           peerID,
		InternalAddr: internalAddr,
		ExternalAddr: externalAddr,
		NATType:      natType,
		LastSeen:     time.Now(),
		IsReachable:  false,
	}

	log.Printf("Added peer %s for NAT traversal", peerID)
}

// RemovePeer удаляет peer
func (n *NATTraversal) RemovePeer(peerID string) {
	n.peersMux.Lock()
	defer n.peersMux.Unlock()

	delete(n.peers, peerID)
	log.Printf("Removed peer %s from NAT traversal", peerID)
}

// EstablishConnection устанавливает соединение с peer
func (n *NATTraversal) EstablishConnection(peerID string) error {
	n.peersMux.RLock()
	peer, exists := n.peers[peerID]
	n.peersMux.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not found", peerID)
	}

	log.Printf("Establishing connection to peer %s", peerID)

	// Выбираем стратегию в зависимости от типов NAT
	strategy := n.selectStrategy(n.natInfo.Type, peer.NATType)

	switch strategy {
	case "direct":
		return n.directConnection(peer)
	case "hole_punching":
		return n.holePunching(peer)
	case "relay":
		return n.relayConnection(peer)
	default:
		return fmt.Errorf("unsupported strategy: %s", strategy)
	}
}

// selectStrategy выбирает стратегию соединения
func (n *NATTraversal) selectStrategy(localNAT, peerNAT NATType) string {
	// Упрощенная логика выбора стратегии
	if localNAT == NATNone && peerNAT == NATNone {
		return "direct"
	}

	if (localNAT == NATFullCone || localNAT == NATRestrictedCone) &&
		(peerNAT == NATFullCone || peerNAT == NATRestrictedCone) {
		return "hole_punching"
	}

	return "relay"
}

// directConnection устанавливает прямое соединение
func (n *NATTraversal) directConnection(peer *NATPeer) error {
	log.Printf("Attempting direct connection to %s", peer.InternalAddr)

	// Пытаемся подключиться напрямую
	conn, err := net.Dial("udp", peer.InternalAddr)
	if err != nil {
		return fmt.Errorf("direct connection failed: %v", err)
	}
	defer conn.Close()

	// Отправляем тестовый пакет
	_, err = conn.Write([]byte("NAT_TEST"))
	if err != nil {
		return fmt.Errorf("failed to send test packet: %v", err)
	}

	peer.IsReachable = true
	peer.LastSeen = time.Now()

	log.Printf("Direct connection to %s established", peer.ID)
	return nil
}

// holePunching выполняет hole punching
func (n *NATTraversal) holePunching(peer *NATPeer) error {
	log.Printf("Attempting hole punching to %s", peer.ExternalAddr)

	// Создаем UDP соединение
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	// Отправляем пакеты для создания "дыры" в NAT
	for i := 0; i < n.config.MaxRetries; i++ {
		_, err = conn.WriteToUDP([]byte("HOLE_PUNCH"),
			&net.UDPAddr{IP: net.ParseIP(peer.ExternalAddr), Port: 8080})
		if err != nil {
			log.Printf("Hole punch attempt %d failed: %v", i+1, err)
			time.Sleep(time.Second)
			continue
		}

		// Пытаемся прочитать ответ
		conn.SetReadDeadline(time.Now().Add(n.config.HolePunchTimeout))
		buffer := make([]byte, 1024)
		_, _, err := conn.ReadFromUDP(buffer)
		if err == nil {
			peer.IsReachable = true
			peer.LastSeen = time.Now()
			log.Printf("Hole punching to %s successful", peer.ID)
			return nil
		}
	}

	return fmt.Errorf("hole punching failed after %d attempts", n.config.MaxRetries)
}

// relayConnection устанавливает соединение через relay
func (n *NATTraversal) relayConnection(peer *NATPeer) error {
	log.Printf("Attempting relay connection to %s", peer.ID)

	// В реальной реализации здесь должен быть поиск и подключение к relay серверу
	// Пока что просто логируем
	log.Printf("Relay connection to %s not implemented yet", peer.ID)

	return fmt.Errorf("relay connection not implemented")
}

// keepAliveLoop поддерживает соединения
func (n *NATTraversal) keepAliveLoop() {
	ticker := time.NewTicker(n.config.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.sendKeepAlive()
		}
	}
}

// sendKeepAlive отправляет keep-alive пакеты
func (n *NATTraversal) sendKeepAlive() {
	n.peersMux.RLock()
	defer n.peersMux.RUnlock()

	for _, peer := range n.peers {
		if peer.IsReachable {
			// Отправляем keep-alive пакет
			log.Printf("Sending keep-alive to %s", peer.ID)
			// В реальной реализации здесь должна быть отправка пакета
		}
	}
}

// GetNATInfo возвращает информацию о NAT
func (n *NATTraversal) GetNATInfo() *NATInfo {
	return n.natInfo
}

// GetPeers возвращает список peer'ов
func (n *NATTraversal) GetPeers() []*NATPeer {
	n.peersMux.RLock()
	defer n.peersMux.RUnlock()

	peers := make([]*NATPeer, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}

	return peers
}

// GetStats возвращает статистику NAT traversal
func (n *NATTraversal) GetStats() map[string]interface{} {
	n.peersMux.RLock()
	defer n.peersMux.RUnlock()

	reachableCount := 0
	for _, peer := range n.peers {
		if peer.IsReachable {
			reachableCount++
		}
	}

	return map[string]interface{}{
		"nat_type":        n.natInfo.Type,
		"external_ip":     n.natInfo.ExternalIP,
		"external_port":   n.natInfo.ExternalPort,
		"is_behind_nat":   n.natInfo.IsBehindNAT,
		"total_peers":     len(n.peers),
		"reachable_peers": reachableCount,
	}
}
