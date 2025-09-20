package cache

import (
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// Cache представляет интерфейс для кэширования
type Cache interface {
	// Блоки
	GetBlock(hash []byte) (*blockchain.Block, bool)
	SetBlock(hash []byte, block *blockchain.Block)
	GetBlockByHeight(height int64) (*blockchain.Block, bool)
	SetBlockByHeight(height int64, block *blockchain.Block)

	// UTXO
	GetUTXOs(address string) ([]*blockchain.UTXO, bool)
	SetUTXOs(address string, utxos []*blockchain.UTXO)
	GetBalance(address string) (int64, bool)
	SetBalance(address string, balance int64)

	// Метаданные
	GetHeight() (int64, bool)
	SetHeight(height int64)
	GetDifficulty() (int, bool)
	SetDifficulty(difficulty int)

	// Управление
	Clear()
	Stats() CacheStats
}

// CacheStats представляет статистику кэша
type CacheStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Size   int   `json:"size"`
}

// LRUCache реализует кэш с алгоритмом LRU (Least Recently Used)
type LRUCache struct {
	// Кэш блоков по хешу
	blocksByHash map[string]*blockchain.Block
	// Кэш блоков по высоте
	blocksByHeight map[int64]*blockchain.Block
	// Кэш UTXO по адресу
	utxosByAddress map[string][]*blockchain.UTXO
	// Кэш балансов по адресу
	balancesByAddress map[string]int64
	// Кэш метаданных
	height     int64
	difficulty int

	// Порядок доступа для LRU
	accessOrder []string
	heightOrder []int64

	// Настройки
	maxBlocks   int
	maxUTXOs    int
	maxBalances int
	ttl         time.Duration

	// Статистика
	hits   int64
	misses int64

	// Синхронизация
	mutex sync.RWMutex
}

// NewLRUCache создает новый LRU кэш
func NewLRUCache(maxBlocks, maxUTXOs, maxBalances int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		blocksByHash:      make(map[string]*blockchain.Block),
		blocksByHeight:    make(map[int64]*blockchain.Block),
		utxosByAddress:    make(map[string][]*blockchain.UTXO),
		balancesByAddress: make(map[string]int64),
		accessOrder:       make([]string, 0),
		heightOrder:       make([]int64, 0),
		maxBlocks:         maxBlocks,
		maxUTXOs:          maxUTXOs,
		maxBalances:       maxBalances,
		ttl:               ttl,
	}
}

// GetBlock получает блок по хешу из кэша
func (c *LRUCache) GetBlock(hash []byte) (*blockchain.Block, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := string(hash)
	block, exists := c.blocksByHash[key]
	if exists {
		c.hits++
		c.updateAccessOrder(key)
		return block, true
	}

	c.misses++
	return nil, false
}

// SetBlock сохраняет блок в кэш по хешу
func (c *LRUCache) SetBlock(hash []byte, block *blockchain.Block) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := string(hash)
	c.blocksByHash[key] = block
	c.updateAccessOrder(key)

	// Очищаем старые записи если превышен лимит
	if len(c.blocksByHash) > c.maxBlocks {
		c.evictOldest()
	}
}

// GetBlockByHeight получает блок по высоте из кэша
func (c *LRUCache) GetBlockByHeight(height int64) (*blockchain.Block, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	block, exists := c.blocksByHeight[height]
	if exists {
		c.hits++
		c.updateHeightOrder(height)
		return block, true
	}

	c.misses++
	return nil, false
}

// SetBlockByHeight сохраняет блок в кэш по высоте
func (c *LRUCache) SetBlockByHeight(height int64, block *blockchain.Block) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.blocksByHeight[height] = block
	c.updateHeightOrder(height)

	// Очищаем старые записи если превышен лимит
	if len(c.blocksByHeight) > c.maxBlocks {
		c.evictOldestHeight()
	}
}

// GetUTXOs получает UTXO по адресу из кэша
func (c *LRUCache) GetUTXOs(address string) ([]*blockchain.UTXO, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	utxos, exists := c.utxosByAddress[address]
	if exists {
		c.hits++
		return utxos, true
	}

	c.misses++
	return nil, false
}

