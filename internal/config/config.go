package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config представляет конфигурацию узла
type Config struct {
	Node       NodeConfig       `yaml:"node"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
	Network    NetworkConfig    `yaml:"network"`
	Wallet     WalletConfig     `yaml:"wallet"`
	Logging    LoggingConfig    `yaml:"logging"`
	Mining     MiningConfig     `yaml:"mining"`
}

// NodeConfig конфигурация узла
type NodeConfig struct {
	ID          string        `yaml:"id"`
	Address     string        `yaml:"address"`
	Port        int           `yaml:"port"`
	Peers       []string      `yaml:"peers"`
	Mining      bool          `yaml:"mining"`
	DataDir     string        `yaml:"data_dir"`
	MaxPeers    int           `yaml:"max_peers"`
	PeerTimeout time.Duration `yaml:"peer_timeout"`
}

// BlockchainConfig конфигурация блокчейна
type BlockchainConfig struct {
	Difficulty    int           `yaml:"difficulty"`
	BlockTime     time.Duration `yaml:"block_time"`
	MaxBlockSize  int           `yaml:"max_block_size"`
	MaxTxPerBlock int           `yaml:"max_tx_per_block"`
	GenesisReward int64         `yaml:"genesis_reward"`
	MinTxFee      int64         `yaml:"min_tx_fee"`
}

// NetworkConfig конфигурация сети
type NetworkConfig struct {
	WebSocketPort   int           `yaml:"websocket_port"`
	DHTPort         int           `yaml:"dht_port"`
	BootstrapNodes  []string      `yaml:"bootstrap_nodes"`
	ConnectionLimit int           `yaml:"connection_limit"`
	MessageTimeout  time.Duration `yaml:"message_timeout"`
	PingInterval    time.Duration `yaml:"ping_interval"`
}

// WalletConfig конфигурация кошелька
type WalletConfig struct {
	DataDir     string `yaml:"data_dir"`
	DefaultKey  string `yaml:"default_key"`
	KeyStrength int    `yaml:"key_strength"`
}

// LoggingConfig конфигурация логирования
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// MiningConfig конфигурация майнинга
type MiningConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Threads    int           `yaml:"threads"`
	Timeout    time.Duration `yaml:"timeout"`
	Reward     int64         `yaml:"reward"`
	MinTxCount int           `yaml:"min_tx_count"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			ID:          "",
			Address:     "localhost",
			Port:        8080,
			Peers:       []string{},
			Mining:      true,
			DataDir:     "data",
			MaxPeers:    50,
			PeerTimeout: 30 * time.Second,
		},
		Blockchain: BlockchainConfig{
			Difficulty:    4,
			BlockTime:     10 * time.Second,
			MaxBlockSize:  1024 * 1024, // 1MB
			MaxTxPerBlock: 1000,
			GenesisReward: 100,
			MinTxFee:      1,
		},
		Network: NetworkConfig{
			WebSocketPort:   9080,
			DHTPort:         10080,
			BootstrapNodes:  []string{},
			ConnectionLimit: 100,
			MessageTimeout:  5 * time.Second,
			PingInterval:    30 * time.Second,
		},
		Wallet: WalletConfig{
			DataDir:     "wallet_data",
			DefaultKey:  "",
			KeyStrength: 256,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		},
		Mining: MiningConfig{
			Enabled:    true,
			Threads:    4,
			Timeout:    30 * time.Second,
			Reward:     50,
			MinTxCount: 1,
		},
	}
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(filename string) (*Config, error) {
	// Если файл не существует, создаем конфигурацию по умолчанию
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := config.Save(filename); err != nil {
			return nil, fmt.Errorf("failed to save default config: %v", err)
		}
		return config, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Применяем значения по умолчанию для пустых полей
	config.applyDefaults()

	return &config, nil
}

