package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/controls"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/models"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/output"
)

var (
	filePath string
	apiToken string
	apiURL   string
	logLevel string
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud-savings",
		Short: "Calculate potential monthly USD savings from an Orca cost optimization report",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to Orca cost optimization JSON report (required)")
	cmd.Flags().StringVarP(&apiToken, "api-token", "t", "", "Orca API token (required)")
	cmd.Flags().StringVar(&apiURL, "api-url", "https://app.us.orcasecurity.io", "Orca API base URL")
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Log level: debug, info, warn, error (overrides LOG_LEVEL env var)")

	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("api-token")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	configureLogging()

	// Read and parse report
	entries, err := readReport(filePath)
	if err != nil {
		return fmt.Errorf("read report: %w", err)
	}
	log.Info().Int("total_entries", len(entries)).Msg("parsed report")

	// Filter non-compliant entries
	var findings []models.ReportEntry
	for _, e := range entries {
		if e.AssetStatus == "Non Compliant" {
			findings = append(findings, e)
		}
	}
	log.Info().Int("findings", len(findings)).Msg("filtered non-compliant entries")

	// Log per-control summary and calculator availability
	logControlSummary(findings)

	// Create the Orca API client
	fetcher := api.NewOrcaClient(apiURL, apiToken)

	// Group by control and calculate savings
	results := processFindingsByControl(cmd.Context(), fetcher, findings)

	// Render output
	f := output.NewFormatter(os.Stdout)
	f.Render(results)

	return nil
}

// logControlSummary logs the count of non-compliant assets per control (INFO)
// and which controls have calculators registered vs not (DEBUG).
func logControlSummary(findings []models.ReportEntry) {
	type controlInfo struct {
		count         int
		cloudProvider string
	}

	byControl := make(map[string]*controlInfo)
	for _, f := range findings {
		name := f.Control.Name
		if _, ok := byControl[name]; !ok {
			byControl[name] = &controlInfo{cloudProvider: cloudProvider(f)}
		}
		byControl[name].count++
	}

	for name, info := range byControl {
		log.Info().
			Str("control", name).
			Str("cloud_provider", info.cloudProvider).
			Int("assets", info.count).
			Msg("non-compliant assets")
	}

	registered := controls.All()
	for name, info := range byControl {
		if _, ok := registered[name]; ok {
			log.Debug().
				Str("control", name).
				Str("cloud_provider", info.cloudProvider).
				Msg("calculator available")
		} else {
			log.Debug().
				Str("control", name).
				Str("cloud_provider", info.cloudProvider).
				Msg("no calculator available")
		}
	}
}

func processFindingsByControl(ctx context.Context, fetcher api.AssetFetcher, findings []models.ReportEntry) []output.ControlResult {
	type controlGroup struct {
		name          string
		cloudProvider string
		assets        []models.ReportEntry
	}

	grouped := make(map[string]*controlGroup)
	for _, f := range findings {
		name := f.Control.Name
		if _, ok := grouped[name]; !ok {
			grouped[name] = &controlGroup{name: name, cloudProvider: cloudProvider(f)}
		}
		grouped[name].assets = append(grouped[name].assets, f)
	}

	var results []output.ControlResult
	for _, g := range grouped {
		calc := controls.Get(g.name)
		if calc == nil {
			log.Debug().
				Str("control", g.name).
				Str("cloud_provider", g.cloudProvider).
				Msg("no calculator registered, skipping")
			continue
		}

		var totalSavings float64
		var successCount int
		for _, asset := range g.assets {
			// Sub-logger with common context for every log in this iteration.
			assetLog := log.With().
				Str("control", g.name).
				Str("cloud_provider", cloudProvider(asset)).
				Str("asset_type", asset.Asset.Type).
				Str("asset_unique_id", asset.Asset.AssetUniqueId).
				Logger()

			assetLog.Debug().
				Str("asset_name", asset.Asset.Name).
				Msg("fetching asset details")

			details, err := fetcher.FetchAsset(ctx, asset.Asset.Type, asset.Asset.AssetUniqueId)
			if err != nil {
				if errors.Is(err, api.ErrUnsupportedAssetType) {
					assetLog.Debug().Msg("asset type not supported, skipping")
				} else {
					assetLog.Error().Err(err).Msg("failed to fetch asset details")
				}
				continue
			}

			savings, err := calc.Calculate(details)
			if err != nil {
				assetLog.Error().Err(err).Msg("failed to calculate savings")
				continue
			}

			assetLog.Debug().Float64("savings", savings).Msg("calculated savings")

			totalSavings += savings
			successCount++
		}

		results = append(results, output.ControlResult{
			ControlName:    g.name,
			AssetCount:     successCount,
			MonthlySavings: totalSavings,
		})
	}

	return results
}

// cloudProvider extracts the primary cloud provider from a report entry.
func cloudProvider(entry models.ReportEntry) string {
	providers := entry.Control.ApplicableCloudProviders
	if len(providers) == 0 {
		return "unknown"
	}
	// Filter out non-provider values like "shiftleft"
	var real []string
	for _, p := range providers {
		switch strings.ToLower(p) {
		case "aws", "azure", "gcp", "alicloud":
			real = append(real, strings.ToLower(p))
		}
	}
	if len(real) == 0 {
		return strings.ToLower(providers[0])
	}
	return real[0]
}

func readReport(path string) ([]models.ReportEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", path, err)
	}

	var entries []models.ReportEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	return entries, nil
}

func configureLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Resolve log level: flag > env var > default (info)
	level := logLevel
	if level == "" {
		level = os.Getenv("LOG_LEVEL")
	}
	if level == "" {
		level = "info"
	}

	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