// SetUTXOs сохраняет UTXO в кэш по адресу
func (c *LRUCache) SetUTXOs(address string, utxos []*blockchain.UTXO) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.utxosByAddress[address] = utxos

	// Очищаем старые записи если превышен лимит
	if len(c.utxosByAddress) > c.maxUTXOs {
		c.evictOldestUTXOs()
	}
}

// GetBalance получает баланс по адресу из кэша
func (c *LRUCache) GetBalance(address string) (int64, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	balance, exists := c.balancesByAddress[address]
	if exists {
		c.hits++
		return balance, true
	}

	c.misses++
	return 0, false
}

// SetBalance сохраняет баланс в кэш по адресу
func (c *LRUCache) SetBalance(address string, balance int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.balancesByAddress[address] = balance

	// Очищаем старые записи если превышен лимит
	if len(c.balancesByAddress) > c.maxBalances {
		c.evictOldestBalances()
	}
}

// GetHeight получает высоту из кэша
func (c *LRUCache) GetHeight() (int64, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	c.hits++
	return c.height, true
}

// SetHeight сохраняет высоту в кэш
func (c *LRUCache) SetHeight(height int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.height = height
}

// GetDifficulty получает сложность из кэша
func (c *LRUCache) GetDifficulty() (int, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	c.hits++
	return c.difficulty, true
}

// SetDifficulty сохраняет сложность в кэш
func (c *LRUCache) SetDifficulty(difficulty int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.difficulty = difficulty
}

// Clear очищает весь кэш
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.blocksByHash = make(map[string]*blockchain.Block)
	c.blocksByHeight = make(map[int64]*blockchain.Block)
	c.utxosByAddress = make(map[string][]*blockchain.UTXO)
	c.balancesByAddress = make(map[string]int64)
	c.accessOrder = make([]string, 0)
	c.heightOrder = make([]int64, 0)
	c.hits = 0
	c.misses = 0
}

// Stats возвращает статистику кэша
func (c *LRUCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		Hits:   c.hits,
		Misses: c.misses,
		Size:   len(c.blocksByHash) + len(c.blocksByHeight) + len(c.utxosByAddress) + len(c.balancesByAddress),
	}
}

// Вспомогательные методы для LRU

func (c *LRUCache) updateAccessOrder(key string) {
	// Удаляем ключ из текущей позиции
	for i, k := range c.accessOrder {
		if k == key {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			break
		}
	}
	// Добавляем в конец (самый недавно использованный)
	c.accessOrder = append(c.accessOrder, key)
}

func (c *LRUCache) updateHeightOrder(height int64) {
	// Удаляем высоту из текущей позиции
	for i, h := range c.heightOrder {
		if h == height {
			c.heightOrder = append(c.heightOrder[:i], c.heightOrder[i+1:]...)
			break
		}
	}
	// Добавляем в конец (самый недавно использованный)
	c.heightOrder = append(c.heightOrder, height)
}

func (c *LRUCache) evictOldest() {
	if len(c.accessOrder) > 0 {
		oldestKey := c.accessOrder[0]
		delete(c.blocksByHash, oldestKey)
		c.accessOrder = c.accessOrder[1:]
	}
}

func (c *LRUCache) evictOldestHeight() {
	if len(c.heightOrder) > 0 {
		oldestHeight := c.heightOrder[0]
		delete(c.blocksByHeight, oldestHeight)
		c.heightOrder = c.heightOrder[1:]
	}
}

func (c *LRUCache) evictOldestUTXOs() {
	// Простая эвристика - удаляем случайный элемент
	for address := range c.utxosByAddress {
		delete(c.utxosByAddress, address)
		break
	}
}

func (c *LRUCache) evictOldestBalances() {
	// Простая эвристика - удаляем случайный элемент
	for address := range c.balancesByAddress {
		delete(c.balancesByAddress, address)
		break
	}
}
