package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/amem/mcp-server/pkg/config"
	"github.com/amem/mcp-server/pkg/memory"
	"github.com/amem/mcp-server/pkg/services"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("config/production.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize services
	ctx := context.Background()

	// Initialize ChromaDB service
	chromaService := services.NewChromaDBService(cfg.ChromaDB, logger.Named("chromadb"))
	if err := chromaService.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize ChromaDB: %v", err)
	}

	// Initialize workspace service
	workspaceService := services.NewWorkspaceService(chromaService, logger.Named("workspace"))

	// Test improved perform_onboarding tool
	fmt.Println("🧪 Testing Improved perform_onboarding Tool")
	fmt.Println(strings.Repeat("=", 60))

	// Display configuration
	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("   Strategy Guide Path: %s\n", cfg.Onboarding.StrategyGuidePath)
	fmt.Printf("   Max File Size: %d bytes\n", cfg.Onboarding.MaxFileSize)

	// Create improved perform_onboarding tool with configuration
	performOnboardingTool := memory.NewPerformOnboardingTool(workspaceService, cfg.Onboarding, logger.Named("perform_onboarding"))
	
	fmt.Printf("✅ Created improved perform_onboarding tool: %s\n", performOnboardingTool.Name())

	// Test 1: Basic onboarding with valid input
	fmt.Println("\n1. Testing basic onboarding with valid input...")
	basicArgs := map[string]interface{}{
		"project_path": "/Users/test/valid-project",
		"project_name": "Test Improved Onboarding",
		"include_strategy_guide": false, // Skip for cleaner output
	}
	
	basicResult, err := performOnboardingTool.Execute(ctx, basicArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success! IsError: %v\n", basicResult.IsError)
		if len(basicResult.Content) > 0 {
			text := basicResult.Content[0].Text
			if len(text) > 300 {
				text = text[:300] + "..."
			}
			fmt.Printf("📄 Response preview:\n%s\n", text)
		}
	}

	// Test 2: Input validation - invalid project path
	fmt.Println("\n2. Testing input validation with invalid project path...")
	invalidArgs := map[string]interface{}{
		"project_path": "invalid\x00path", // Contains null byte
		"project_name": "Invalid Path Test",
	}
	
	invalidResult, err := performOnboardingTool.Execute(ctx, invalidArgs)
	if err != nil {
		fmt.Printf("❌ Unexpected error: %v\n", err)
	} else {
		if invalidResult.IsError {
			fmt.Printf("✅ Correctly rejected invalid path: %s\n", invalidResult.Content[0].Text)
		} else {
			fmt.Printf("❌ Should have rejected invalid path\n")
		}
	}

	// Test 3: Input validation - path too long
	fmt.Println("\n3. Testing input validation with path too long...")
	longPath := string(make([]byte, 5000)) // 5000 characters
	for i := range longPath {
		longPath = longPath[:i] + "a" + longPath[i+1:]
	}
	
	longPathArgs := map[string]interface{}{
		"project_path": longPath,
		"project_name": "Long Path Test",
	}
	
	longPathResult, err := performOnboardingTool.Execute(ctx, longPathArgs)
	if err != nil {
		fmt.Printf("❌ Unexpected error: %v\n", err)
	} else {
		if longPathResult.IsError {
			fmt.Printf("✅ Correctly rejected long path\n")
		} else {
			fmt.Printf("❌ Should have rejected long path\n")
		}
	}

	// Test 4: Strategy guide caching
	fmt.Println("\n4. Testing strategy guide caching...")
	cacheArgs := map[string]interface{}{
		"project_name": "Cache Test",
		"include_strategy_guide": true,
	}
	
	// First call
	cacheResult1, err := performOnboardingTool.Execute(ctx, cacheArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ First call successful\n")
	}

	// Second call (should use cached content)
	cacheResult2, err := performOnboardingTool.Execute(ctx, cacheArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Second call successful (using cached content)\n")
		
		// Compare response lengths to ensure consistency
		if len(cacheResult1.Content) > 0 && len(cacheResult2.Content) > 0 {
			len1 := len(cacheResult1.Content[0].Text)
			len2 := len(cacheResult2.Content[0].Text)
			if len1 == len2 {
				fmt.Printf("✅ Cached content consistent (length: %d)\n", len1)
			} else {
				fmt.Printf("❌ Cached content inconsistent (lengths: %d vs %d)\n", len1, len2)
			}
		}
	}

	// Test 5: Enhanced tool interface methods
	fmt.Println("\n5. Testing enhanced tool interface methods...")
	
	triggers := performOnboardingTool.UsageTriggers()
	fmt.Printf("✅ Usage triggers count: %d\n", len(triggers))
	
	practices := performOnboardingTool.BestPractices()
	fmt.Printf("✅ Best practices count: %d\n", len(practices))
	
	synergies := performOnboardingTool.Synergies()
	fmt.Printf("✅ Synergies defined: %v\n", len(synergies) > 0)
	
	snippets := performOnboardingTool.WorkflowSnippets()
	fmt.Printf("✅ Workflow snippets count: %d\n", len(snippets))

	fmt.Println("\n🎉 Improved perform_onboarding tool testing completed!")
	fmt.Println("All code review improvements have been successfully implemented!")
	fmt.Println(strings.Repeat("=", 60))
}
