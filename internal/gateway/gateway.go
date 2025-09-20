package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mirochain/internal/graphql"
)

// APIGateway предоставляет единую точку входа для всех API
type APIGateway struct {
	graphqlHandler *graphql.GraphQLHandler
	graphiqlHandler *graphql.GraphiQLHandler
	introspectionHandler *graphql.IntrospectionHandler
	webhookManager *WebhookManager
	versionManager *VersionManager
}

// NewAPIGateway создает новый API Gateway
func NewAPIGateway(
	graphqlHandler *graphql.GraphQLHandler,
	graphiqlHandler *graphql.GraphiQLHandler,
	introspectionHandler *graphql.IntrospectionHandler,
	webhookManager *WebhookManager,
	versionManager *VersionManager,
) *APIGateway {
	return &APIGateway{
		graphqlHandler: graphqlHandler,
		graphiqlHandler: graphiqlHandler,
		introspectionHandler: introspectionHandler,
		webhookManager: webhookManager,
		versionManager: versionManager,
	}
}

// RegisterRoutes регистрирует все маршруты API Gateway
func (g *APIGateway) RegisterRoutes(mux *http.ServeMux) {
	// GraphQL endpoints
	mux.HandleFunc("/graphql", g.handleGraphQL)
	mux.HandleFunc("/graphiql", g.handleGraphiQL)
	mux.HandleFunc("/introspection", g.handleIntrospection)
	
	// Versioned API endpoints
	mux.HandleFunc("/api/v1/", g.handleVersionedAPI)
	mux.HandleFunc("/api/v2/", g.handleVersionedAPI)
	mux.HandleFunc("/api/latest/", g.handleVersionedAPI)
	
	// Webhook endpoints
	mux.HandleFunc("/webhooks/", g.handleWebhooks)
	mux.HandleFunc("/webhooks/register", g.handleWebhookRegistration)
	mux.HandleFunc("/webhooks/test", g.handleWebhookTest)
	
	// Health check
	mux.HandleFunc("/health", g.handleHealth)
	
	// API documentation
	mux.HandleFunc("/docs", g.handleDocs)
	mux.HandleFunc("/swagger.json", g.handleSwagger)
}

// handleGraphQL обрабатывает GraphQL запросы
func (g *APIGateway) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	// Добавляем версионирование в заголовки
	w.Header().Set("X-API-Version", "v1")
	w.Header().Set("X-GraphQL-Version", "1.0")
	
	g.graphqlHandler.ServeHTTP(w, r)
}

// handleGraphiQL обрабатывает GraphiQL интерфейс
func (g *APIGateway) handleGraphiQL(w http.ResponseWriter, r *http.Request) {
	g.graphiqlHandler.ServeHTTP(w, r)
}

// handleIntrospection обрабатывает GraphQL introspection
func (g *APIGateway) handleIntrospection(w http.ResponseWriter, r *http.Request) {
	g.introspectionHandler.ServeHTTP(w, r)
}

// handleVersionedAPI обрабатывает версионированные API запросы
func (g *APIGateway) handleVersionedAPI(w http.ResponseWriter, r *http.Request) {
	// Извлекаем версию из URL
	version := g.extractVersionFromPath(r.URL.Path)
	
	// Устанавливаем заголовки версии
	w.Header().Set("X-API-Version", version)
	w.Header().Set("X-API-Deprecated", strconv.FormatBool(g.versionManager.IsDeprecated(version)))
	
	// Перенаправляем на соответствующую версию API
	g.routeToVersion(w, r, version)
}

// handleWebhooks обрабатывает webhook запросы
func (g *APIGateway) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	webhookID := strings.TrimPrefix(r.URL.Path, "/webhooks/")
	if webhookID == "" {
		http.Error(w, "Webhook ID required", http.StatusBadRequest)
		return
	}
	
	// Обрабатываем webhook
	g.webhookManager.ProcessWebhook(w, r, webhookID)
}

// handleWebhookRegistration обрабатывает регистрацию webhooks
func (g *APIGateway) handleWebhookRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req WebhookRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	webhook, err := g.webhookManager.RegisterWebhook(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(webhook)
}

