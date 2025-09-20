# MiroChain - Блокчейн сеть на Go

## Описание проекта
MiroChain - это полнофункциональная блокчейн сеть на языке Go с поддержкой P2P сети, майнинга, смарт-контрактов, токенов, NFT, sidechains и state channels.

## Архитектура

### Основные компоненты
- **Core** - основные структуры данных и логика блокчейна
- **Crypto** - криптографические функции и алгоритмы
- **Network** - P2P сеть и коммуникация
- **Mining** - алгоритм майнинга и консенсуса
- **Wallet** - система кошельков
- **API** - REST API и CLI интерфейс
- **VM** - виртуальная машина для смарт-контрактов
- **Tokens** - система токенов ERC-20
- **NFT** - система NFT ERC-721
- **Sidechains** - боковые цепи
- **StateChannels** - каналы состояния

## План реализации

Все основные этапы (1-13) уже реализованы! См. раздел "Что уже реализовано" ниже.

## Структура проекта

```
mirochain/
├── cmd/                    # Основные исполняемые файлы
│   ├── node/              # Узел блокчейна
│   └── wallet/            # CLI кошелек
├── internal/              # Внутренние пакеты
│   ├── blockchain/        # Основная логика блокчейна
│   ├── crypto/           # Криптографические функции
│   ├── network/          # P2P сеть
│   ├── mining/           # Майнинг
│   ├── wallet/           # Система кошельков
│   ├── api/              # API сервер
│   ├── vm/               # Виртуальная машина
│   ├── tokens/           # Система токенов
│   ├── nft/              # Система NFT
│   ├── sidechain/        # Sidechains
│   └── statechannel/     # State Channels
├── pkg/                   # Публичные пакеты
├── tests/                 # Тесты
├── examples/              # Примеры использования
├── configs/              # Конфигурационные файлы
├── docs/                 # Документация
├── go.mod
├── go.sum
└── README.md
```

## Технические требования

- Go 1.25+
- Git
- Make (опционально)

## 🔐 Переключение алгоритмов

MiroChain поддерживает гибкое переключение между различными криптографическими алгоритмами:

### Классические алгоритмы
- **ECDSA** - Elliptic Curve Digital Signature Algorithm
- **Ed25519** - Edwards Curve Digital Signature Algorithm
- **RSA** - Rivest-Shamir-Adleman
- **Schnorr** - Schnorr Digital Signature Scheme

### Квантово-устойчивые алгоритмы
- **SPHINCS+** - Stateless hash-based signatures
- **Dilithium** - Lattice-based signatures
- **Falcon** - Lattice-based signatures
- **XMSS** - Stateful hash-based signatures
- **LMS** - Leighton-Micali signatures

### Переключение алгоритмов

```bash
# Запуск с классическими алгоритмами
go run cmd/node/main.go -algorithms=classic

# Запуск с квантово-устойчивыми алгоритмами
go run cmd/node/main.go -algorithms=quantum

# Запуск со смешанными алгоритмами
go run cmd/node/main.go -algorithms=mixed

# Конфигурация через JSON
go run cmd/node/main.go -config=algorithm_config.json
```

## Быстрый старт

### Локальная разработка

```bash
# Клонируем репозиторий
git clone https://github.com/utkonoser/mirochain.git
cd mirochain

# Устанавливаем зависимости
go mod download

# Запускаем узел
go run cmd/node/main.go -port=8080 -mining=false

# В другом терминале запускаем майнер
go run cmd/node/main.go -port=9080 -mining=true
```

### Docker

```bash
# Сборка и запуск с Docker Compose
docker-compose up -d

# Или сборка и запуск отдельного контейнера
docker build -t mirochain .
docker run -p 8080:8080 mirochain
```

### Kubernetes

```bash
# Применение манифестов
kubectl apply -f k8s/

# Проверка статуса
kubectl get pods -n mirochain
```

### Makefile