// Save сохраняет конфигурацию в файл
func (c *Config) Save(filename string) error {
	// Создаем директорию если не существует
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// applyDefaults применяет значения по умолчанию
func (c *Config) applyDefaults() {
	defaults := DefaultConfig()

	if c.Node.ID == "" {
		c.Node.ID = defaults.Node.ID
	}
	if c.Node.Address == "" {
		c.Node.Address = defaults.Node.Address
	}
	if c.Node.Port == 0 {
		c.Node.Port = defaults.Node.Port
	}
	if c.Node.DataDir == "" {
		c.Node.DataDir = defaults.Node.DataDir
	}
	if c.Node.MaxPeers == 0 {
		c.Node.MaxPeers = defaults.Node.MaxPeers
	}
	if c.Node.PeerTimeout == 0 {
		c.Node.PeerTimeout = defaults.Node.PeerTimeout
	}

	if c.Blockchain.Difficulty == 0 {
		c.Blockchain.Difficulty = defaults.Blockchain.Difficulty
	}
	if c.Blockchain.BlockTime == 0 {
		c.Blockchain.BlockTime = defaults.Blockchain.BlockTime
	}
	if c.Blockchain.MaxBlockSize == 0 {
		c.Blockchain.MaxBlockSize = defaults.Blockchain.MaxBlockSize
	}
	if c.Blockchain.MaxTxPerBlock == 0 {
		c.Blockchain.MaxTxPerBlock = defaults.Blockchain.MaxTxPerBlock
	}
	if c.Blockchain.GenesisReward == 0 {
		c.Blockchain.GenesisReward = defaults.Blockchain.GenesisReward
	}
	if c.Blockchain.MinTxFee == 0 {
		c.Blockchain.MinTxFee = defaults.Blockchain.MinTxFee
	}

	if c.Network.WebSocketPort == 0 {
		c.Network.WebSocketPort = defaults.Network.WebSocketPort
	}
	if c.Network.DHTPort == 0 {
		c.Network.DHTPort = defaults.Network.DHTPort
	}
	if c.Network.ConnectionLimit == 0 {
		c.Network.ConnectionLimit = defaults.Network.ConnectionLimit
	}
	if c.Network.MessageTimeout == 0 {
		c.Network.MessageTimeout = defaults.Network.MessageTimeout
	}
	if c.Network.PingInterval == 0 {
		c.Network.PingInterval = defaults.Network.PingInterval
	}

	if c.Wallet.DataDir == "" {
		c.Wallet.DataDir = defaults.Wallet.DataDir
	}
	if c.Wallet.KeyStrength == 0 {
		c.Wallet.KeyStrength = defaults.Wallet.KeyStrength
	}

	if c.Logging.Level == "" {
		c.Logging.Level = defaults.Logging.Level
	}
	if c.Logging.Format == "" {
		c.Logging.Format = defaults.Logging.Format
	}
	if c.Logging.Output == "" {
		c.Logging.Output = defaults.Logging.Output
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = defaults.Logging.MaxSize
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = defaults.Logging.MaxBackups
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = defaults.Logging.MaxAge
	}

	if c.Mining.Threads == 0 {
		c.Mining.Threads = defaults.Mining.Threads
	}
	if c.Mining.Timeout == 0 {
		c.Mining.Timeout = defaults.Mining.Timeout
	}
	if c.Mining.Reward == 0 {
		c.Mining.Reward = defaults.Mining.Reward
	}
	if c.Mining.MinTxCount == 0 {
		c.Mining.MinTxCount = defaults.Mining.MinTxCount
	}
}

// Validate проверяет конфигурацию
func (c *Config) Validate() error {
	if c.Node.Port <= 0 || c.Node.Port > 65535 {
		return fmt.Errorf("invalid node port: %d", c.Node.Port)
	}

	if c.Node.MaxPeers < 0 {
		return fmt.Errorf("invalid max peers: %d", c.Node.MaxPeers)
	}

	if c.Blockchain.Difficulty < 0 {
		return fmt.Errorf("invalid difficulty: %d", c.Blockchain.Difficulty)
	}

	if c.Blockchain.BlockTime <= 0 {
		return fmt.Errorf("invalid block time: %v", c.Blockchain.BlockTime)
	}

	if c.Network.WebSocketPort <= 0 || c.Network.WebSocketPort > 65535 {
		return fmt.Errorf("invalid websocket port: %d", c.Network.WebSocketPort)
	}

	if c.Network.DHTPort <= 0 || c.Network.DHTPort > 65535 {
		return fmt.Errorf("invalid DHT port: %d", c.Network.DHTPort)
	}

	if c.Mining.Threads < 0 {
		return fmt.Errorf("invalid mining threads: %d", c.Mining.Threads)
	}

	return nil
}

// GetConfigPath возвращает путь к файлу конфигурации
func GetConfigPath() string {
	// Проверяем переменные окружения
	if path := os.Getenv("MIROCHAIN_CONFIG"); path != "" {
		return path
	}

	// Проверяем текущую директорию
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}

	// Проверяем домашнюю директорию
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".mirochain", "config.yaml")
		return configPath
	}

	// По умолчанию в текущей директории
	return "config.yaml"
}
