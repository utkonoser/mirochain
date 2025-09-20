package network

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// RateLimiterType определяет тип rate limiter'а
type RateLimiterType string

const (
	TokenBucket   RateLimiterType = "token_bucket"
	SlidingWindow RateLimiterType = "sliding_window"
	FixedWindow   RateLimiterType = "fixed_window"
)

// RateLimiterConfig конфигурация для rate limiter'а
type RateLimiterConfig struct {
	Type        RateLimiterType `json:"type"`
	MaxRequests int             `json:"max_requests"` // Максимальное количество запросов
	WindowSize  time.Duration   `json:"window_size"`  // Размер окна времени
	BurstSize   int             `json:"burst_size"`   // Размер burst (для token bucket)
	RefillRate  int             `json:"refill_rate"`  // Скорость пополнения (для token bucket)
}

// RateLimiter интерфейс для rate limiting
type RateLimiter interface {
	Allow(clientID string) bool
	GetStats(clientID string) map[string]interface{}
	Reset(clientID string)
	GetAllStats() map[string]interface{}
}

// TokenBucketRateLimiter реализует token bucket алгоритм
type TokenBucketRateLimiter struct {
	config     RateLimiterConfig
	buckets    map[string]*TokenBucketData
	bucketsMux sync.RWMutex
}

