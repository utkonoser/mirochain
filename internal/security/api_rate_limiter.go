package security

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// APIRateLimiter представляет улучшенный rate limiter для API
type APIRateLimiter struct {
	limiters    map[string]*RateLimiterConfig
	globalLimit *RateLimiterConfig
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// RateLimiterConfig представляет конфигурацию rate limiter'а
type RateLimiterConfig struct {
	Type        string        `json:"type"`        // token_bucket, sliding_window, fixed_window
	Limit       int           `json:"limit"`       // максимальное количество запросов
	Window      time.Duration `json:"window"`      // окно времени
	Burst       int           `json:"burst"`       // размер буфера для token bucket
	RefillRate  int           `json:"refill_rate"` // скорость пополнения для token bucket
	BlockTime   time.Duration `json:"block_time"`  // время блокировки при превышении лимита
	Whitelist   []string      `json:"whitelist"`   // список разрешенных IP/пользователей
	Blacklist   []string      `json:"blacklist"`   // список заблокированных IP/пользователей
}

// RateLimiter представляет отдельный rate limiter
type RateLimiter struct {
	config      *RateLimiterConfig
	tokens      int
	lastRefill  time.Time
	requests    map[string]*RequestInfo
	blocked     map[string]time.Time
	mutex       sync.RWMutex
}

// RequestInfo содержит информацию о запросе
type RequestInfo struct {
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Blocked   bool      `json:"blocked"`
}

// RateLimitResult представляет результат проверки rate limit
type RateLimitResult struct {
	Allowed    bool          `json:"allowed"`
	Remaining  int           `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
	Reason     string        `json:"reason,omitempty"`
}

// NewAPIRateLimiter создает новый API rate limiter
func NewAPIRateLimiter() *APIRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	
	limiter := &APIRateLimiter{
		limiters: make(map[string]*RateLimiterConfig),
		ctx:      ctx,
		cancel:   cancel,
	}
	
	// Устанавливаем глобальные лимиты по умолчанию
	limiter.SetGlobalLimit(&RateLimiterConfig{
		Type:       "token_bucket",
		Limit:      1000,
		Window:     1 * time.Minute,
		Burst:      100,
		RefillRate: 10,
		BlockTime:  5 * time.Minute,
	})
	
	// Устанавливаем лимиты для разных типов запросов
	limiter.SetLimiter("api", &RateLimiterConfig{
		Type:       "sliding_window",
		Limit:      100,
		Window:     1 * time.Minute,
		BlockTime:  1 * time.Minute,
	})
	
	limiter.SetLimiter("mining", &RateLimiterConfig{
		Type:       "token_bucket",
		Limit:      10,
		Window:     1 * time.Minute,
		Burst:      5,
		RefillRate: 1,
		BlockTime:  10 * time.Minute,
	})
	
	limiter.SetLimiter("wallet", &RateLimiterConfig{
		Type:       "fixed_window",
		Limit:      50,
		Window:     1 * time.Hour,
		BlockTime:  1 * time.Hour,
	})
	
	limiter.SetLimiter("admin", &RateLimiterConfig{
		Type:       "token_bucket",
		Limit:      1000,
		Window:     1 * time.Minute,
		Burst:      200,
		RefillRate: 50,
		BlockTime:  1 * time.Minute,
	})
	
	return limiter
}

// SetLimiter устанавливает конфигурацию для типа запроса
func (arl *APIRateLimiter) SetLimiter(requestType string, config *RateLimiterConfig) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()
	
	arl.limiters[requestType] = config
}

// SetGlobalLimit устанавливает глобальный лимит
func (arl *APIRateLimiter) SetGlobalLimit(config *RateLimiterConfig) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()
	
	arl.globalLimit = config
}

// CheckRateLimit проверяет rate limit для запроса
func (arl *APIRateLimiter) CheckRateLimit(requestType, clientID string) *RateLimitResult {
	arl.mutex.RLock()
	config, exists := arl.limiters[requestType]
	globalConfig := arl.globalLimit
	arl.mutex.RUnlock()
	
	if !exists {
		config = globalConfig
	}
	
	// Проверяем whitelist
	if arl.isWhitelisted(clientID, config) {
		return &RateLimitResult{
			Allowed:   true,
			Remaining: config.Limit,
			ResetTime: time.Now().Add(config.Window),
		}
	}
	
	// Проверяем blacklist
	if arl.isBlacklisted(clientID, config) {
		return &RateLimitResult{
			Allowed:    false,
			Remaining:  0,
			ResetTime:  time.Now().Add(config.BlockTime),
			RetryAfter: config.BlockTime,
			Reason:     "Client is blacklisted",
		}
	}
	
	// Создаем rate limiter для клиента
	limiter := arl.getOrCreateLimiter(requestType, clientID, config)
	
	// Проверяем лимит
	allowed, remaining, resetTime, retryAfter := limiter.CheckLimit()
	
	result := &RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: resetTime,
	}
	
	if !allowed {
		result.RetryAfter = retryAfter
		result.Reason = "Rate limit exceeded"
	}
	
	return result
}

// isWhitelisted проверяет, находится ли клиент в whitelist
func (arl *APIRateLimiter) isWhitelisted(clientID string, config *RateLimiterConfig) bool {
	if config == nil {
		return false
	}
	
	for _, whitelisted := range config.Whitelist {
		if whitelisted == clientID {
			return true
		}
	}
	return false
}

// isBlacklisted проверяет, находится ли клиент в blacklist
func (arl *APIRateLimiter) isBlacklisted(clientID string, config *RateLimiterConfig) bool {
	if config == nil {
		return false
	}
	
	for _, blacklisted := range config.Blacklist {
		if blacklisted == clientID {
			return true
		}
	}
	return false
}

// getOrCreateLimiter получает или создает rate limiter для клиента
func (arl *APIRateLimiter) getOrCreateLimiter(requestType, clientID string, config *RateLimiterConfig) *RateLimiter {
	key := fmt.Sprintf("%s:%s", requestType, clientID)
	
	// Здесь должна быть реализация кэширования или хранения
	// Пока что создаем новый limiter для каждого запроса
	_ = key // Используем переменную
	return NewRateLimiter(config)
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		config:     config,
		tokens:     config.Burst,
		lastRefill: time.Now(),
		requests:   make(map[string]*RequestInfo),
		blocked:    make(map[string]time.Time),
	}
}

// CheckLimit проверяет лимит для запроса
func (rl *RateLimiter) CheckLimit() (allowed bool, remaining int, resetTime time.Time, retryAfter time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	// Проверяем, не заблокирован ли клиент
	if blockTime, exists := rl.blocked["client"]; exists {
		if now.Before(blockTime) {
			return false, 0, blockTime, blockTime.Sub(now)
		}
		delete(rl.blocked, "client")
	}
	
	switch rl.config.Type {
	case "token_bucket":
		return rl.checkTokenBucket(now)
	case "sliding_window":
		return rl.checkSlidingWindow(now)
	case "fixed_window":
		return rl.checkFixedWindow(now)
	default:
		return true, rl.config.Limit, now.Add(rl.config.Window), 0
	}
}

// checkTokenBucket проверяет лимит по алгоритму token bucket
func (rl *RateLimiter) checkTokenBucket(now time.Time) (bool, int, time.Time, time.Duration) {
	// Пополняем токены
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.config.RefillRate
	
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.config.Burst {
			rl.tokens = rl.config.Burst
		}
		rl.lastRefill = now
	}
	
	// Проверяем, есть ли токены
	if rl.tokens > 0 {
		rl.tokens--
		return true, rl.tokens, now.Add(rl.config.Window), 0
	}
	
	// Нет токенов - блокируем
	blockTime := now.Add(rl.config.BlockTime)
	rl.blocked["client"] = blockTime
	
	return false, 0, blockTime, rl.config.BlockTime
}

// checkSlidingWindow проверяет лимит по алгоритму sliding window
func (rl *RateLimiter) checkSlidingWindow(now time.Time) (bool, int, time.Time, time.Duration) {
	windowStart := now.Add(-rl.config.Window)
	
	// Очищаем старые запросы
	for clientID, info := range rl.requests {
		if info.FirstSeen.Before(windowStart) {
			delete(rl.requests, clientID)
		}
	}
	
	// Подсчитываем текущие запросы
	totalRequests := 0
	for _, info := range rl.requests {
		totalRequests += info.Count
	}
	
	if totalRequests < rl.config.Limit {
		// Добавляем новый запрос
		rl.requests["client"] = &RequestInfo{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}
		return true, rl.config.Limit - totalRequests - 1, now.Add(rl.config.Window), 0
	}
	
	// Лимит превышен
	blockTime := now.Add(rl.config.BlockTime)
	rl.blocked["client"] = blockTime
	
	return false, 0, blockTime, rl.config.BlockTime
}

// checkFixedWindow проверяет лимит по алгоритму fixed window
func (rl *RateLimiter) checkFixedWindow(now time.Time) (bool, int, time.Time, time.Duration) {
	windowStart := now.Truncate(rl.config.Window)
	
	// Очищаем старые окна
	for clientID, info := range rl.requests {
		if info.FirstSeen.Before(windowStart) {
			delete(rl.requests, clientID)
		}
	}
	
	// Подсчитываем запросы в текущем окне
	totalRequests := 0
	for _, info := range rl.requests {
		if info.FirstSeen.After(windowStart) {
			totalRequests += info.Count
		}
	}
	
	if totalRequests < rl.config.Limit {
		// Добавляем новый запрос
		rl.requests["client"] = &RequestInfo{
			Count:     1,
			FirstSeen: now,
			LastSeen:  now,
		}
		return true, rl.config.Limit - totalRequests - 1, windowStart.Add(rl.config.Window), 0
	}
	
	// Лимит превышен
	blockTime := now.Add(rl.config.BlockTime)
	rl.blocked["client"] = blockTime
	
	return false, 0, blockTime, rl.config.BlockTime
}

// GetStats возвращает статистику rate limiter'а
func (arl *APIRateLimiter) GetStats() map[string]interface{} {
	arl.mutex.RLock()
	defer arl.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"limiters": make(map[string]interface{}),
		"global_limit": arl.globalLimit,
	}
	
	for requestType, config := range arl.limiters {
		stats["limiters"].(map[string]interface{})[requestType] = map[string]interface{}{
			"type":        config.Type,
			"limit":       config.Limit,
			"window":      config.Window.String(),
			"burst":       config.Burst,
			"refill_rate": config.RefillRate,
			"block_time":  config.BlockTime.String(),
			"whitelist":   config.Whitelist,
			"blacklist":   config.Blacklist,
		}
	}
	
	return stats
}

// AddToWhitelist добавляет клиента в whitelist
func (arl *APIRateLimiter) AddToWhitelist(requestType, clientID string) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()
	
	if config, exists := arl.limiters[requestType]; exists {
		config.Whitelist = append(config.Whitelist, clientID)
	}
}

// AddToBlacklist добавляет клиента в blacklist
func (arl *APIRateLimiter) AddToBlacklist(requestType, clientID string) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()
	
	if config, exists := arl.limiters[requestType]; exists {
		config.Blacklist = append(config.Blacklist, clientID)
	}
}

// RemoveFromWhitelist удаляет клиента из whitelist
func (arl *APIRateLimiter) RemoveFromWhitelist(requestType, clientID string) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()
	
	if config, exists := arl.limiters[requestType]; exists {
		for i, id := range config.Whitelist {
			if id == clientID {
				config.Whitelist = append(config.Whitelist[:i], config.Whitelist[i+1:]...)
				break
			}
		}
	}
}

// RemoveFromBlacklist удаляет клиента из blacklist
func (arl *APIRateLimiter) RemoveFromBlacklist(requestType, clientID string) {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()
	
	if config, exists := arl.limiters[requestType]; exists {
		for i, id := range config.Blacklist {
			if id == clientID {
				config.Blacklist = append(config.Blacklist[:i], config.Blacklist[i+1:]...)
				break
			}
		}
	}
}

// Stop останавливает rate limiter
func (arl *APIRateLimiter) Stop() {
	arl.cancel()
	log.Println("API Rate Limiter stopped")
}
