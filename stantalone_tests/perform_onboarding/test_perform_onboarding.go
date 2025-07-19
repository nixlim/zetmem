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

	// Test perform_onboarding tool
	fmt.Println("🧪 Testing perform_onboarding Tool")
	fmt.Println(strings.Repeat("=", 50))

	// Create perform_onboarding tool with configuration
	performOnboardingTool := memory.NewPerformOnboardingTool(workspaceService, cfg.Onboarding, logger.Named("perform_onboarding"))
	
	fmt.Printf("✅ Created perform_onboarding tool: %s\n", performOnboardingTool.Name())
	fmt.Printf("📝 Description: %s\n", performOnboardingTool.Description())

	// Test 1: Basic onboarding with current directory
	fmt.Println("\n1. Testing basic onboarding with current directory...")
	basicArgs := map[string]interface{}{
		"project_name": "Test Onboarding Project",
		"include_strategy_guide": false, // Skip strategy guide for cleaner output
	}
	
	basicResult, err := performOnboardingTool.Execute(ctx, basicArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success! IsError: %v\n", basicResult.IsError)
		fmt.Printf("📄 Content count: %d\n", len(basicResult.Content))
		if len(basicResult.Content) > 0 {
			// Show first 500 characters of response
			text := basicResult.Content[0].Text
			if len(text) > 500 {
				text = text[:500] + "..."
			}
			fmt.Printf("📄 Response preview:\n%s\n", text)
		}
	}

	// Test 2: Onboarding with specific project path
	fmt.Println("\n2. Testing onboarding with specific project path...")
	pathArgs := map[string]interface{}{
		"project_path": "/Users/test/specific-project",
		"project_name": "Specific Project Test",
		"include_strategy_guide": false,
	}
	
	pathResult, err := performOnboardingTool.Execute(ctx, pathArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success! IsError: %v\n", pathResult.IsError)
		fmt.Printf("📄 Content count: %d\n", len(pathResult.Content))
	}

	// Test 3: Onboarding with strategy guide included
	fmt.Println("\n3. Testing onboarding with strategy guide included...")
	strategyArgs := map[string]interface{}{
		"project_name": "Strategy Guide Test",
		"include_strategy_guide": true,
	}
	
	strategyResult, err := performOnboardingTool.Execute(ctx, strategyArgs)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success! IsError: %v\n", strategyResult.IsError)
		fmt.Printf("📄 Content count: %d\n", len(strategyResult.Content))
		if len(strategyResult.Content) > 0 {
			text := strategyResult.Content[0].Text
			fmt.Printf("📄 Response length: %d characters\n", len(text))
			if len(text) > 1000 {
				fmt.Printf("📄 Strategy guide included: YES (response > 1000 chars)\n")
			} else {
				fmt.Printf("📄 Strategy guide included: NO (response < 1000 chars)\n")
			}
		}
	}

	// Test 4: Validate enhanced tool interface methods
	fmt.Println("\n4. Testing enhanced tool interface methods...")
	
	triggers := performOnboardingTool.UsageTriggers()
	fmt.Printf("✅ Usage triggers count: %d\n", len(triggers))
	
	practices := performOnboardingTool.BestPractices()
	fmt.Printf("✅ Best practices count: %d\n", len(practices))
	
	synergies := performOnboardingTool.Synergies()
	fmt.Printf("✅ Synergies defined: %v\n", len(synergies) > 0)
	
	snippets := performOnboardingTool.WorkflowSnippets()
	fmt.Printf("✅ Workflow snippets count: %d\n", len(snippets))

	fmt.Println("\n🎉 perform_onboarding tool testing completed!")
	fmt.Println(strings.Repeat("=", 50))
}
