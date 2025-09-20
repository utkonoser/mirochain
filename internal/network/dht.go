package network

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"math/big"
	"sort"
	"sync"
	"time"
)

// DHT представляет Distributed Hash Table для peer discovery
type DHT struct {
	nodeID     []byte
	bucketSize int
	buckets    []*Bucket
	peers      map[string]*PeerInfo
	peersMux   sync.RWMutex
	bootstrap  []string
	server     *Server
}

// Bucket представляет корзину в DHT
type Bucket struct {
	peers []*PeerInfo
	mutex sync.RWMutex
}

// PeerInfo содержит информацию о peer'е в DHT
type PeerInfo struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	LastSeen  time.Time `json:"last_seen"`
	Distance  *big.Int  `json:"distance"`
	PublicKey []byte    `json:"public_key"`
}

// DHTMessage представляет сообщение DHT
type DHTMessage struct {
	Type      string      `json:"type"`
	SenderID  string      `json:"sender_id"`
	TargetID  string      `json:"target_id"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// DHTMessageType определяет тип сообщения DHT
type DHTMessageType string

const (
	DHTMessageTypePing        DHTMessageType = "ping"
	DHTMessageTypePong        DHTMessageType = "pong"
	DHTMessageTypeFindNode    DHTMessageType = "find_node"
	DHTMessageTypeFindNodeRes DHTMessageType = "find_node_res"
	DHTMessageTypeStore       DHTMessageType = "store"
	DHTMessageTypeStoreRes    DHTMessageType = "store_res"
	DHTMessageTypeGetValue    DHTMessageType = "get_value"
	DHTMessageTypeGetValueRes DHTMessageType = "get_value_res"
)

// FindNodeData содержит данные для поиска узла
type FindNodeData struct {
	TargetID string `json:"target_id"`
	Count    int    `json:"count"`
}

// FindNodeResponse содержит ответ на поиск узла
type FindNodeResponse struct {
	Peers []*PeerInfo `json:"peers"`
}

// StoreData содержит данные для хранения
type StoreData struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   int64       `json:"ttl"`
}

// GetValueData содержит данные для получения значения
type GetValueData struct {
	Key string `json:"key"`
}

// GetValueResponse содержит ответ на получение значения
type GetValueResponse struct {
	Value interface{} `json:"value"`
	Found bool        `json:"found"`
}

// NewDHT создает новый DHT
func NewDHT(nodeID []byte, bucketSize int, server *Server) *DHT {
	dht := &DHT{
		nodeID:     nodeID,
		bucketSize: bucketSize,
		buckets:    make([]*Bucket, 256), // 256 бит для SHA256
		peers:      make(map[string]*PeerInfo),
		bootstrap:  []string{},
		server:     server,
	}

	// Инициализируем корзины
	for i := 0; i < 256; i++ {
		dht.buckets[i] = &Bucket{
			peers: make([]*PeerInfo, 0, bucketSize),
		}
	}

	return dht
}

// AddBootstrapNode добавляет bootstrap узел
func (dht *DHT) AddBootstrapNode(address string) {
	dht.bootstrap = append(dht.bootstrap, address)
}

// Start запускает DHT
func (dht *DHT) Start() error {
	slog.Info("Starting DHT", "node_id", fmt.Sprintf("%x", dht.nodeID))

	// Подключаемся к bootstrap узлам
	go dht.connectToBootstrapNodes()

	// Запускаем периодическую очистку
	go dht.cleanupRoutine()

	return nil
}

// Stop останавливает DHT
func (dht *DHT) Stop() error {
	slog.Info("Stopping DHT")
	return nil
}

// AddPeer добавляет peer в DHT
func (dht *DHT) AddPeer(peer *PeerInfo) {
	dht.peersMux.Lock()
	defer dht.peersMux.Unlock()

	// Вычисляем расстояние
	distance := dht.calculateDistance(dht.nodeID, []byte(peer.ID))
	peer.Distance = distance

	// Добавляем в соответствующую корзину
	bucketIndex := dht.getBucketIndex(distance)
	bucket := dht.buckets[bucketIndex]

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	// Проверяем, есть ли уже такой peer
	for i, existingPeer := range bucket.peers {
		if existingPeer.ID == peer.ID {
			// Обновляем информацию
			bucket.peers[i] = peer
			dht.peers[peer.ID] = peer
			return
		}
	}

	// Если корзина не полная, добавляем peer
	if len(bucket.peers) < dht.bucketSize {
		bucket.peers = append(bucket.peers, peer)
		dht.peers[peer.ID] = peer
	} else {
		// Заменяем самого старого peer'а
		oldestIndex := 0
		oldestTime := bucket.peers[0].LastSeen
		for i, p := range bucket.peers {
			if p.LastSeen.Before(oldestTime) {
				oldestTime = p.LastSeen
				oldestIndex = i
			}
		}
		bucket.peers[oldestIndex] = peer
		dht.peers[peer.ID] = peer
	}

	slog.Debug("Added peer to DHT", "peer_id", peer.ID, "bucket", bucketIndex)
}

// FindNode ищет ближайших peer'ов к targetID
func (dht *DHT) FindNode(targetID string) []*PeerInfo {
	dht.peersMux.RLock()
	defer dht.peersMux.RUnlock()

	targetBytes := []byte(targetID)
	var candidates []*PeerInfo

	// Собираем всех peer'ов
	for _, peer := range dht.peers {
		candidates = append(candidates, peer)
	}

	// Сортируем по расстоянию
	sort.Slice(candidates, func(i, j int) bool {
		distI := dht.calculateDistance(targetBytes, []byte(candidates[i].ID))
		distJ := dht.calculateDistance(targetBytes, []byte(candidates[j].ID))
		return distI.Cmp(distJ) < 0
	})

	// Возвращаем ближайших
	maxCount := 8 // Kademlia K=8
	if len(candidates) < maxCount {
		maxCount = len(candidates)
	}

	return candidates[:maxCount]
}

// GetPeer возвращает информацию о peer'е
func (dht *DHT) GetPeer(peerID string) (*PeerInfo, bool) {
	dht.peersMux.RLock()
	defer dht.peersMux.RUnlock()

	peer, exists := dht.peers[peerID]
	return peer, exists
}

// GetAllPeers возвращает всех peer'ов
func (dht *DHT) GetAllPeers() []*PeerInfo {
	dht.peersMux.RLock()
	defer dht.peersMux.RUnlock()

	var peers []*PeerInfo
	for _, peer := range dht.peers {
		peers = append(peers, peer)
	}
	return peers
}

// GetPeerCount возвращает количество peer'ов
func (dht *DHT) GetPeerCount() int {
	dht.peersMux.RLock()
	defer dht.peersMux.RUnlock()

	return len(dht.peers)
}

// calculateDistance вычисляет XOR расстояние между двумя ID
func (dht *DHT) calculateDistance(id1, id2 []byte) *big.Int {
	// Дополняем до одинаковой длины
	maxLen := len(id1)
	if len(id2) > maxLen {
		maxLen = len(id2)
	}

	// Создаем массивы одинаковой длины
	bytes1 := make([]byte, maxLen)
	bytes2 := make([]byte, maxLen)
	copy(bytes1[maxLen-len(id1):], id1)
	copy(bytes2[maxLen-len(id2):], id2)

	// Вычисляем XOR
	result := make([]byte, maxLen)
	for i := 0; i < maxLen; i++ {
		result[i] = bytes1[i] ^ bytes2[i]
	}

	return new(big.Int).SetBytes(result)
}

// getBucketIndex возвращает индекс корзины для расстояния
func (dht *DHT) getBucketIndex(distance *big.Int) int {
	// Находим позицию первого ненулевого бита
	bytes := distance.Bytes()
	if len(bytes) == 0 {
		return 255 // Максимальная корзина для нулевого расстояния
	}

	// Ищем первый ненулевой байт
	for i, b := range bytes {
		if b != 0 {
			// Находим позицию первого ненулевого бита в байте
			bitPos := 0
			for j := 7; j >= 0; j-- {
				if b&(1<<j) != 0 {
					bitPos = j
					break
				}
			}
			return 255 - (i*8 + (7 - bitPos))
		}
	}

	return 0
}

// connectToBootstrapNodes подключается к bootstrap узлам
func (dht *DHT) connectToBootstrapNodes() {
	for _, bootstrapAddr := range dht.bootstrap {
		go func(addr string) {
			// Здесь должна быть логика подключения к bootstrap узлу
			// Пока просто логируем
			slog.Info("Connecting to bootstrap node", "address", addr)
		}(bootstrapAddr)
	}
}

// cleanupRoutine периодически очищает старые peer'ы
func (dht *DHT) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		dht.cleanupOldPeers()
	}
}

// cleanupOldPeers удаляет старых peer'ов
func (dht *DHT) cleanupOldPeers() {
	dht.peersMux.Lock()
	defer dht.peersMux.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)

	for _, bucket := range dht.buckets {
		bucket.mutex.Lock()
		var activePeers []*PeerInfo

		for _, peer := range bucket.peers {
			if peer.LastSeen.After(cutoff) {
				activePeers = append(activePeers, peer)
			} else {
				delete(dht.peers, peer.ID)
			}
		}

		bucket.peers = activePeers
		bucket.mutex.Unlock()
	}
}

// generateNodeID генерирует ID узла для DHT
func GenerateDHTNodeID() []byte {
	// Используем текущее время + случайные данные
	data := fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// String возвращает строковое представление DHT
func (dht *DHT) String() string {
	return fmt.Sprintf("DHT{NodeID: %x, Peers: %d, Buckets: %d}",
		dht.nodeID, dht.GetPeerCount(), len(dht.buckets))
}
