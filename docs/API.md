# MiroChain API Documentation

## Overview

MiroChain предоставляет REST API для взаимодействия с блокчейн сетью. API позволяет получать информацию о блокчейне, управлять кошельками, создавать транзакции и мониторить состояние сети.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

В текущей версии API не требует аутентификации. В будущих версиях будет добавлена система токенов.

## Endpoints

### Blockchain Information

#### GET /blockchain/status
Получить общую информацию о блокчейне.

**Response:**
```json
{
  "height": 10,
  "difficulty": 4,
  "total_transactions": 25,
  "last_block_hash": "abc123...",
  "genesis_hash": "def456...",
  "is_valid": true
}
```

#### GET /blockchain/blocks
Получить список блоков.

**Query Parameters:**
- `limit` (optional): Количество блоков (по умолчанию: 10)
- `offset` (optional): Смещение (по умолчанию: 0)

**Response:**
```json
{
  "blocks": [
    {
      "index": 1,
      "timestamp": 1640995200,
      "hash": "abc123...",
      "previous_hash": "def456...",
      "merkle_root": "ghi789...",
      "nonce": 12345,
      "difficulty": 4,
      "transaction_count": 2
    }
  ],
  "total": 10,
  "limit": 10,
  "offset": 0
}
```

#### GET /blockchain/blocks/{hash}
Получить информацию о конкретном блоке.

**Response:**
```json
{
  "index": 1,
  "timestamp": 1640995200,
  "hash": "abc123...",
  "previous_hash": "def456...",
  "merkle_root": "ghi789...",
  "nonce": 12345,
  "difficulty": 4,
  "transactions": [
    {
      "id": "tx123...",
      "inputs": [...],
      "outputs": [...],
      "fee": 10
    }
  ]
}
```

### Wallet Management

#### GET /wallets
Получить список всех кошельков.

**Response:**
```json
{
  "wallets": [
    {
      "address": "1A2B3C4D5E6F...",
      "public_key": "04...",
      "balance": 1000000
    }
  ],
  "total": 1
}
```

#### POST /wallets
Создать новый кошелек.

**Response:**
```json
{
  "address": "1A2B3C4D5E6F...",
  "public_key": "04...",
  "private_key": "1234567890abcdef...",
  "message": "Wallet created successfully"
}
```

#### GET /wallets/{address}
Получить информацию о конкретном кошельке.

**Response:**
```json
{
  "address": "1A2B3C4D5E6F...",
  "public_key": "04...",
  "balance": 1000000,
  "utxo_count": 5
}
```

#### GET /wallets/{address}/balance
Получить баланс кошелька.

**Response:**
```json
{
  "address": "1A2B3C4D5E6F...",
  "balance": 1000000,
  "confirmed_balance": 950000,
  "unconfirmed_balance": 50000
}
```

#### GET /wallets/{address}/utxos
Получить список UTXO для кошелька.

**Response:**
```json
{
  "address": "1A2B3C4D5E6F...",
  "utxos": [
    {
      "transaction_id": "tx123...",
      "output_index": 0,
      "value": 100000,
      "address": "1A2B3C4D5E6F...",
      "public_key": "04..."
    }
  ],
  "total": 1
}
```

### Transactions

#### POST /transactions
Создать новую транзакцию.

**Request Body:**
```json
{
  "from": "1A2B3C4D5E6F...",
  "to": "2B3C4D5E6F7G...",
  "amount": 100000,
  "private_key": "1234567890abcdef..."
}
```

**Response:**
```json
{
  "transaction_id": "tx123...",
  "status": "pending",
  "message": "Transaction created successfully"
}
```

#### GET /transactions/{id}
Получить информацию о транзакции.

**Response:**
```json
{
  "id": "tx123...",
  "status": "confirmed",
  "block_height": 5,
  "inputs": [...],
  "outputs": [...],
  "fee": 10,
  "timestamp": 1640995200
}
```

### Mining

#### GET /mining/status
Получить статус майнинга.

**Response:**
```json
{
  "is_mining": true,
  "active_miners": 2,
  "total_blocks": 15,
  "hash_rate": 1250000.5,
  "difficulty": 4,
  "mempool_size": 5
}
```

