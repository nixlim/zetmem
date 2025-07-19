package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amem/mcp-server/pkg/config"
	"github.com/amem/mcp-server/pkg/mcp"
	"github.com/amem/mcp-server/pkg/memory"
	"github.com/amem/mcp-server/pkg/monitoring"
	"github.com/amem/mcp-server/pkg/scheduler"
	"github.com/amem/mcp-server/pkg/services"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Parse command line flags
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		envFile    = flag.String("env", ".env", "Path to environment file")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Load environment variables
	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			// Don't fail if .env file doesn't exist
			fmt.Fprintf(os.Stderr, "Warning: Could not load .env file: %v\n", err)
		}
	}

	// Initialize logger
	logger, err := initLogger(*logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting A-MEM MCP Server",
		zap.String("version", "1.0.0"),
		zap.String("config_path", *configPath))

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("chromadb_url", cfg.ChromaDB.URL),
		zap.String("default_model", cfg.LiteLLM.DefaultModel),
		zap.Bool("evolution_enabled", cfg.Evolution.Enabled))

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	logger.Info("Initializing services...")

	// Initialize LiteLLM service
	llmService := services.NewLiteLLMService(cfg.LiteLLM, logger.Named("litellm"))

	// Initialize embedding service
	embeddingService := services.NewEmbeddingService(cfg.Embedding, logger.Named("embedding"))

	// Initialize ChromaDB service
	chromaService := services.NewChromaDBService(cfg.ChromaDB, logger.Named("chromadb"))

	// Initialize ChromaDB collection
	if err := chromaService.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize ChromaDB", zap.Error(err))
	}

	// Initialize prompt manager
	_ = services.NewPromptManager(cfg.Prompts, logger.Named("prompts"))

	// Initialize workspace service
	workspaceService := services.NewWorkspaceService(chromaService, logger.Named("workspace"))

	// Initialize memory system
	memorySystem := memory.NewSystem(logger.Named("memory"), llmService, chromaService, embeddingService, workspaceService)

	// Initialize evolution manager
	evolutionManager := memory.NewEvolutionManager(memorySystem, logger.Named("evolution"))

	// Initialize monitoring
	metricsServer := monitoring.NewMetricsServer(cfg.Monitoring.MetricsPort, logger.Named("metrics"))
	go func() {
		if err := metricsServer.Start(ctx); err != nil {
			logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	// Initialize scheduler
	taskScheduler := scheduler.NewScheduler(cfg.Evolution, evolutionManager, logger.Named("scheduler"))
	if err := taskScheduler.Start(ctx); err != nil {
		logger.Error("Failed to start scheduler", zap.Error(err))
	}

	// Initialize MCP server
	mcpServer := mcp.NewServer(logger.Named("mcp"))

	// Register tools
	logger.Info("Registering MCP tools...")

	storeTool := memory.NewStoreCodingMemoryTool(memorySystem, logger.Named("store_tool"))
	mcpServer.RegisterTool(storeTool)

	retrieveTool := memory.NewRetrieveRelevantMemoriesTool(memorySystem, logger.Named("retrieve_tool"))
	mcpServer.RegisterTool(retrieveTool)

	evolveTool := memory.NewEvolveMemoryNetworkTool(evolutionManager, logger.Named("evolve_tool"))
	mcpServer.RegisterTool(evolveTool)

	// Register workspace management tools
	workspaceInitTool := memory.NewWorkspaceInitTool(workspaceService, logger.Named("workspace_init_tool"))
	mcpServer.RegisterTool(workspaceInitTool)

	workspaceCreateTool := memory.NewWorkspaceCreateTool(workspaceService, logger.Named("workspace_create_tool"))
	mcpServer.RegisterTool(workspaceCreateTool)

	workspaceRetrieveTool := memory.NewWorkspaceRetrieveTool(workspaceService, logger.Named("workspace_retrieve_tool"))
	mcpServer.RegisterTool(workspaceRetrieveTool)

	// Register onboarding tool
	performOnboardingTool := memory.NewPerformOnboardingTool(workspaceService, cfg.Onboarding, logger.Named("perform_onboarding_tool"))
	mcpServer.RegisterTool(performOnboardingTool)

	logger.Info("All tools registered successfully")

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start MCP server
	logger.Info("Starting MCP server...")
	if err := mcpServer.Start(ctx); err != nil && err != context.Canceled {
		logger.Fatal("MCP server failed", zap.Error(err))
	}

	logger.Info("A-MEM MCP Server shutdown complete")
}

// initLogger initializes the logger with the specified level
func initLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()
}