```bash
# Сборка
make build

# Запуск в режиме разработки
make dev

# Запуск с майнингом
make dev-mining

# Запуск нескольких узлов
make dev-multi

# Docker
make docker-build
make docker-compose-up

# Kubernetes
make k8s-apply
make k8s-status

# Тестирование
make test
make test-coverage

# Демонстрации
make demo
make demo-contracts
make demo-tokens
```

### Запуск узла

```bash
# Запуск узла с майнингом
go run cmd/node/main.go -port=8080 -mining=true

# Запуск узла без майнинга
go run cmd/node/main.go -port=8080 -mining=false

# Запуск с подключением к другим узлам
go run cmd/node/main.go -peers="127.0.0.1:8081,127.0.0.1:8082"
```

### Использование CLI

```bash
# Создание кошелька
go run cmd/wallet/main.go create

# Получение баланса
go run cmd/wallet/main.go balance <address>

# Отправка транзакции
go run cmd/wallet/main.go send <from> <to> <amount>
```

## API Endpoints

### Основной API (порт +1000)
- `GET /api/blockchain/status` - Статус блокчейна
- `GET /api/blockchain/blocks` - Список блоков
- `GET /api/blockchain/blocks/{hash}` - Информация о блоке
- `POST /api/blockchain/transactions` - Создание транзакции
- `GET /api/blockchain/transactions/{hash}` - Информация о транзакции
- `GET /api/blockchain/balance/{address}` - Баланс адреса
- `GET /api/blockchain/utxos/{address}` - UTXO адреса

### WebSocket API (порт +1000)
- `ws://localhost:9080/ws` - WebSocket подключение для real-time уведомлений

### DHT API (порт +2000)
- `GET /api/dht/peers` - Список пиров
- `GET /api/dht/stats` - Статистика DHT
- `POST /api/dht/bootstrap` - Bootstrap DHT

### Contract API (порт +2000)
- `POST /api/contracts/deploy` - Деплой контракта
- `POST /api/contracts/call` - Вызов контракта
- `GET /api/contracts/{address}` - Информация о контракте
- `GET /api/contracts` - Список контрактов
- `GET /api/contracts/templates` - Шаблоны контрактов
- `POST /api/contracts/compile` - Компиляция контракта
- `POST /api/contracts/estimate-gas` - Оценка газа
- `GET /api/contracts/gas-report` - Отчет по газу

### Token API (порт +3000)
- `POST /api/tokens/create` - Создание токена
- `POST /api/tokens/transfer` - Перевод токенов
- `POST /api/tokens/approve` - Одобрение токенов
- `POST /api/tokens/transferFrom` - Перевод от имени
- `GET /api/tokens/balance` - Баланс токенов
- `GET /api/tokens/allowance` - Разрешение токенов
- `POST /api/tokens/mint` - Создание токенов
- `POST /api/tokens/burn` - Сжигание токенов
- `GET /api/tokens/stats` - Статистика токенов
- `GET /api/tokens/search` - Поиск токенов
- `POST /api/tokens/export` - Экспорт токенов
- `POST /api/tokens/import` - Импорт токенов
- `GET /api/tokens/holders` - Держатели токенов
- `GET /api/tokens/circulation` - Обращение токенов

### NFT API (порт +4000)
- `POST /api/nft/create-contract` - Создание NFT контракта
- `POST /api/nft/mint` - Создание NFT
- `POST /api/nft/transfer` - Перевод NFT
- `POST /api/nft/approve` - Одобрение NFT
- `POST /api/nft/setApprovalForAll` - Одобрение всех NFT
- `POST /api/nft/transferFrom` - Перевод от имени
- `GET /api/nft/ownerOf` - Владелец NFT
- `GET /api/nft/getApproved` - Одобренный адрес
- `GET /api/nft/isApprovedForAll` - Одобрение всех
- `GET /api/nft/balanceOf` - Баланс NFT
- `POST /api/nft/burn` - Сжигание NFT
- `GET /api/nft/tokenURI` - URI токена
- `GET /api/nft/contracts` - Список контрактов
- `GET /api/nft/tokens` - Список токенов
- `GET /api/nft/search` - Поиск NFT
- `GET /api/nft/stats` - Статистика NFT
- `POST /api/nft/export` - Экспорт NFT
- `POST /api/nft/import` - Импорт NFT

