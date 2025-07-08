package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"

	"github.com/9elements/rebaiser/internal/config"
)

var CLI struct {
	Config    string `short:"c" help:"Path to configuration file" default:"config.yaml"`
	LogLevel  string `short:"l" help:"Log level (debug, info, warn, error)" default:"info"`
	DryRun    bool   `short:"d" help:"Dry run mode - don't make actual changes"`
	RunOnce   bool   `short:"o" help:"Run once and exit (don't run periodically)"`
	Version   bool   `short:"v" help:"Show version information"`
}

func main() {
	ctx := kong.Parse(&CLI)

	if CLI.Version {
		logrus.Info("AI Rebaser v1.0.0")
		return
	}

	// Setup structured logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	
	level, err := logrus.ParseLevel(CLI.LogLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Invalid log level")
	}
	logrus.SetLevel(level)

	log := logrus.WithField("component", "main")
	log.Info("Starting AI Rebaser")

	// Load configuration
	cfg, err := config.LoadConfig(CLI.Config)
	if err != nil {
		log.WithError(err).Fatal("Failed to load configuration")
	}

	// Apply CLI overrides
	if CLI.DryRun {
		cfg.DryRun = true
	}

	// Create context for graceful shutdown
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Received shutdown signal")
		cancel()
	}()

	// Start the rebaser service
	if err := runRebaser(appCtx, cfg); err != nil {
		log.WithError(err).Fatal("Rebaser failed")
	}

	log.Info("AI Rebaser stopped")
}

func runRebaser(ctx context.Context, cfg *config.Config) error {
	log := logrus.WithField("component", "rebaser")
	
	// Run once if requested
	if CLI.RunOnce {
		log.Info("Running single rebase operation")
		return performRebase(ctx, cfg)
	}

	// Create ticker for periodic rebasing
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.WithField("interval", cfg.Interval).Info("Starting rebaser with configured interval")

	// Run initial rebase
	if err := performRebase(ctx, cfg); err != nil {
		log.WithError(err).Error("Initial rebase failed")
	}

	// Run periodic rebases
	for {
		select {
		case <-ctx.Done():
			log.Info("Shutting down rebaser")
			return nil
		case <-ticker.C:
			if err := performRebase(ctx, cfg); err != nil {
				log.WithError(err).Error("Periodic rebase failed")
			}
		}
	}
}

func performRebase(ctx context.Context, cfg *config.Config) error {
	log := logrus.WithField("component", "rebase")
	log.Info("Starting rebase operation")

	// TODO: Implement rebase logic
	// 1. Git operations (fetch, rebase, conflict detection)
	// 2. AI conflict resolution
	// 3. Run tests
	// 4. Create PR
	// 5. Send notifications

	log.Info("Rebase operation completed")
	return nil
}