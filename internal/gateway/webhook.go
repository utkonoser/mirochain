package gateway

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WebhookManager управляет webhooks
type WebhookManager struct {
	webhooks map[string]*Webhook
	mu       sync.RWMutex
}

// Webhook представляет webhook
type Webhook struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Events      []string          `json:"events"`
	Secret      string            `json:"secret,omitempty"`
	Headers     map[string]string `json:"headers"`
	Active      bool              `json:"active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastTrigger time.Time         `json:"last_trigger,omitempty"`
	RetryCount  int               `json:"retry_count"`
}

// WebhookRegistrationRequest представляет запрос на регистрацию webhook
type WebhookRegistrationRequest struct {
	URL     string            `json:"url"`
	Events  []string          `json:"events"`
	Secret  string            `json:"secret,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// WebhookTestRequest представляет запрос на тестирование webhook
type WebhookTestRequest struct {
	WebhookID string `json:"webhook_id"`
	Event     string `json:"event"`
	Data      interface{} `json:"data,omitempty"`
}

// WebhookTestResult представляет результат тестирования webhook
type WebhookTestResult struct {
	Success   bool          `json:"success"`
	Status    int           `json:"status"`
	Response  string        `json:"response"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// WebhookEvent представляет событие webhook
type WebhookEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
}

// NewWebhookManager создает новый WebhookManager
func NewWebhookManager() *WebhookManager {
	return &WebhookManager{
		webhooks: make(map[string]*Webhook),
	}
}

// RegisterWebhook регистрирует новый webhook
func (wm *WebhookManager) RegisterWebhook(req WebhookRegistrationRequest) (*Webhook, error) {
	// Валидация URL
	if req.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	
	// Валидация событий
	if len(req.Events) == 0 {
		return nil, fmt.Errorf("at least one event is required")
	}
	
	// Создаем webhook
	webhook := &Webhook{
		ID:        generateWebhookID(),
		URL:       req.URL,
		Events:    req.Events,
		Secret:    req.Secret,
		Headers:   req.Headers,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Сохраняем webhook
	wm.mu.Lock()
	wm.webhooks[webhook.ID] = webhook
	wm.mu.Unlock()
	
	return webhook, nil
}

// ProcessWebhook обрабатывает webhook запрос
func (wm *WebhookManager) ProcessWebhook(w http.ResponseWriter, r *http.Request, webhookID string) {
	wm.mu.RLock()
	webhook, exists := wm.webhooks[webhookID]
	wm.mu.RUnlock()
	
	if !exists {
		http.Error(w, "Webhook not found", http.StatusNotFound)
		return
	}
	
	if !webhook.Active {
		http.Error(w, "Webhook is inactive", http.StatusForbidden)
		return
	}
	
	// Обрабатываем webhook в зависимости от метода
	switch r.Method {
	case "GET":
		wm.handleWebhookGet(w, r, webhook)
	case "POST":
		wm.handleWebhookPost(w, r, webhook)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWebhookGet обрабатывает GET запросы к webhook
func (wm *WebhookManager) handleWebhookGet(w http.ResponseWriter, r *http.Request, webhook *Webhook) {
	response := map[string]interface{}{
		"webhook_id": webhook.ID,
		"status": "active",
		"events": webhook.Events,
		"created_at": webhook.CreatedAt,
		"last_trigger": webhook.LastTrigger,
		"retry_count": webhook.RetryCount,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleWebhookPost обрабатывает POST запросы к webhook
func (wm *WebhookManager) handleWebhookPost(w http.ResponseWriter, r *http.Request, webhook *Webhook) {
	// Читаем данные события
	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Проверяем, поддерживается ли событие
	if !wm.isEventSupported(webhook, event.Type) {
		http.Error(w, "Event type not supported", http.StatusBadRequest)
		return
	}
	
	// Обрабатываем событие
	wm.processEvent(webhook, event)
	
	// Отправляем подтверждение
	response := map[string]interface{}{
		"success": true,
		"webhook_id": webhook.ID,
		"event_id": event.ID,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TestWebhook тестирует webhook
func (wm *WebhookManager) TestWebhook(req WebhookTestRequest) (*WebhookTestResult, error) {
	wm.mu.RLock()
	webhook, exists := wm.webhooks[req.WebhookID]
	wm.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("webhook not found")
	}
	
	// Создаем тестовое событие
	event := WebhookEvent{
		ID:        generateEventID(),
		Type:      req.Event,
		Data:      req.Data,
		Timestamp: time.Now(),
		Source:    "test",
	}
	
	// Отправляем webhook
	start := time.Now()
	success, status, response, err := wm.sendWebhook(webhook, event)
	duration := time.Since(start)
	
	result := &WebhookTestResult{
		Success:  success,
		Status:   status,
		Response: response,
		Duration: duration,
	}
	
	if err != nil {
		result.Error = err.Error()
	}
	
	return result, nil
}

// processEvent обрабатывает событие
func (wm *WebhookManager) processEvent(webhook *Webhook, event WebhookEvent) {
	// Отправляем webhook асинхронно
	go func() {
		success, status, _, err := wm.sendWebhook(webhook, event)
		
		// Обновляем статистику webhook
		wm.mu.Lock()
		webhook.LastTrigger = time.Now()
		if !success {
			webhook.RetryCount++
		}
		wm.mu.Unlock()
		
		// Логируем результат
		fmt.Printf("Webhook %s: success=%v, status=%d, error=%v\n", 
			webhook.ID, success, status, err)
	}()
}

// sendWebhook отправляет webhook
func (wm *WebhookManager) sendWebhook(webhook *Webhook, event WebhookEvent) (bool, int, string, error) {
	// Подготавливаем данные
	eventData, err := json.Marshal(event)
	if err != nil {
		return false, 0, "", err
	}
	
	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(eventData))
	if err != nil {
		return false, 0, "", err
	}
	
	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MiroChain-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", event.Type)
	req.Header.Set("X-Webhook-ID", webhook.ID)
	
	// Добавляем пользовательские заголовки
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}
	
	// Добавляем подпись если есть секрет
	if webhook.Secret != "" {
		signature := wm.generateSignature(eventData, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}
	
	// Отправляем запрос
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, "", err
	}
	defer resp.Body.Close()
	
	// Читаем ответ
	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)
	
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	return success, resp.StatusCode, responseBody.String(), nil
}

// isEventSupported проверяет, поддерживается ли событие
func (wm *WebhookManager) isEventSupported(webhook *Webhook, eventType string) bool {
	for _, supportedEvent := range webhook.Events {
		if supportedEvent == eventType || supportedEvent == "*" {
			return true
		}
	}
	return false
}

// generateSignature генерирует подпись для webhook
func (wm *WebhookManager) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// generateWebhookID генерирует уникальный ID для webhook
func generateWebhookID() string {
	return fmt.Sprintf("wh_%d", time.Now().UnixNano())
}

// generateEventID генерирует уникальный ID для события
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// TriggerEvent триггерит событие для всех подписанных webhooks
func (wm *WebhookManager) TriggerEvent(eventType string, data interface{}) {
	event := WebhookEvent{
		ID:        generateEventID(),
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
		Source:    "mirochain",
	}
	
	wm.mu.RLock()
	for _, webhook := range wm.webhooks {
		if webhook.Active && wm.isEventSupported(webhook, eventType) {
			go wm.processEvent(webhook, event)
		}
	}
	wm.mu.RUnlock()
}

// GetWebhook возвращает webhook по ID
func (wm *WebhookManager) GetWebhook(id string) (*Webhook, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	webhook, exists := wm.webhooks[id]
	return webhook, exists
}

// ListWebhooks возвращает список всех webhooks
func (wm *WebhookManager) ListWebhooks() []*Webhook {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	webhooks := make([]*Webhook, 0, len(wm.webhooks))
	for _, webhook := range wm.webhooks {
		webhooks = append(webhooks, webhook)
	}
	
	return webhooks
}

// DeleteWebhook удаляет webhook
func (wm *WebhookManager) DeleteWebhook(id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	if _, exists := wm.webhooks[id]; !exists {
		return fmt.Errorf("webhook not found")
	}
	
	delete(wm.webhooks, id)
	return nil
}

// UpdateWebhook обновляет webhook
func (wm *WebhookManager) UpdateWebhook(id string, updates map[string]interface{}) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	webhook, exists := wm.webhooks[id]
	if !exists {
		return fmt.Errorf("webhook not found")
	}
	
	// Обновляем поля
	if url, ok := updates["url"].(string); ok {
		webhook.URL = url
	}
	if events, ok := updates["events"].([]string); ok {
		webhook.Events = events
	}
	if secret, ok := updates["secret"].(string); ok {
		webhook.Secret = secret
	}
	if headers, ok := updates["headers"].(map[string]string); ok {
		webhook.Headers = headers
	}
	if active, ok := updates["active"].(bool); ok {
		webhook.Active = active
	}
	
	webhook.UpdatedAt = time.Now()
	
	return nil
}