### Sidechain API (порт +5000)
- `POST /api/sidechain/create` - Создание sidechain
- `GET /api/sidechain/list` - Список sidechains
- `GET /api/sidechain/{id}` - Информация о sidechain
- `POST /api/sidechain/add-block` - Добавление блока
- `POST /api/sidechain/add-transaction` - Добавление транзакции
- `POST /api/sidechain/create-asset` - Создание актива
- `GET /api/sidechain/assets` - Список активов
- `POST /api/sidechain/bridge-transaction` - Мостовая транзакция
- `POST /api/sidechain/cross-chain-message` - Кросс-чейн сообщение
- `GET /api/sidechain/messages` - Список сообщений
- `POST /api/sidechain/add-validator` - Добавление валидатора
- `POST /api/sidechain/remove-validator` - Удаление валидатора
- `GET /api/sidechain/validators` - Список валидаторов
- `GET /api/sidechain/stats` - Статистика sidechain
- `GET /api/sidechain/health` - Здоровье sidechain
- `POST /api/sidechain/export` - Экспорт sidechain
- `POST /api/sidechain/import` - Импорт sidechain
- `GET /api/sidechain/transactions` - Транзакции sidechain
- `GET /api/sidechain/blocks` - Блоки sidechain
- `POST /api/sidechain/consensus` - Изменение консенсуса
- `GET /api/sidechain/consensus` - Текущий консенсус
- `POST /api/sidechain/upgrade` - Обновление sidechain

### State Channel API (порт +6000)
- `POST /api/statechannel/create` - Создание канала
- `GET /api/statechannel/list` - Список каналов
- `GET /api/statechannel/{id}` - Информация о канале
- `POST /api/statechannel/deposit` - Депозит в канал
- `POST /api/statechannel/withdraw` - Вывод из канала
- `POST /api/statechannel/update-state` - Обновление состояния
- `POST /api/statechannel/create-transaction` - Создание транзакции
- `POST /api/statechannel/close` - Закрытие канала
- `POST /api/statechannel/initiate-dispute` - Инициация спора
- `POST /api/statechannel/settle` - Разрешение спора
- `GET /api/statechannel/stats` - Статистика канала
- `GET /api/statechannel/history` - История канала

## Демонстрация функций

### Основные демонстрации
```bash
# Демонстрация блокчейна
go run examples/blockchain_demo.go

# Демонстрация P2P сети
go run examples/p2p_network.go

# Демонстрация майнинга
go run examples/mining_demo.go

# Демонстрация кошельков
go run examples/wallet_demo.go

# Демонстрация API
go run examples/api_demo.go
```

