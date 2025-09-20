//go:build graphql_demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== MiroChain GraphQL API Demo ===")
	
	// Ждем запуска узла
	fmt.Println("Waiting for node to start...")
	time.Sleep(3 * time.Second)
	
	// Тестируем различные GraphQL запросы
	testQueries()
	testMutations()
	testIntrospection()
	
	fmt.Println("\n=== GraphQL Demo Complete ===")
}

func testQueries() {
	fmt.Println("\n1. Testing GraphQL Queries...")
	
	// Тест 1: Получение информации о блокчейне
	fmt.Println("\n1.1. Getting blockchain information...")
	query := `
		query GetBlockchain {
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
	`
	executeGraphQLQuery(query, nil)
	
	// Тест 2: Получение блока
	fmt.Println("\n1.2. Getting block information...")
	query = `
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
	`
	variables := map[string]interface{}{
		"height": 0,
	}
	executeGraphQLQuery(query, variables)
	
	// Тест 3: Получение информации о сети
	fmt.Println("\n1.3. Getting network information...")
	query = `
		query GetNetworkInfo {
			peers {
				id
				address
				port
				lastSeen
				latency
			}
			networkStats {
				totalPeers
				connectedPeers
				averageLatency
				bandwidth
			}
			consensus {
				algorithm
				validators
				currentRound
				nextValidator
			}
		}
	`
	executeGraphQLQuery(query, nil)
}

func testMutations() {
	fmt.Println("\n2. Testing GraphQL Mutations...")
	
	// Тест 1: Отправка транзакции
	fmt.Println("\n2.1. Sending a transaction...")
	query := `
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
	`
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"to":     "recipient_address",
			"amount": "100",
			"fee":    "1",
			"data":   "Hello MiroChain!",
		},
	}
	executeGraphQLQuery(query, variables)
	
	// Тест 2: Развертывание контракта
	fmt.Println("\n2.2. Deploying a smart contract...")
	query = `
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
		}
	`
	variables = map[string]interface{}{
		"input": map[string]interface{}{
			"code":            "PUSH 42\nSTORE counter\nRETURN",
			"owner":           "alice",
			"initialBalance":  "0",
		},
	}
	executeGraphQLQuery(query, variables)
}

func testIntrospection() {
	fmt.Println("\n3. Testing GraphQL Introspection...")
	
	// Тест 1: Получение схемы
	fmt.Println("\n3.1. Getting GraphQL schema...")
	query := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					name
					kind
					description
				}
			}
		}
	`
	executeGraphQLQuery(query, nil)
	
	// Тест 2: Получение типов запросов
	fmt.Println("\n3.2. Getting query types...")
	query = `
		query GetQueryTypes {
			__schema {
				queryType {
					fields {
						name
						description
						type {
							name
							kind
						}
					}
				}
			}
		}
	`
	executeGraphQLQuery(query, nil)
}

func executeGraphQLQuery(query string, variables map[string]interface{}) {
	// Подготавливаем запрос
	request := map[string]interface{}{
		"query": query,
	}
	
	if variables != nil {
		request["variables"] = variables
	}
	
	// Кодируем в JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error encoding request: %v", err)
		return
	}
	
	// Отправляем запрос
	resp, err := http.Post("http://localhost:12901/graphql", 
		"application/json", 
		bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()
	
	// Декодируем ответ
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}
	
	// Выводим результат
	if errors, ok := result["errors"].([]interface{}); ok && len(errors) > 0 {
		fmt.Printf("GraphQL Errors: %v\n", errors)
	} else {
		fmt.Printf("GraphQL Response: %v\n", result["data"])
	}
}

func testAPIEndpoints() {
	fmt.Println("\n4. Testing API Gateway Endpoints...")
	
	// Тест 1: Health check
	fmt.Println("\n4.1. Testing health check...")
	testEndpoint("GET", "http://localhost:12901/health", nil)
	
	// Тест 2: API документация
	fmt.Println("\n4.2. Testing API documentation...")
	testEndpoint("GET", "http://localhost:12901/docs", nil)
	
	// Тест 3: Swagger спецификация
	fmt.Println("\n4.3. Testing Swagger specification...")
	testEndpoint("GET", "http://localhost:12901/swagger.json", nil)
	
	// Тест 4: Версионированные API
	fmt.Println("\n4.4. Testing versioned API...")
	testEndpoint("GET", "http://localhost:12901/api/v1/blockchain", nil)
	testEndpoint("GET", "http://localhost:12901/api/v2/blockchain", nil)
	testEndpoint("GET", "http://localhost:12901/api/latest/blockchain", nil)
}

func testWebhooks() {
	fmt.Println("\n5. Testing Webhook System...")
	
	// Тест 1: Регистрация webhook
	fmt.Println("\n5.1. Registering a webhook...")
	webhookData := map[string]interface{}{
		"url": "http://localhost:8080/webhook/test",
		"events": []string{"new_block", "new_transaction"},
		"secret": "test_secret",
		"headers": map[string]string{
			"Authorization": "Bearer test_token",
		},
	}
	testEndpoint("POST", "http://localhost:12901/webhooks/register", webhookData)
	
	// Тест 2: Тестирование webhook
	fmt.Println("\n5.2. Testing webhook...")
	testData := map[string]interface{}{
		"webhook_id": "wh_test_123",
		"event": "new_block",
		"data": map[string]interface{}{
			"block_height": 1,
			"block_hash": "test_hash",
		},
	}
	testEndpoint("POST", "http://localhost:12901/webhooks/test", testData)
}

func testEndpoint(method, url string, data interface{}) {
	var resp *http.Response
	var err error
	
	if data != nil {
		jsonData, _ := json.Marshal(data)
		req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
	} else {
		resp, err = http.Get(url)
	}
	
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	fmt.Printf("%s %s: %d\n", method, url, resp.StatusCode)
	if len(result) > 0 {
		fmt.Printf("Response: %v\n", result)
	}
}
