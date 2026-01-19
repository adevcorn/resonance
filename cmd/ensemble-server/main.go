package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adevcorn/ensemble/internal/config"
	"github.com/adevcorn/ensemble/internal/protocol"
	"github.com/adevcorn/ensemble/internal/server/agent"
	"github.com/adevcorn/ensemble/internal/server/api"
	"github.com/adevcorn/ensemble/internal/server/orchestration"
	"github.com/adevcorn/ensemble/internal/server/provider"
	"github.com/adevcorn/ensemble/internal/server/provider/anthropic"
	"github.com/adevcorn/ensemble/internal/server/provider/gemini"
	"github.com/adevcorn/ensemble/internal/server/provider/openai"
	"github.com/adevcorn/ensemble/internal/server/provider/zai"
	"github.com/adevcorn/ensemble/internal/server/storage"
	"github.com/adevcorn/ensemble/internal/server/tool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	configFile string

	rootCmd = &cobra.Command{
		Use:   "ensemble-server",
		Short: "Ensemble Server - Multi-Agent Coordination Backend",
		Long: `Ensemble Server is the backend component of the Ensemble multi-agent 
coordination tool. It manages agent orchestration, LLM provider communication, 
and session management.`,
		RunE: runServer,
	}
)

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// 1. Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Starting Ensemble Server")
	log.Info().Str("version", version).Msg("Server version")

	// 2. Load configuration
	cfg, err := config.LoadServerConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set log level from config
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		log.Warn().Str("level", cfg.Logging.Level).Msg("Invalid log level, using info")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Str("host", cfg.Server.Host).Int("port", cfg.Server.Port).Msg("Server configuration loaded")

	// 3. Initialize provider registry
	registry := provider.NewRegistry()

	// Register Anthropic provider if configured
	if cfg.Providers.Anthropic.APIKey != "" {
		anthProvider, err := anthropic.NewProvider(cfg.Providers.Anthropic.APIKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize Anthropic provider")
		} else {
			registry.Register(anthProvider)
			log.Info().Msg("Anthropic provider registered")
		}
	}

	// Register OpenAI provider if configured
	if cfg.Providers.OpenAI.APIKey != "" {
		openaiProvider, err := openai.NewProvider(cfg.Providers.OpenAI.APIKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize OpenAI provider")
		} else {
			registry.Register(openaiProvider)
			log.Info().Msg("OpenAI provider registered")
		}
	}

	// Register Zai provider if configured
	if cfg.Providers.Zai.APIKey != "" {
		zaiProvider, err := zai.NewProvider(zai.Config{
			APIKey:  cfg.Providers.Zai.APIKey,
			BaseURL: cfg.Providers.Zai.BaseURL,
		})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize Zai provider")
		} else {
			registry.Register(zaiProvider)
			log.Info().Msg("Zai provider registered")
		}
	}

	// Register Gemini provider if configured
	if cfg.Providers.Gemini.UseCLI {
		// Use Gemini CLI bridge (Node.js)
		bridgeURL := cfg.Providers.Gemini.BridgeURL
		if bridgeURL == "" {
			bridgeURL = "http://localhost:3001"
		}

		geminiProvider, err := gemini.NewCLIProvider(bridgeURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize Gemini CLI provider")
		} else {
			registry.Register(geminiProvider)
			log.Info().Str("bridge_url", bridgeURL).Msg("Gemini CLI provider registered")
		}
	} else if cfg.Providers.Gemini.APIKey != "" {
		// Use direct SDK with API key
		ctx := context.Background()
		geminiProvider, err := gemini.NewProvider(ctx, cfg.Providers.Gemini.APIKey)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize Gemini provider")
		} else {
			registry.Register(geminiProvider)
			log.Info().Msg("Gemini provider registered")
		}
	}

	if len(registry.List()) == 0 {
		return fmt.Errorf("no LLM providers configured")
	}

	// 4. Load agents
	agentPool := agent.NewPool(registry)

	agentsPath := cfg.Agents.Path
	if agentsPath == "" {
		agentsPath = "./agents"
	}

	loader := agent.NewLoader(agentsPath)
	definitions, err := loader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load agents: %w", err)
	}

	if err := agentPool.Load(definitions); err != nil {
		return fmt.Errorf("failed to load agents into pool: %w", err)
	}

	log.Info().Int("count", agentPool.Count()).Msg("Agents loaded")

	// Start hot-reload watcher if enabled
	if cfg.Agents.Watch {
		watcher, err := agent.NewWatcher(loader, agentPool, true)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to start agent watcher")
		} else {
			go watcher.Start(context.Background())
			log.Info().Msg("Agent hot-reload watcher started")
		}
	}

	// 5. Setup tool registry
	toolRegistry := tool.NewRegistry()

	// Register collaborate tool
	// Note: Actual handling done in orchestration engine's HandleCollaboration method
	collaborateTool := tool.NewCollaborateTool(agentPool, func(from string, input *protocol.CollaborateInput) error {
		// This will be handled by the orchestration engine
		return nil
	})
	toolRegistry.Register(collaborateTool)

	// Register assemble_team tool (with no-op callback for now)
	assembleTeamTool := tool.NewAssembleTeamTool(agentPool, func(agents []string, reason string) error {
		// This will be handled by the orchestration engine
		return nil
	})
	toolRegistry.Register(assembleTeamTool)

	// Register file system tools
	// Use current working directory as base for file operations
	cwd, err := os.Getwd()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get working directory, using '.'")
		cwd = "."
	}

	readFileTool := tool.NewReadFileTool(cwd)
	toolRegistry.Register(readFileTool)

	// Register write_file as a client-side tool (executed on client, schema provided by server)
	writeFileTool := tool.NewWriteFileTool()
	toolRegistry.Register(writeFileTool)

	// Register execute_command as a client-side tool
	executeCommandTool := tool.NewExecuteCommandTool()
	toolRegistry.Register(executeCommandTool)

	listDirectoryTool := tool.NewListDirectoryTool(cwd)
	toolRegistry.Register(listDirectoryTool)

	// Register web tools
	fetchURLTool := tool.NewFetchURLTool()
	toolRegistry.Register(fetchURLTool)

	webSearchTool := tool.NewWebSearchTool()
	toolRegistry.Register(webSearchTool)

	log.Info().Int("count", toolRegistry.Count()).Msg("Server tools registered")

	// 6. Initialize storage
	storagePath := cfg.Storage.Path
	if storagePath == "" {
		storagePath = "./data"
	}

	jsonStorage, err := storage.NewJSONStorage(storagePath)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	sessionManager := storage.NewSessionManager(jsonStorage)
	log.Info().Str("path", storagePath).Msg("Storage initialized")

	// 7. Create orchestration engine
	// Note: For now we create a simplified engine without streaming callbacks
	// In production, we'd need to refactor the engine to support per-session callbacks
	engine, err := orchestration.NewEngine(
		agentPool,
		toolRegistry,
		nil, // onMessage callback
		nil, // onToolCall callback
	)
	if err != nil {
		return fmt.Errorf("failed to create orchestration engine: %w", err)
	}

	log.Info().Msg("Orchestration engine created")

	// 8. Create HTTP server
	apiServer := api.NewServer(sessionManager, agentPool, engine, toolRegistry)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      apiServer.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Start server
	serverErrors := make(chan error, 1)

	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server listening")
		serverErrors <- httpServer.ListenAndServe()
	}()

	// 10. Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Error during shutdown")
			if err := httpServer.Close(); err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}
		}

		log.Info().Msg("Server stopped gracefully")
	}

	return nil
}