#### POST /mining/start
Запустить майнинг.

**Response:**
```json
{
  "status": "started",
  "message": "Mining started successfully"
}
```

#### POST /mining/stop
Остановить майнинг.

**Response:**
```json
{
  "status": "stopped",
  "message": "Mining stopped successfully"
}
```

### Network

#### GET /network/peers
Получить список подключенных пиров.

**Response:**
```json
{
  "peers": [
    {
      "id": "peer123...",
      "address": "127.0.0.1:8081",
      "status": "connected",
      "last_seen": 1640995200
    }
  ],
  "total": 1
}
```

#### GET /network/stats
Получить статистику сети.

**Response:**
```json
{
  "total_peers": 5,
  "connected_peers": 3,
  "messages_sent": 150,
  "messages_received": 142,
  "uptime": 3600
}
```

## Error Responses

Все ошибки возвращаются в следующем формате:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

### Common Error Codes

- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error
- `WALLET_NOT_FOUND` - Кошелек не найден
- `INSUFFICIENT_FUNDS` - Недостаточно средств
- `INVALID_TRANSACTION` - Неверная транзакция
- `BLOCK_NOT_FOUND` - Блок не найден

## Rate Limiting

В текущей версии rate limiting не реализован. В будущих версиях будет добавлен лимит 100 запросов в минуту на IP.

## WebSocket Support

### Connection
```
ws://localhost:8080/ws
```

### Events

#### New Block
```json
{
  "type": "new_block",
  "data": {
    "height": 11,
    "hash": "abc123...",
    "timestamp": 1640995200
  }
}
```

#### New Transaction
```json
{
  "type": "new_transaction",
  "data": {
    "id": "tx123...",
    "from": "1A2B3C4D5E6F...",
    "to": "2B3C4D5E6F7G...",
    "amount": 100000
  }
}
```

## Examples

### Создание кошелька и транзакции

```bash
# Создать кошелек
curl -X POST http://localhost:8080/api/v1/wallets

# Получить баланс
curl http://localhost:8080/api/v1/wallets/1A2B3C4D5E6F.../balance

# Создать транзакцию
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "from": "1A2B3C4D5E6F...",
    "to": "2B3C4D5E6F7G...",
    "amount": 100000,
    "private_key": "1234567890abcdef..."
  }'
```

### Мониторинг блокчейна

```bash
# Получить статус
curl http://localhost:8080/api/v1/blockchain/status

# Получить последние блоки
curl http://localhost:8080/api/v1/blockchain/blocks?limit=5

# Получить информацию о блоке
curl http://localhost:8080/api/v1/blockchain/blocks/abc123...
```

## SDK Examples

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type Wallet struct {
    Address    string `json:"address"`
    PublicKey  string `json:"public_key"`
    PrivateKey string `json:"private_key"`
}

func createWallet() (*Wallet, error) {
    resp, err := http.Post("http://localhost:8080/api/v1/wallets", "application/json", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var wallet Wallet
    err = json.NewDecoder(resp.Body).Decode(&wallet)
    return &wallet, err
}

func main() {
    wallet, err := createWallet()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created wallet: %s\n", wallet.Address)
}
```

### JavaScript
```javascript
class MiroChainAPI {
    constructor(baseURL = 'http://localhost:8080/api/v1') {
        this.baseURL = baseURL;
    }

    async createWallet() {
        const response = await fetch(`${this.baseURL}/wallets`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        });
        return await response.json();
    }

    async getBalance(address) {
        const response = await fetch(`${this.baseURL}/wallets/${address}/balance`);
        return await response.json();
    }

    async createTransaction(from, to, amount, privateKey) {
        const response = await fetch(`${this.baseURL}/transactions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ from, to, amount, private_key: privateKey })
        });
        return await response.json();
    }
}

// Usage
const api = new MiroChainAPI();
const wallet = await api.createWallet();
console.log('Created wallet:', wallet.address);
```

## Changelog

### v1.0.0
- Initial API release
- Basic blockchain operations
- Wallet management
- Transaction creation
- Mining control
- P2P network monitoring