// TokenBucketData представляет bucket с токенами
type TokenBucketData struct {
	Tokens     int       `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
	MaxTokens  int       `json:"max_tokens"`
	RefillRate int       `json:"refill_rate"`
}

// NewTokenBucketRateLimiter создает новый token bucket rate limiter
func NewTokenBucketRateLimiter(config RateLimiterConfig) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		config:  config,
		buckets: make(map[string]*TokenBucketData),
	}
}

// Allow проверяет, разрешен ли запрос
func (r *TokenBucketRateLimiter) Allow(clientID string) bool {
	r.bucketsMux.Lock()
	defer r.bucketsMux.Unlock()

	bucket, exists := r.buckets[clientID]
	if !exists {
		bucket = &TokenBucketData{
			Tokens:     r.config.BurstSize,
			LastRefill: time.Now(),
			MaxTokens:  r.config.BurstSize,
			RefillRate: r.config.RefillRate,
		}
		r.buckets[clientID] = bucket
	}

	// Пополняем токены
	r.refillBucket(bucket)

	if bucket.Tokens > 0 {
		bucket.Tokens--
		return true
	}

	return false
}

// refillBucket пополняет bucket токенами
func (r *TokenBucketRateLimiter) refillBucket(bucket *TokenBucketData) {
	now := time.Now()
	timePassed := now.Sub(bucket.LastRefill)

	// Вычисляем количество токенов для пополнения
	tokensToAdd := int(timePassed.Seconds()) * bucket.RefillRate
	if tokensToAdd > 0 {
		bucket.Tokens += tokensToAdd
		if bucket.Tokens > bucket.MaxTokens {
			bucket.Tokens = bucket.MaxTokens
		}
		bucket.LastRefill = now
	}
}

// GetStats возвращает статистику для клиента
func (r *TokenBucketRateLimiter) GetStats(clientID string) map[string]interface{} {
	r.bucketsMux.RLock()
	defer r.bucketsMux.RUnlock()

	bucket, exists := r.buckets[clientID]
	if !exists {
		return map[string]interface{}{
			"tokens":      r.config.BurstSize,
			"max_tokens":  r.config.BurstSize,
			"refill_rate": r.config.RefillRate,
		}
	}

	return map[string]interface{}{
		"tokens":      bucket.Tokens,
		"max_tokens":  bucket.MaxTokens,
		"refill_rate": bucket.RefillRate,
		"last_refill": bucket.LastRefill,
	}
}

// Reset сбрасывает статистику клиента
func (r *TokenBucketRateLimiter) Reset(clientID string) {
	r.bucketsMux.Lock()
	defer r.bucketsMux.Unlock()

	delete(r.buckets, clientID)
}

// GetAllStats возвращает общую статистику
func (r *TokenBucketRateLimiter) GetAllStats() map[string]interface{} {
	r.bucketsMux.RLock()
	defer r.bucketsMux.RUnlock()

	return map[string]interface{}{
		"type":          r.config.Type,
		"total_clients": len(r.buckets),
		"max_requests":  r.config.MaxRequests,
		"window_size":   r.config.WindowSize,
		"burst_size":    r.config.BurstSize,
		"refill_rate":   r.config.RefillRate,
	}
}

// SlidingWindowRateLimiter реализует sliding window алгоритм
type SlidingWindowRateLimiter struct {
	config     RateLimiterConfig
	windows    map[string]*SlidingWindowData
	windowsMux sync.RWMutex
}

// SlidingWindowData представляет sliding window
type SlidingWindowData struct {
	Requests []time.Time   `json:"requests"`
	MaxSize  int           `json:"max_size"`
	Window   time.Duration `json:"window"`
}

// NewSlidingWindowRateLimiter создает новый sliding window rate limiter
func NewSlidingWindowRateLimiter(config RateLimiterConfig) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		config:  config,
		windows: make(map[string]*SlidingWindowData),
	}
}

// Allow проверяет, разрешен ли запрос
func (r *SlidingWindowRateLimiter) Allow(clientID string) bool {
	r.windowsMux.Lock()
	defer r.windowsMux.Unlock()

	window, exists := r.windows[clientID]
	if !exists {
		window = &SlidingWindowData{
			Requests: make([]time.Time, 0),
			MaxSize:  r.config.MaxRequests,
			Window:   r.config.WindowSize,
		}
		r.windows[clientID] = window
	}

	now := time.Now()
	cutoff := now.Add(-window.Window)

	// Удаляем старые запросы
	newRequests := make([]time.Time, 0)
	for _, reqTime := range window.Requests {
		if reqTime.After(cutoff) {
			newRequests = append(newRequests, reqTime)
		}
	}
	window.Requests = newRequests

	// Проверяем, есть ли место для нового запроса
	if len(window.Requests) < window.MaxSize {
		window.Requests = append(window.Requests, now)
		return true
	}

	return false
}

// GetStats возвращает статистику для клиента
func (r *SlidingWindowRateLimiter) GetStats(clientID string) map[string]interface{} {
	r.windowsMux.RLock()
	defer r.windowsMux.RUnlock()

	window, exists := r.windows[clientID]
	if !exists {
		return map[string]interface{}{
			"requests":     0,
			"max_requests": r.config.MaxRequests,
			"window_size":  r.config.WindowSize,
		}
	}

	return map[string]interface{}{
		"requests":     len(window.Requests),
		"max_requests": window.MaxSize,
		"window_size":  window.Window,
	}
}

// Reset сбрасывает статистику клиента
func (r *SlidingWindowRateLimiter) Reset(clientID string) {
	r.windowsMux.Lock()
	defer r.windowsMux.Unlock()

	delete(r.windows, clientID)
}

// GetAllStats возвращает общую статистику
func (r *SlidingWindowRateLimiter) GetAllStats() map[string]interface{} {
	r.windowsMux.RLock()
	defer r.windowsMux.RUnlock()

	return map[string]interface{}{
		"type":          r.config.Type,
		"total_clients": len(r.windows),
		"max_requests":  r.config.MaxRequests,
		"window_size":   r.config.WindowSize,
	}
}

// FixedWindowRateLimiter реализует fixed window алгоритм
type FixedWindowRateLimiter struct {
	config     RateLimiterConfig
	windows    map[string]*FixedWindowData
	windowsMux sync.RWMutex
}

// FixedWindowData представляет fixed window
type FixedWindowData struct {
	Count       int           `json:"count"`
	WindowStart time.Time     `json:"window_start"`
	MaxCount    int           `json:"max_count"`
	Window      time.Duration `json:"window"`
}

// NewFixedWindowRateLimiter создает новый fixed window rate limiter
func NewFixedWindowRateLimiter(config RateLimiterConfig) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		config:  config,
		windows: make(map[string]*FixedWindowData),
	}
}

// Allow проверяет, разрешен ли запрос
func (r *FixedWindowRateLimiter) Allow(clientID string) bool {
	r.windowsMux.Lock()
	defer r.windowsMux.Unlock()

	window, exists := r.windows[clientID]
	if !exists {
		window = &FixedWindowData{
			Count:       0,
			WindowStart: time.Now(),
			MaxCount:    r.config.MaxRequests,
			Window:      r.config.WindowSize,
		}
		r.windows[clientID] = window
	}

	now := time.Now()

	// Проверяем, нужно ли сбросить окно
	if now.Sub(window.WindowStart) >= window.Window {
		window.Count = 0
		window.WindowStart = now
	}

	// Проверяем, есть ли место для нового запроса
	if window.Count < window.MaxCount {
		window.Count++
		return true
	}

	return false
}

// GetStats возвращает статистику для клиента
func (r *FixedWindowRateLimiter) GetStats(clientID string) map[string]interface{} {
	r.windowsMux.RLock()
	defer r.windowsMux.RUnlock()

	window, exists := r.windows[clientID]
	if !exists {
		return map[string]interface{}{
			"count":       0,
			"max_count":   r.config.MaxRequests,
			"window_size": r.config.WindowSize,
		}
	}

	return map[string]interface{}{
		"count":        window.Count,
		"max_count":    window.MaxCount,
		"window_size":  window.Window,
		"window_start": window.WindowStart,
	}
}

// Reset сбрасывает статистику клиента
func (r *FixedWindowRateLimiter) Reset(clientID string) {
	r.windowsMux.Lock()
	defer r.windowsMux.Unlock()

	delete(r.windows, clientID)
}

// GetAllStats возвращает общую статистику
func (r *FixedWindowRateLimiter) GetAllStats() map[string]interface{} {
	r.windowsMux.RLock()
	defer r.windowsMux.RUnlock()

	return map[string]interface{}{
		"type":          r.config.Type,
		"total_clients": len(r.windows),
		"max_requests":  r.config.MaxRequests,
		"window_size":   r.config.WindowSize,
	}
}

// RateLimiterManager управляет rate limiter'ами
type RateLimiterManager struct {
	limiters map[string]RateLimiter
	configs  map[string]RateLimiterConfig
	mux      sync.RWMutex
}

// NewRateLimiterManager создает новый менеджер rate limiter'ов
func NewRateLimiterManager() *RateLimiterManager {
	return &RateLimiterManager{
		limiters: make(map[string]RateLimiter),
		configs:  make(map[string]RateLimiterConfig),
	}
}

// AddRateLimiter добавляет rate limiter
func (m *RateLimiterManager) AddRateLimiter(name string, config RateLimiterConfig) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	var limiter RateLimiter

	switch config.Type {
	case TokenBucket:
		limiter = NewTokenBucketRateLimiter(config)
	case SlidingWindow:
		limiter = NewSlidingWindowRateLimiter(config)
	case FixedWindow:
		limiter = NewFixedWindowRateLimiter(config)
	default:
		return fmt.Errorf("unsupported rate limiter type: %s", config.Type)
	}

	m.limiters[name] = limiter
	m.configs[name] = config

	log.Printf("Added rate limiter '%s' with type %s", name, config.Type)
	return nil
}

// Allow проверяет, разрешен ли запрос для указанного limiter'а
func (m *RateLimiterManager) Allow(limiterName, clientID string) bool {
	m.mux.RLock()
	limiter, exists := m.limiters[limiterName]
	m.mux.RUnlock()

	if !exists {
		log.Printf("Rate limiter '%s' not found", limiterName)
		return true // Разрешаем, если limiter не найден
	}

	return limiter.Allow(clientID)
}

// GetStats возвращает статистику для limiter'а
func (m *RateLimiterManager) GetStats(limiterName, clientID string) map[string]interface{} {
	m.mux.RLock()
	limiter, exists := m.limiters[limiterName]
	m.mux.RUnlock()

	if !exists {
		return map[string]interface{}{"error": "limiter not found"}
	}

	return limiter.GetStats(clientID)
}

// GetAllStats возвращает общую статистику
func (m *RateLimiterManager) GetAllStats() map[string]interface{} {
	m.mux.RLock()
	defer m.mux.RUnlock()

	stats := make(map[string]interface{})
	for name, limiter := range m.limiters {
		stats[name] = limiter.GetAllStats()
	}

	return stats
}

// Reset сбрасывает статистику для limiter'а
func (m *RateLimiterManager) Reset(limiterName, clientID string) {
	m.mux.RLock()
	limiter, exists := m.limiters[limiterName]
	m.mux.RUnlock()

	if exists {
		limiter.Reset(clientID)
	}
}

// GetLimiterNames возвращает список имен limiter'ов
func (m *RateLimiterManager) GetLimiterNames() []string {
	m.mux.RLock()
	defer m.mux.RUnlock()

	names := make([]string, 0, len(m.limiters))
	for name := range m.limiters {
		names = append(names, name)
	}

	return names
}
