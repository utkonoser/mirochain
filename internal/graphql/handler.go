package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
)

// GraphQLHandler обрабатывает GraphQL запросы
type GraphQLHandler struct {
	schema graphql.Schema
}

// NewGraphQLHandler создает новый GraphQL handler
func NewGraphQLHandler(schema graphql.Schema) *GraphQLHandler {
	return &GraphQLHandler{
		schema: schema,
	}
}

// GraphQLRequest представляет GraphQL запрос
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

// GraphQLResponse представляет GraphQL ответ
type GraphQLResponse struct {
	Data   interface{} `json:"data,omitempty"`
	Errors []error     `json:"errors,omitempty"`
}

// ServeHTTP обрабатывает HTTP запросы
func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем CORS заголовки
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Обрабатываем OPTIONS запросы
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Поддерживаем только GET и POST
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest

	// Парсим запрос в зависимости от метода
	if r.Method == "GET" {
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}
		req.Query = query

		// Парсим переменные из query string
		if variables := r.URL.Query().Get("variables"); variables != "" {
			if err := json.Unmarshal([]byte(variables), &req.Variables); err != nil {
				http.Error(w, "Invalid variables JSON", http.StatusBadRequest)
				return
			}
		}

		req.OperationName = r.URL.Query().Get("operationName")
	} else {
		// POST запрос
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	}

	// Выполняем GraphQL запрос
	result := h.executeQuery(req)

	// Отправляем ответ
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// executeQuery выполняет GraphQL запрос
func (h *GraphQLHandler) executeQuery(req GraphQLRequest) GraphQLResponse {
	// Выполняем запрос
	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
	})

	// Проверяем на ошибки
	if len(result.Errors) > 0 {
		errors := make([]error, len(result.Errors))
		for i, err := range result.Errors {
			errors[i] = err
		}
		return GraphQLResponse{
			Data:   result.Data,
			Errors: errors,
		}
	}

	return GraphQLResponse{
		Data: result.Data,
	}
}

// GraphiQLHandler предоставляет GraphiQL интерфейс
type GraphiQLHandler struct{}

// NewGraphiQLHandler создает новый GraphiQL handler
func NewGraphiQLHandler() *GraphiQLHandler {
	return &GraphiQLHandler{}
}

// ServeHTTP отдает GraphiQL интерфейс
func (h *GraphiQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>MiroChain GraphQL Playground</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
    <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
    <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
    <div id="root">
        <style>
            body {
                background-color: rgb(23, 42, 58);
                font-family: Open Sans, sans-serif;
                height: 90vh;
                margin: 0;
                width: 100%;
            }
            #root {
                height: 100vh;
                width: 100%;
            }
            .loading {
                color: white;
                display: flex;
                justify-content: center;
                align-items: center;
                height: 100vh;
                font-size: 14px;
                font-family: 'Open Sans', sans-serif;
            }
        </style>
        <div class="loading">Loading GraphQL Playground...</div>
    </div>
    <script>
        window.addEventListener('load', function (event) {
            GraphQLPlayground.init(document.getElementById('root'), {
                endpoint: window.location.origin + '/graphql',
                settings: {
                    'request.credentials': 'include',
                },
                tabs: [
                    {
                        endpoint: window.location.origin + '/graphql',
                        query: ` + "`" + `query GetBlockchain {
  blockchain {
    height
    difficulty
    totalSupply
    hashRate
    stats {
      totalBlocks
      totalTransactions
      totalAddresses
      averageBlockTime
      cacheHitRate
    }
  }
}

query GetBlock($height: Int!) {
  block(height: $height) {
    hash
    height
    timestamp
    previousHash
    merkleRoot
    nonce
    difficulty
    size
    gasUsed
    gasLimit
  }
}

query GetWallet($address: String!) {
  wallet(address: $address) {
    address
    balance
    nonce
    createdAt
  }
}

query GetContract($address: String!) {
  contract(address: $address) {
    address
    code
    owner
    balance
    createdAt
    updatedAt
    storage {
      address
      values {
        key
        value
      }
    }
  }
}

query GetPeers {
  peers {
    id
    address
    port
    lastSeen
    latency
  }
}

query GetNetworkStats {
  networkStats {
    totalPeers
    connectedPeers
    averageLatency
    bandwidth
  }
}

query GetConsensus {
  consensus {
    algorithm
    validators
    currentRound
    nextValidator
  }
}

mutation SendTransaction($input: SendTransactionInput!) {
  sendTransaction(input: $input) {
    success
    transaction {
      hash
      from
      to
      amount
      fee
      timestamp
      status
    }
    error
  }
}

mutation DeployContract($input: DeployContractInput!) {
  deployContract(input: $input) {
    success
    contract {
      address
      code
      owner
      balance
      createdAt
    }
    gasUsed
    error
  }
}` + "`" + `,
                        variables: {
                            "height": 0
                        }
                    }
                ]
            })
        })
    </script>
</body>
</html>`
	
	fmt.Fprint(w, html)
}

// IntrospectionHandler предоставляет GraphQL introspection
type IntrospectionHandler struct {
	schema graphql.Schema
}

// NewIntrospectionHandler создает новый introspection handler
func NewIntrospectionHandler(schema graphql.Schema) *IntrospectionHandler {
	return &IntrospectionHandler{
		schema: schema,
	}
}

// ServeHTTP отдает GraphQL introspection
func (h *IntrospectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Выполняем introspection запрос
	result := graphql.Do(graphql.Params{
		Schema: h.schema,
		RequestString: introspectionQuery,
	})
	
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode introspection", http.StatusInternalServerError)
		return
	}
}

// Introspection query для получения схемы
const introspectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type { ...TypeRef }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`