### Расширенные демонстрации
```bash
# Демонстрация WebSocket
go run -tags websocket_demo examples/websocket_demo.go

# Демонстрация DHT
go run -tags dht_demo examples/dht_demo.go

# Демонстрация Gossip протокола
go run -tags gossip_demo examples/gossip_demo.go

# Демонстрация Rate Limiting
go run -tags rate_limiting_demo examples/rate_limiting_demo.go

# Демонстрация NAT Traversal
go run -tags nat_traversal_demo examples/nat_traversal_demo.go

# Демонстрация CLI
go run -tags cli_demo examples/cli_demo.go

# Демонстрация конфигурации
go run -tags config_demo examples/config_demo.go

# Демонстрация логирования
go run -tags logging_demo examples/logging_demo.go

# Демонстрация трассировки
go run -tags tracing_demo examples/tracing_demo.go

# Демонстрация безопасности
go run -tags security_demo examples/security_demo.go

# Демонстрация консенсуса
go run -tags consensus_demo examples/consensus_demo.go

# Демонстрация мультиподписей
go run -tags multisig_demo examples/multisig_demo.go

# Демонстрация квантово-устойчивой криптографии
go run -tags quantum_demo examples/quantum_demo.go

# Демонстрация переключения алгоритмов
go run -tags algorithm_switching_demo examples/algorithm_switching_demo.go

# Демонстрация конфигурации алгоритмов
go run -tags algorithm_config_demo examples/algorithm_config_demo.go

# Демонстрация простого переключения
go run -tags simple_algorithm_switch examples/simple_algorithm_switch.go

# Демонстрация смарт-контрактов
go run -tags contract_demo examples/contract_demo.go

# Простая демонстрация смарт-контрактов
go run -tags simple_contracts_demo examples/simple_contracts_demo.go

# Демонстрация системы хранения контрактов
go run -tags contract_storage_demo examples/contract_storage_demo.go

# Тест Contract API
go run -tags contract_api_demo examples/contract_api_demo.go

# Демонстрация токенов
go run -tags token_demo examples/token_demo.go

# Простая демонстрация токенов
go run -tags simple_token_demo examples/simple_token_demo.go

# Тест Token API
go run -tags token_api_demo examples/token_api_demo.go

# Демонстрация NFT
go run -tags nft_demo examples/nft_demo.go

# Простая демонстрация NFT
go run -tags simple_nft_demo examples/simple_nft_demo.go

# Тест NFT API
go run -tags nft_api_demo examples/nft_api_demo.go

# Демонстрация Sidechains
go run -tags sidechain_demo examples/sidechain_demo.go

# Простая демонстрация Sidechains
go run -tags simple_sidechain_demo examples/simple_sidechain_demo.go

# Тест Sidechain API
go run -tags sidechain_api_demo examples/sidechain_api_demo.go

# Демонстрация State Channels
go run -tags statechannel_demo examples/statechannel_demo.go

# Тест State Channel API
go run -tags statechannel_api_demo examples/statechannel_api_demo.go

# Запуск узла с подключением к другим узлам
go run cmd/node/main.go -peers="127.0.0.1:8081,127.0.0.1:8082"
```

## Что уже реализовано

### ✅ Этап 1: Базовая структура блокчейна
- [x] **Блоки и транзакции**
  - Структура блока с заголовком и транзакциями
  - Система UTXO для отслеживания балансов
  - Валидация транзакций и блоков
  - Генерация генезис-блока
- [x] **Криптография**
  - Генерация ключей (ECDSA)
  - Подписание и проверка транзакций
  - Хеширование (SHA-256)
  - Адреса кошельков

### ✅ Этап 2: P2P сеть
- [x] **Базовая P2P сеть**
  - TCP соединения между узлами
  - Протокол обмена сообщениями
  - Обнаружение пиров
  - Синхронизация блокчейна
- [x] **Сетевые сообщения**
  - Handshake между узлами
  - Запрос и передача блоков
  - Запрос и передача транзакций
  - Уведомления о новых блоках

### ✅ Этап 3: Майнинг и консенсус
- [x] **Proof of Work**
  - Алгоритм майнинга с настраиваемой сложностью
  - Валидация proof of work
  - Награда за майнинг
  - Адаптивная сложность
- [x] **Система наград**
  - Награда за блок (50 монет)
  - Комиссии за транзакции
  - Валидация наград

### ✅ Этап 4: REST API
- [x] **HTTP API сервер**
  - Получение информации о блокчейне
  - Создание и отправка транзакций
  - Получение баланса адреса
  - Получение UTXO
- [x] **API endpoints**
  - `GET /api/blockchain/status` - статус блокчейна
  - `GET /api/blockchain/blocks` - список блоков
  - `POST /api/blockchain/transactions` - создание транзакции
  - `GET /api/blockchain/balance/{address}` - баланс адреса

### ✅ Этап 5: Система кошельков
- [x] **Кошелек**
  - Генерация ключей
  - Создание транзакций
  - Подписание транзакций
  - Управление адресами
- [x] **CLI интерфейс**
  - Создание кошелька
  - Получение баланса
  - Отправка транзакций
  - Просмотр истории

