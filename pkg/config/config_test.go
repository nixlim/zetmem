package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Test loading with defaults (no config file)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Test default values
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Server.LogLevel)
	}

	if cfg.ChromaDB.Collection != "amem_memories" {
		t.Errorf("Expected default collection 'amem_memories', got %s", cfg.ChromaDB.Collection)
	}

	if cfg.LiteLLM.DefaultModel != "gpt-4-turbo" {
		t.Errorf("Expected default model 'gpt-4-turbo', got %s", cfg.LiteLLM.DefaultModel)
	}

	if cfg.LiteLLM.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", cfg.LiteLLM.MaxRetries)
	}

	if cfg.LiteLLM.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", cfg.LiteLLM.Timeout)
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("AMEM_PORT", "9090")
	os.Setenv("AMEM_LOG_LEVEL", "debug")
	os.Setenv("CHROMADB_COLLECTION", "test_collection")
	defer func() {
		os.Unsetenv("AMEM_PORT")
		os.Unsetenv("AMEM_LOG_LEVEL")
		os.Unsetenv("CHROMADB_COLLECTION")
	}()

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port from env var 9090, got %d", cfg.Server.Port)
	}

	if cfg.Server.LogLevel != "debug" {
		t.Errorf("Expected log level from env var 'debug', got %s", cfg.Server.LogLevel)
	}

	if cfg.ChromaDB.Collection != "test_collection" {
		t.Errorf("Expected collection from env var 'test_collection', got %s", cfg.ChromaDB.Collection)
	}
}

func TestConfigValidation(t *testing.T) {
	// Test invalid port
	cfg := &Config{
		Server:   ServerConfig{Port: -1},
		ChromaDB: ChromaDBConfig{URL: "http://localhost:8000"},
		LiteLLM:  LiteLLMConfig{DefaultModel: "gpt-4"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid port")
	}

	// Test missing ChromaDB URL
	cfg = &Config{
		Server:   ServerConfig{Port: 8080},
		ChromaDB: ChromaDBConfig{URL: ""},
		LiteLLM:  LiteLLMConfig{DefaultModel: "gpt-4"},
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing ChromaDB URL")
	}

	// Test missing LiteLLM model
	cfg = &Config{
		Server:   ServerConfig{Port: 8080},
		ChromaDB: ChromaDBConfig{URL: "http://localhost:8000"},
		LiteLLM:  LiteLLMConfig{DefaultModel: ""},
	}

	err = cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for missing LiteLLM model")
	}

	// Test valid config
	cfg = &Config{
		Server:   ServerConfig{Port: 8080},
		ChromaDB: ChromaDBConfig{URL: "http://localhost:8000"},
		LiteLLM:  LiteLLMConfig{DefaultModel: "gpt-4", MaxRetries: 3},
	}

	err = cfg.Validate()
	if err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Test getEnvString
	os.Setenv("TEST_STRING", "test_value")
	defer os.Unsetenv("TEST_STRING")

	value := getEnvString("TEST_STRING", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %s", value)
	}

	value = getEnvString("NON_EXISTENT", "default")
	if value != "default" {
		t.Errorf("Expected 'default', got %s", value)
	}

	// Test getEnvInt
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	intValue := getEnvInt("TEST_INT", 0)
	if intValue != 42 {
		t.Errorf("Expected 42, got %d", intValue)
	}

	intValue = getEnvInt("NON_EXISTENT", 10)
	if intValue != 10 {
		t.Errorf("Expected 10, got %d", intValue)
	}

	// Test getEnvBool
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	boolValue := getEnvBool("TEST_BOOL", false)
	if !boolValue {
		t.Errorf("Expected true, got %v", boolValue)
	}

	boolValue = getEnvBool("NON_EXISTENT", false)
	if boolValue {
		t.Errorf("Expected false, got %v", boolValue)
	}
}
