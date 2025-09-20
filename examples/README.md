# MiroChain Examples

Эта папка содержит примеры использования различных компонентов MiroChain.

## Запуск примеров

Каждый пример имеет build tag, поэтому для запуска нужно указать соответствующий тег:

### Базовое использование
```bash
go run -tags basic_usage examples/basic_usage.go
```

### Полная интеграция
```bash
go run -tags full_integration examples/full_integration.go
```

### Демо полной интеграции
```bash
go run -tags full_integration_demo examples/full_integration_demo.go
```

### Метрики
```bash
go run -tags metrics_usage examples/metrics_usage.go
```

### Майнинг
```bash
go run -tags mining_demo examples/mining_demo.go
go run -tags mining_example examples/mining_example.go
```

### P2P сеть
```bash
go run -tags p2p_network examples/p2p_network.go
```

### Персистентное хранение
```bash
go run -tags persistent_blockchain examples/persistent_blockchain.go
```

### Профилирование
```bash
go run -tags profiling_usage examples/profiling_usage.go
```

### Prometheus метрики
```bash
go run -tags prometheus_metrics examples/prometheus_metrics.go
```

### Обработчик транзакций
```bash
go run -tags transaction_processor_demo examples/transaction_processor_demo.go
```

## Описание примеров

- **basic_usage.go** - Базовый пример создания блокчейна, кошельков и транзакций
- **full_integration.go** - Полная интеграция всех компонентов системы
- **metrics_usage.go** - Пример использования системы метрик
- **mining_demo.go** - Демонстрация майнинга блоков
- **p2p_network.go** - Пример P2P сети
- **persistent_blockchain.go** - Пример персистентного хранения
- **profiling_usage.go** - Пример профилирования производительности
- **prometheus_metrics.go** - Пример интеграции с Prometheus
- **transaction_processor_demo.go** - Демонстрация параллельной обработки транзакций
- **full_integration_demo.go** - Основное демо с полной интеграцией

## Примечания

- Все примеры используют build tags для избежания конфликтов при компиляции
- Некоторые примеры создают временные файлы (например, для хранения данных)
- Примеры с метриками запускают HTTP серверы на портах 8080 и 6060