### ✅ Этап 6: Тестирование
- [x] **Unit тесты**
  - Тесты для блокчейна
  - Тесты для P2P сети
  - Тесты для майнинга
  - Тесты для API
- [x] **Интеграционные тесты**
  - Тесты полного цикла
  - Тесты сети из нескольких узлов
  - Тесты производительности

### ✅ Этап 7: Документация
- [x] **README**
  - Описание проекта
  - Инструкции по установке
  - Примеры использования
  - API документация
- [x] **Комментарии в коде**
  - Документация функций
  - Примеры использования
  - Описание алгоритмов

### ✅ Этап 8: Оптимизация
- [x] **Производительность**
  - Оптимизация майнинга
  - Кэширование блоков
  - Параллельная обработка транзакций
  - Оптимизация памяти
- [x] **Мониторинг**
  - Метрики производительности
  - Логирование
  - Профилирование

### ✅ Этап 9: Персистентность
- [x] **Хранение данных**
  - BadgerDB для хранения блоков
  - BadgerDB для хранения UTXO
  - Сериализация данных
  - Восстановление состояния
- [x] **Кэширование**
  - LRU кэш для блоков
  - LRU кэш для UTXO
  - Настраиваемый размер кэша
  - Статистика кэша

### ✅ Этап 10: Расширенная P2P сеть
- [x] **WebSocket для real-time уведомлений**
  - WebSocket сервер для real-time обновлений
  - Уведомления о новых блоках и транзакциях
  - Подписка на события блокчейна
  - Автоматическое переподключение
- [x] **Улучшенная система peer discovery**
  - DHT (Distributed Hash Table) для децентрализованного обнаружения пиров
  - Kademlia-подобный алгоритм для маршрутизации
  - Bootstrap узлы для первоначального подключения
  - Автоматическое обнаружение и подключение к пирам
- [x] **Gossip протокол**
  - Эффективное распространение данных по сети
  - Случайный выбор узлов для отправки (fanout)
  - TTL (Time To Live) для предотвращения зацикливания
  - Heartbeat для поддержания соединений
  - Система репутации узлов
  - Автоматическая очистка неактивных узлов
- [x] **Rate Limiting**
  - Token Bucket алгоритм для API
  - Sliding Window алгоритм для P2P
  - Настраиваемые лимиты и окна времени
  - Поддержка множественных клиентов
  - Статистика и мониторинг
- [x] **NAT Traversal**
  - Определение типа NAT (STUN)
  - Hole punching для установки соединений
  - Поддержка различных типов NAT
  - Keep-alive для поддержания соединений
  - Автоматический выбор стратегии соединения

### ✅ Этап 11: CLI и управление
- [x] **Расширенный CLI интерфейс**
  - Интерактивная консоль для управления узлом
  - Команды для мониторинга сети
  - Управление кошельками через CLI
- [x] **Конфигурация узла**
  - YAML/JSON конфигурационные файлы
  - Переменные окружения
  - Валидация конфигурации
- [x] **Логирование и отладка**
  - Структурированные логи (JSON)
  - Уровни логирования
  - Трассировка транзакций

### ✅ Этап 12: Безопасность и консенсус
- [x] **Улучшенная безопасность**
  - Защита от атак 51% с мониторингом хеш-рейта
  - Валидация входных данных для блоков и транзакций
  - Улучшенный Rate limiting для API с разными алгоритмами
- [x] **Альтернативные алгоритмы консенсуса**
  - Proof of Stake (PoS) с валидаторами и стейкингом
  - Delegated Proof of Stake (DPoS) с делегатами и голосованием
  - Сравнение производительности алгоритмов консенсуса
- [x] **Криптографические улучшения**
  - Поддержка разных алгоритмов подписи (ECDSA, Ed25519, RSA, Schnorr)
  - Система мультиподписей с настраиваемыми порогами
  - Менеджер алгоритмов подписи
