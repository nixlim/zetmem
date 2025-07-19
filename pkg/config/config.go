package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	ChromaDB   ChromaDBConfig   `yaml:"chromadb"`
	LiteLLM    LiteLLMConfig    `yaml:"litellm"`
	Embedding  EmbeddingConfig  `yaml:"embedding"`
	Evolution  EvolutionConfig  `yaml:"evolution"`
	Prompts    PromptsConfig    `yaml:"prompts"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Onboarding OnboardingConfig `yaml:"onboarding"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port           int    `yaml:"port"`
	LogLevel       string `yaml:"log_level"`
	MaxRequestSize string `yaml:"max_request_size"`
}

// ChromaDBConfig represents ChromaDB configuration
type ChromaDBConfig struct {
	URL        string `yaml:"url"`
	Collection string `yaml:"collection"`
	BatchSize  int    `yaml:"batch_size"`
}

// LiteLLMConfig represents LiteLLM configuration
type LiteLLMConfig struct {
	DefaultModel   string        `yaml:"default_model"`
	FallbackModels []string      `yaml:"fallback_models"`
	MaxRetries     int           `yaml:"max_retries"`
	Timeout        time.Duration `yaml:"timeout"`
	RateLimit      int           `yaml:"rate_limit"`
}

// EmbeddingConfig represents embedding service configuration
type EmbeddingConfig struct {
	Service   string `yaml:"service"`
	Model     string `yaml:"model"`
	BatchSize int    `yaml:"batch_size"`
	URL       string `yaml:"url"`
}

// EvolutionConfig represents memory evolution configuration
type EvolutionConfig struct {
	Enabled     bool   `yaml:"enabled"`
	Schedule    string `yaml:"schedule"`
	BatchSize   int    `yaml:"batch_size"`
	WorkerCount int    `yaml:"worker_count"`
}

// PromptsConfig represents prompt management configuration
type PromptsConfig struct {
	Directory    string `yaml:"directory"`
	CacheEnabled bool   `yaml:"cache_enabled"`
	HotReload    bool   `yaml:"hot_reload"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	MetricsPort   int     `yaml:"metrics_port"`
	EnableTracing bool    `yaml:"enable_tracing"`
	SampleRate    float64 `yaml:"sample_rate"`
}

// OnboardingConfig represents agent onboarding configuration
type OnboardingConfig struct {
	StrategyGuidePath string `yaml:"strategy_guide_path"`
	MaxFileSize       int64  `yaml:"max_file_size"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{
		// Set defaults
		Server: ServerConfig{
			Port:           getEnvInt("ZETMEM_PORT", 8080),
			LogLevel:       getEnvString("ZETMEM_LOG_LEVEL", "info"),
			MaxRequestSize: getEnvString("ZETMEM_MAX_REQUEST_SIZE", "10MB"),
		},
		ChromaDB: ChromaDBConfig{
			URL:        getEnvString("CHROMADB_HOST", "http://localhost:8000"),
			Collection: getEnvString("CHROMADB_COLLECTION", "zetmem_memories"),
			BatchSize:  getEnvInt("CHROMADB_BATCH_SIZE", 100),
		},
		LiteLLM: LiteLLMConfig{
			DefaultModel:   getEnvString("LITELLM_DEFAULT_MODEL", "gpt-4-turbo"),
			FallbackModels: []string{"gpt-3.5-turbo", "claude-2"},
			MaxRetries:     getEnvInt("LITELLM_MAX_RETRIES", 3),
			Timeout:        time.Duration(getEnvInt("LITELLM_TIMEOUT_SECONDS", 30)) * time.Second,
			RateLimit:      getEnvInt("LITELLM_RATE_LIMIT", 60),
		},
		Embedding: EmbeddingConfig{
			Service:   getEnvString("EMBEDDING_SERVICE", "sentence-transformers"),
			Model:     getEnvString("EMBEDDING_MODEL", "all-MiniLM-L6-v2"),
			BatchSize: getEnvInt("EMBEDDING_BATCH_SIZE", 32),
			URL:       getEnvString("EMBEDDING_SERVICE_URL", "http://localhost:8005"),
		},
		Evolution: EvolutionConfig{
			Enabled:     getEnvBool("ZETMEM_EVOLUTION_ENABLED", true),
			Schedule:    getEnvString("ZETMEM_EVOLUTION_SCHEDULE", "0 2 * * *"),
			BatchSize:   getEnvInt("ZETMEM_EVOLUTION_BATCH_SIZE", 50),
			WorkerCount: getEnvInt("ZETMEM_EVOLUTION_WORKER_COUNT", 3),
		},
		Prompts: PromptsConfig{
			Directory:    getEnvString("ZETMEM_PROMPTS_PATH", "/app/prompts"),
			CacheEnabled: getEnvBool("ZETMEM_PROMPTS_CACHE_ENABLED", true),
			HotReload:    getEnvBool("ZETMEM_PROMPTS_HOT_RELOAD", true),
		},
		Monitoring: MonitoringConfig{
			MetricsPort:   getEnvInt("ZETMEM_METRICS_PORT", 9090),
			EnableTracing: getEnvBool("ZETMEM_METRICS_ENABLED", true),
			SampleRate:    getEnvFloat("ZETMEM_TRACING_SAMPLE_RATE", 0.1),
		},
		Onboarding: OnboardingConfig{
			StrategyGuidePath: getEnvString("ZETMEM_STRATEGY_GUIDE_PATH", "ZETMEM_ONBOARDING_STRATEGY.md"),
			MaxFileSize:       getEnvInt64("ZETMEM_STRATEGY_GUIDE_MAX_SIZE", 1024*1024), // 1MB default
		},
	}

	// Load from YAML file if provided
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.ChromaDB.URL == "" {
		return fmt.Errorf("ChromaDB URL is required")
	}

	if c.LiteLLM.DefaultModel == "" {
		return fmt.Errorf("LiteLLM default model is required")
	}

	if c.LiteLLM.MaxRetries < 0 {
		return fmt.Errorf("LiteLLM max retries must be non-negative")
	}

	return nil
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}