// handleWebhookTest обрабатывает тестирование webhooks
func (g *APIGateway) handleWebhookTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req WebhookTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	result, err := g.webhookManager.TestWebhook(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleHealth обрабатывает health check
func (g *APIGateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version": "1.0.0",
		"services": map[string]string{
			"graphql": "running",
			"webhooks": "running",
			"versioning": "running",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleDocs обрабатывает документацию API
func (g *APIGateway) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>MiroChain API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { font-weight: bold; color: #0066cc; }
        .version { background: #e6f3ff; padding: 5px; border-radius: 3px; display: inline-block; margin: 5px; }
    </style>
</head>
<body>
    <h1>MiroChain API Documentation</h1>
    
    <h2>GraphQL API</h2>
    <div class="endpoint">
        <span class="method">POST</span> /graphql - GraphQL endpoint
        <div class="version">v1</div>
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /graphiql - GraphiQL playground
        <div class="version">v1</div>
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /introspection - GraphQL schema introspection
        <div class="version">v1</div>
    </div>
    
    <h2>REST API</h2>
    <div class="endpoint">
        <span class="method">GET</span> /api/v1/blockchain - Blockchain information
        <div class="version">v1</div>
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /api/v2/blockchain - Enhanced blockchain information
        <div class="version">v2</div>
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /api/latest/blockchain - Latest blockchain information
        <div class="version">latest</div>
    </div>
    
    <h2>Webhooks</h2>
    <div class="endpoint">
        <span class="method">POST</span> /webhooks/register - Register webhook
        <div class="version">v1</div>
    </div>
    <div class="endpoint">
        <span class="method">POST</span> /webhooks/test - Test webhook
        <div class="version">v1</div>
    </div>
    
    <h2>Health & Documentation</h2>
    <div class="endpoint">
        <span class="method">GET</span> /health - Health check
        <div class="version">v1</div>
    </div>
    <div class="endpoint">
        <span class="method">GET</span> /swagger.json - OpenAPI specification
        <div class="version">v1</div>
    </div>
    
    <p><a href="/graphiql">Try GraphQL Playground</a></p>
    <p><a href="/swagger.json">View OpenAPI Spec</a></p>
</body>
</html>`
	
	fmt.Fprint(w, html)
}

// handleSwagger обрабатывает OpenAPI спецификацию
func (g *APIGateway) handleSwagger(w http.ResponseWriter, r *http.Request) {
	swagger := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title": "MiroChain API",
			"version": "1.0.0",
			"description": "Comprehensive blockchain API with GraphQL and REST endpoints",
		},
		"servers": []map[string]interface{}{
			{"url": "http://localhost:8080", "description": "Development server"},
		},
		"paths": map[string]interface{}{
			"/graphql": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "GraphQL endpoint",
					"description": "Execute GraphQL queries and mutations",
					"requestBody": map[string]interface{}{
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"query": map[string]interface{}{"type": "string"},
										"variables": map[string]interface{}{"type": "object"},
										"operationName": map[string]interface{}{"type": "string"},
									},
								},
							},
						},
					},
				},
			},
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health check",
					"description": "Check API health status",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "API is healthy",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]interface{}{"type": "string"},
											"timestamp": map[string]interface{}{"type": "string"},
											"version": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swagger)
}

// extractVersionFromPath извлекает версию из пути
func (g *APIGateway) extractVersionFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		return parts[2] // /api/v1/... -> v1
	}
	return "v1"
}

// routeToVersion перенаправляет на соответствующую версию API
func (g *APIGateway) routeToVersion(w http.ResponseWriter, r *http.Request, version string) {
	// Удаляем префикс версии из пути
	path := strings.TrimPrefix(r.URL.Path, "/api/"+version)
	
	// Перенаправляем на соответствующую версию
	switch version {
	case "v1":
		g.handleV1API(w, r, path)
	case "v2":
		g.handleV2API(w, r, path)
	case "latest":
		g.handleLatestAPI(w, r, path)
	default:
		http.Error(w, "Unsupported API version", http.StatusBadRequest)
	}
}

// handleV1API обрабатывает API v1
func (g *APIGateway) handleV1API(w http.ResponseWriter, r *http.Request, path string) {
	// Простая реализация v1 API
	response := map[string]interface{}{
		"version": "v1",
		"path": path,
		"message": "MiroChain API v1",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleV2API обрабатывает API v2
func (g *APIGateway) handleV2API(w http.ResponseWriter, r *http.Request, path string) {
	// Расширенная реализация v2 API
	response := map[string]interface{}{
		"version": "v2",
		"path": path,
		"message": "MiroChain API v2 - Enhanced features",
		"timestamp": time.Now().Format(time.RFC3339),
		"features": []string{
			"Enhanced blockchain queries",
			"Advanced smart contract support",
			"Improved error handling",
			"Better performance",
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleLatestAPI обрабатывает latest API
func (g *APIGateway) handleLatestAPI(w http.ResponseWriter, r *http.Request, path string) {
	// Перенаправляем на v2 как latest
	g.handleV2API(w, r, path)
}