- [x] **Квантово-устойчивая криптография**
  - SPHINCS+ - Stateless hash-based signatures
  - Dilithium - Lattice-based signatures
  - Falcon - Lattice-based signatures
  - XMSS - Stateful hash-based signatures
  - LMS - Leighton-Micali signatures
  - Сравнение производительности пост-квантовых алгоритмов

### ✅ Этап 13: Расширенные функции блокчейна
- [x] **Смарт-контракты**
  - Виртуальная машина (VM) с стек-архитектурой
  - Язык программирования контрактов (20+ операций)
  - Система газа с трекингом и оптимизацией
  - Персистентное хранение состояния контрактов (BadgerDB)
  - HTTP API для управления контрактами (12 endpoints)
  - Шаблоны контрактов и компилятор
  - Управление хранилищем контрактов (SLOAD/SSTORE)
  - Статистика и мониторинг контрактов
- [x] **Система токенов (ERC-20)**
  - Полная реализация стандарта ERC-20
  - Создание, перевод, одобрение токенов
  - Система разрешений (allowances)
  - Создание и сжигание токенов (mint/burn)
  - HTTP API для управления токенами (14 endpoints)
  - Статистика и поиск токенов
  - Экспорт/импорт токенов
- [x] **Система NFT (ERC-721)**
  - Полная реализация стандарта ERC-721
  - Создание NFT контрактов и коллекций
  - Создание уникальных NFT токенов с метаданными
  - Переводы и система одобрений NFT
  - HTTP API для управления NFT (18 endpoints)
  - Поиск и фильтрация NFT по атрибутам
  - Статистика коллекций и владельцев
  - Экспорт/импорт NFT контрактов
- [x] **Sidechains (боковые цепи)**
  - Создание и управление sidechains
  - Различные алгоритмы консенсуса (PoW, PoS, DPoS, PBFT)
  - Система активов в sidechains (нативные, токены, NFT, мостовые)
  - Мостовые транзакции между цепями
  - Кросс-чейн сообщения
  - Управление валидаторами
  - HTTP API для управления sidechains (22 endpoints)
  - Статистика и мониторинг sidechains
- [x] **State Channels (каналы состояния)**
  - Открытие и закрытие каналов
  - Обновление состояния каналов
  - Система депозитов и выводов
  - Разрешение споров
  - Поддержка различных типов каналов (payment, micropayment, gaming, prediction, custom)
  - HTTP API для управления state channels (12 endpoints)
  - Статистика и мониторинг каналов

## Следующие этапы развития

### 🌍 Этап 14: Экосистема и интеграция
- [ ] **API Gateway**
  - GraphQL API
  - Webhook поддержка
  - API версионирование
- [ ] **Интеграции**
  - Docker контейнеризация
  - Kubernetes deployment
  - CI/CD pipeline
- [ ] **Документация и примеры**
  - Swagger/OpenAPI документация
  - SDK для разных языков
  - Туториалы и примеры использования

### 🎯 Рекомендуемый следующий шаг
**Начать с Этапа 14: Экосистема и интеграция**

Все основные функции блокчейна, CLI управление, безопасность и расширенные функции уже реализованы! Следующий логичный шаг:
1. Создать API Gateway с GraphQL
2. Добавить Docker контейнеризацию
3. Создать SDK для разных языков
4. Улучшить документацию

✅ **Завершено**
- Полная система блокчейна с оптимизированным майнингом
- P2P сеть и REST API
- WebSocket уведомления для real-time обновлений
- DHT (Distributed Hash Table) для децентрализованного peer discovery
- Система кошельков с CLI
- Комплексное тестирование и документация
- Персистентное хранение с BadgerDB
- Оптимизация производительности (кэширование, параллельная обработка)
- Мониторинг и метрики (Prometheus, логирование, профилирование)
- Удален устаревший алгоритм майнинга (оставлен только оптимизированный)

---

**Статус проекта**: Этапы 1-13 завершены ✅ (Все основные функции + CLI управление + Безопасность и консенсус + Квантово-устойчивая криптография + Смарт-контракты + Система токенов + NFT система + Sidechains + State Channels реализованы)
**Последнее обновление**: 2025-09-20