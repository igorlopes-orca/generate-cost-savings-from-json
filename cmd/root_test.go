package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/controls"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/models"
)

type fakeFetcher struct {
	responses         map[string]fetchResponse
	snapshotResponses map[string]snapshotResponse
	calls             []fetchCall
	snapshotCalls     []snapshotCall
}

type fetchResponse struct {
	details *api.AssetDetails
	err     error
}

type snapshotResponse struct {
	snapshots []api.SnapshotInfo
	err       error
}

type fetchCall struct {
	assetType     string
	assetUniqueID string
}

type snapshotCall struct {
	diskAssetType string
	diskUniqueID  string
}

func (f *fakeFetcher) FetchAsset(_ context.Context, assetType, assetUniqueID string) (*api.AssetDetails, error) {
	f.calls = append(f.calls, fetchCall{assetType: assetType, assetUniqueID: assetUniqueID})
	resp, ok := f.responses[assetUniqueID]
	if !ok {
		return nil, fmt.Errorf("unexpected asset %s", assetUniqueID)
	}
	return resp.details, resp.err
}

func (f *fakeFetcher) FetchDiskSnapshots(_ context.Context, diskAssetType, diskUniqueID string) ([]api.SnapshotInfo, error) {
	f.snapshotCalls = append(f.snapshotCalls, snapshotCall{diskAssetType: diskAssetType, diskUniqueID: diskUniqueID})
	if f.snapshotResponses == nil {
		return nil, nil
	}
	resp, ok := f.snapshotResponses[diskUniqueID]
	if !ok {
		return nil, fmt.Errorf("no snapshots configured for disk %s", diskUniqueID)
	}
	return resp.snapshots, resp.err
}

type fakeCalculator struct {
	controlName string
	savings     map[string]float64
	errByID     map[string]error
}

func (c *fakeCalculator) ControlName() string {
	return c.controlName
}

func (c *fakeCalculator) Calculate(asset *api.AssetDetails) (float64, error) {
	if err, ok := c.errByID[asset.AssetUniqueID]; ok {
		return 0, err
	}
	if value, ok := c.savings[asset.AssetUniqueID]; ok {
		return value, nil
	}
	return 0, fmt.Errorf("no savings configured for %s", asset.AssetUniqueID)
}

// fakeSnapshotCalculator implements SavingsCalculator + SnapshotEnricher.
// It calculates savings based on the number of snapshots found on each disk.
type fakeSnapshotCalculator struct {
	controlName string
	pricePerSnap float64
}

func (c *fakeSnapshotCalculator) ControlName() string { return c.controlName }

func (c *fakeSnapshotCalculator) NeedsSnapshotEnrichment() bool { return true }

func (c *fakeSnapshotCalculator) Calculate(asset *api.AssetDetails) (float64, error) {
	var total float64
	for _, d := range asset.Disks {
		total += float64(len(d.Snapshots)) * c.pricePerSnap
	}
	return total, nil
}

func TestNewRootCmd(t *testing.T) {
	filePath = "should-reset"
	apiToken = "should-reset"
	apiURL = "should-reset"
	logLevel = "should-reset"

	cmd := NewRootCmd()
	if cmd.Use != "cloud-savings" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "cloud-savings")
	}
	if cmd.Short == "" {
		t.Fatalf("Short is empty")
	}
	if cmd.RunE == nil {
		t.Fatalf("RunE is nil")
	}

	fileFlag := cmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Fatalf("file flag not found")
	}
	if fileFlag.DefValue != "" {
		t.Fatalf("file flag default = %q, want empty", fileFlag.DefValue)
	}
	if !isRequiredFlag(fileFlag) {
		t.Fatalf("file flag not marked required")
	}

	tokenFlag := cmd.Flags().Lookup("api-token")
	if tokenFlag == nil {
		t.Fatalf("api-token flag not found")
	}
	if tokenFlag.DefValue != "" {
		t.Fatalf("api-token flag default = %q, want empty", tokenFlag.DefValue)
	}
	if !isRequiredFlag(tokenFlag) {
		t.Fatalf("api-token flag not marked required")
	}

	urlFlag := cmd.Flags().Lookup("api-url")
	if urlFlag == nil {
		t.Fatalf("api-url flag not found")
	}
	if urlFlag.DefValue != "https://app.us.orcasecurity.io" {
		t.Fatalf("api-url flag default = %q, want %q", urlFlag.DefValue, "https://app.us.orcasecurity.io")
	}

	levelFlag := cmd.Flags().Lookup("log-level")
	if levelFlag == nil {
		t.Fatalf("log-level flag not found")
	}
	if levelFlag.DefValue != "" {
		t.Fatalf("log-level flag default = %q, want empty", levelFlag.DefValue)
	}
}

func TestCloudProvider(t *testing.T) {
	tests := []struct {
		name      string
		providers []string
		want      string
	}{
		{
			name:      "empty providers",
			providers: nil,
			want:      "unknown",
		},
		{
			name:      "recognized provider wins",
			providers: []string{"shiftleft", "AWS", "gcp"},
			want:      "aws",
		},
		{
			name:      "first recognized provider",
			providers: []string{"azure", "aws"},
			want:      "azure",
		},
		{
			name:      "no recognized provider returns first lower",
			providers: []string{"ShiftLeft"},
			want:      "shiftleft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := models.ReportEntry{
				Control: models.Control{ApplicableCloudProviders: tt.providers},
			}
			if got := cloudProvider(entry); got != tt.want {
				t.Fatalf("cloudProvider = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadReport(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) string
		wantErr        bool
		wantErrContain string
		wantEntries    int
	}{
		{
			name: "missing file",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.json")
			},
			wantErr:        true,
			wantErrContain: "open file",
		},
		{
			name: "invalid json",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bad.json")
				if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
					t.Fatalf("write temp file: %v", err)
				}
				return path
			},
			wantErr:        true,
			wantErrContain: "parse json",
		},
		{
			name: "valid report",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "report.json")
				data := `[
				  {
				    "AssetStatus": "Non Compliant",
				    "Control": {"Name": "Test Control", "ApplicableCloudProviders": ["AWS"]},
				    "CloudAccount": {"Name": "acct"},
				    "Asset": {"AssetUniqueId": "asset-1", "Type": "AwsEc2EbsVolume", "Name": "vol-1"}
				  }
				]`
				if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
					t.Fatalf("write temp file: %v", err)
				}
				return path
			},
			wantEntries: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			entries, err := readReport(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrContain != "" && !contains(err.Error(), tt.wantErrContain) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErrContain)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(entries) != tt.wantEntries {
				t.Fatalf("entries = %d, want %d", len(entries), tt.wantEntries)
			}
		})
	}
}

func TestProcessFindingsByControl(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string)
		wantCount    int
		wantSavings  float64
		wantAssets   int
		wantControl  string
		wantProvider string
	}{
		{
			name: "skips unsupported control",
			setup: func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string) {
				findings := []models.ReportEntry{
					{
						Control: models.Control{Name: "unregistered-control", ApplicableCloudProviders: []string{"AWS"}},
						Asset:   models.Asset{AssetUniqueId: "asset-1", Type: "AwsEc2EbsVolume"},
					},
				}
				return &fakeFetcher{responses: map[string]fetchResponse{}}, findings, nil
			},
			wantCount: 0,
		},
		{
			name: "aggregates successful calculations",
			setup: func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string) {
				controlName := "test-control-" + t.Name()
				calc := &fakeCalculator{
					controlName: controlName,
					savings: map[string]float64{
						"asset-1": 10,
						"asset-2": 5.5,
					},
					errByID: map[string]error{
						"asset-3": errors.New("boom"),
					},
				}
				controls.Register(calc)

				fetcher := &fakeFetcher{responses: map[string]fetchResponse{
					"asset-1": {details: &api.AssetDetails{AssetUniqueID: "asset-1", Type: "AwsEc2EbsVolume"}},
					"asset-2": {details: &api.AssetDetails{AssetUniqueID: "asset-2", Type: "AwsEc2EbsVolume"}},
					"asset-3": {details: &api.AssetDetails{AssetUniqueID: "asset-3", Type: "AwsEc2EbsVolume"}},
					"asset-4": {err: api.ErrUnsupportedAssetType},
					"asset-5": {err: errors.New("fetch failed")},
				}}

				findings := []models.ReportEntry{
					{Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}}, Asset: models.Asset{AssetUniqueId: "asset-1", Type: "AwsEc2EbsVolume"}},
					{Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}}, Asset: models.Asset{AssetUniqueId: "asset-2", Type: "AwsEc2EbsVolume"}},
					{Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}}, Asset: models.Asset{AssetUniqueId: "asset-3", Type: "AwsEc2EbsVolume"}},
					{Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}}, Asset: models.Asset{AssetUniqueId: "asset-4", Type: "AwsEc2EbsVolume"}},
					{Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}}, Asset: models.Asset{AssetUniqueId: "asset-5", Type: "AwsEc2EbsVolume"}},
				}
				return fetcher, findings, []string{controlName}
			},
			wantCount:    1,
			wantSavings:  15.5,
			wantAssets:   2,
			wantProvider: "aws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, findings, controlNames := tt.setup(t)
			results := processFindingsByControl(context.Background(), fetcher, findings)

			if len(results) != tt.wantCount {
				t.Fatalf("results = %d, want %d", len(results), tt.wantCount)
			}
			if tt.wantCount == 0 {
				return
			}

			if tt.wantControl == "" && len(controlNames) > 0 {
				tt.wantControl = controlNames[0]
			}
			res := results[0]
			if res.ControlName != tt.wantControl {
				t.Fatalf("ControlName = %q, want %q", res.ControlName, tt.wantControl)
			}
			if res.CloudProvider != tt.wantProvider {
				t.Fatalf("CloudProvider = %q, want %q", res.CloudProvider, tt.wantProvider)
			}
			if res.AssetCount != tt.wantAssets {
				t.Fatalf("AssetCount = %d, want %d", res.AssetCount, tt.wantAssets)
			}
			if res.MonthlySavings != tt.wantSavings {
				t.Fatalf("MonthlySavings = %v, want %v", res.MonthlySavings, tt.wantSavings)
			}
		})
	}
}

func TestDiskAssetType(t *testing.T) {
	tests := []struct {
		name         string
		instanceType string
		want         string
	}{
		{
			name:         "AWS EC2 instance maps to EBS volume",
			instanceType: "AwsEc2Instance",
			want:         "AwsEc2EbsVolume",
		},
		{
			name:         "GCP VM instance maps to GCP disk",
			instanceType: "GcpVmInstance",
			want:         "GcpVmDisk",
		},
		{
			name:         "unknown type returns empty",
			instanceType: "SomeOtherType",
			want:         "",
		},
		{
			name:         "empty type returns empty",
			instanceType: "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := diskAssetType(tt.instanceType); got != tt.want {
				t.Fatalf("diskAssetType(%q) = %q, want %q", tt.instanceType, got, tt.want)
			}
		})
	}
}

func TestProcessFindingsByControl_SnapshotEnrichment(t *testing.T) {
	tests := []struct {
		name              string
		setup             func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string)
		wantCount         int
		wantSavings       float64
		wantAssets        int
		wantSnapshotCalls int
	}{
		{
			name: "enriches disks with snapshots for snapshot calculator",
			setup: func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string) {
				controlName := "test-snapshot-enrichment-" + t.Name()
				calc := &fakeSnapshotCalculator{
					controlName:  controlName,
					pricePerSnap: 5.0,
				}
				controls.Register(calc)

				fetcher := &fakeFetcher{
					responses: map[string]fetchResponse{
						"instance-1": {details: &api.AssetDetails{
							AssetUniqueID: "instance-1",
							Type:          "AwsEc2Instance",
							Disks: []api.DiskInfo{
								{ID: "disk-a", Name: "vol-a", SizeGB: 100, VolumeType: "gp2"},
								{ID: "disk-b", Name: "vol-b", SizeGB: 50, VolumeType: "gp3"},
							},
						}},
					},
					snapshotResponses: map[string]snapshotResponse{
						"disk-a": {snapshots: []api.SnapshotInfo{
							{ID: "snap-1", SizeGB: 100, CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
							{ID: "snap-2", SizeGB: 100, CreatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
						}},
						"disk-b": {snapshots: []api.SnapshotInfo{
							{ID: "snap-3", SizeGB: 50, CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
						}},
					},
				}

				findings := []models.ReportEntry{
					{
						Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}},
						Asset:   models.Asset{AssetUniqueId: "instance-1", Type: "AwsEc2Instance"},
					},
				}
				return fetcher, findings, []string{controlName}
			},
			wantCount:         1,
			wantSavings:       15.0, // (2 snaps on disk-a + 1 snap on disk-b) * $5
			wantAssets:        1,
			wantSnapshotCalls: 2, // one per disk
		},
		{
			name: "handles snapshot fetch errors gracefully",
			setup: func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string) {
				controlName := "test-snapshot-error-" + t.Name()
				calc := &fakeSnapshotCalculator{
					controlName:  controlName,
					pricePerSnap: 10.0,
				}
				controls.Register(calc)

				fetcher := &fakeFetcher{
					responses: map[string]fetchResponse{
						"instance-2": {details: &api.AssetDetails{
							AssetUniqueID: "instance-2",
							Type:          "AwsEc2Instance",
							Disks: []api.DiskInfo{
								{ID: "disk-ok", Name: "vol-ok", SizeGB: 100, VolumeType: "gp2"},
								{ID: "disk-err", Name: "vol-err", SizeGB: 50, VolumeType: "gp3"},
							},
						}},
					},
					snapshotResponses: map[string]snapshotResponse{
						"disk-ok":  {snapshots: []api.SnapshotInfo{{ID: "snap-x", SizeGB: 100}}},
						"disk-err": {err: errors.New("API timeout")},
					},
				}

				findings := []models.ReportEntry{
					{
						Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}},
						Asset:   models.Asset{AssetUniqueId: "instance-2", Type: "AwsEc2Instance"},
					},
				}
				return fetcher, findings, []string{controlName}
			},
			wantCount:         1,
			wantSavings:       10.0, // only disk-ok has 1 snap * $10; disk-err skipped
			wantAssets:        1,
			wantSnapshotCalls: 2,
		},
		{
			name: "GCP instance enrichment uses GcpVmDisk type",
			setup: func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string) {
				controlName := "test-gcp-snapshot-" + t.Name()
				calc := &fakeSnapshotCalculator{
					controlName:  controlName,
					pricePerSnap: 2.0,
				}
				controls.Register(calc)

				fetcher := &fakeFetcher{
					responses: map[string]fetchResponse{
						"gcp-instance-1": {details: &api.AssetDetails{
							AssetUniqueID: "gcp-instance-1",
							Type:          "GcpVmInstance",
							Disks: []api.DiskInfo{
								{ID: "gcp-disk-1", Name: "disk-1", SizeGB: 200},
							},
						}},
					},
					snapshotResponses: map[string]snapshotResponse{
						"gcp-disk-1": {snapshots: []api.SnapshotInfo{
							{ID: "gcp-snap-1", SizeGB: 200},
							{ID: "gcp-snap-2", SizeGB: 200},
							{ID: "gcp-snap-3", SizeGB: 200},
						}},
					},
				}

				findings := []models.ReportEntry{
					{
						Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"GCP"}},
						Asset:   models.Asset{AssetUniqueId: "gcp-instance-1", Type: "GcpVmInstance"},
					},
				}
				return fetcher, findings, []string{controlName}
			},
			wantCount:         1,
			wantSavings:       6.0, // 3 snaps * $2
			wantAssets:        1,
			wantSnapshotCalls: 1,
		},
		{
			name: "no enrichment for non-instance asset types",
			setup: func(t *testing.T) (api.AssetFetcher, []models.ReportEntry, []string) {
				controlName := "test-no-enrichment-" + t.Name()
				calc := &fakeSnapshotCalculator{
					controlName:  controlName,
					pricePerSnap: 5.0,
				}
				controls.Register(calc)

				fetcher := &fakeFetcher{
					responses: map[string]fetchResponse{
						"volume-1": {details: &api.AssetDetails{
							AssetUniqueID: "volume-1",
							Type:          "AwsEc2EbsVolume",
						}},
					},
				}

				findings := []models.ReportEntry{
					{
						Control: models.Control{Name: controlName, ApplicableCloudProviders: []string{"AWS"}},
						Asset:   models.Asset{AssetUniqueId: "volume-1", Type: "AwsEc2EbsVolume"},
					},
				}
				return fetcher, findings, []string{controlName}
			},
			wantCount:         1,
			wantSavings:       0, // no disks, no enrichment (diskAssetType returns "")
			wantAssets:        1,
			wantSnapshotCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher, findings, controlNames := tt.setup(t)
			results := processFindingsByControl(context.Background(), fetcher, findings)

			if len(results) != tt.wantCount {
				t.Fatalf("results = %d, want %d", len(results), tt.wantCount)
			}
			if tt.wantCount == 0 {
				return
			}

			res := results[0]
			if tt.wantAssets != 0 && res.AssetCount != tt.wantAssets {
				t.Fatalf("AssetCount = %d, want %d", res.AssetCount, tt.wantAssets)
			}
			if res.MonthlySavings != tt.wantSavings {
				t.Fatalf("MonthlySavings = %v, want %v", res.MonthlySavings, tt.wantSavings)
			}
			if len(controlNames) > 0 && res.ControlName != controlNames[0] {
				t.Fatalf("ControlName = %q, want %q", res.ControlName, controlNames[0])
			}

			ff, ok := fetcher.(*fakeFetcher)
			if !ok {
				t.Fatal("fetcher is not *fakeFetcher")
			}
			if len(ff.snapshotCalls) != tt.wantSnapshotCalls {
				t.Fatalf("snapshotCalls = %d, want %d", len(ff.snapshotCalls), tt.wantSnapshotCalls)
			}
		})
	}
}

func TestConfigureLogging(t *testing.T) {
	prevLogLevel := logLevel
	prevLogger := log.Logger
	prevGlobal := zerolog.GlobalLevel()

	t.Cleanup(func() {
		logLevel = prevLogLevel
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevGlobal)
	})

	tests := []struct {
		name       string
		flagLevel  string
		envLevel   string
		wantGlobal zerolog.Level
	}{
		{
			name:       "flag level overrides env",
			flagLevel:  "debug",
			envLevel:   "error",
			wantGlobal: zerolog.DebugLevel,
		},
		{
			name:       "env level used when flag empty",
			flagLevel:  "",
			envLevel:   "warn",
			wantGlobal: zerolog.WarnLevel,
		},
		{
			name:       "default info",
			flagLevel:  "",
			envLevel:   "",
			wantGlobal: zerolog.InfoLevel,
		},
		{
			name:       "error level",
			flagLevel:  "error",
			envLevel:   "",
			wantGlobal: zerolog.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logLevel = tt.flagLevel
			if tt.envLevel == "" {
				t.Setenv("LOG_LEVEL", "")
			} else {
				t.Setenv("LOG_LEVEL", tt.envLevel)
			}

			configureLogging()

			if got := zerolog.GlobalLevel(); got != tt.wantGlobal {
				t.Fatalf("GlobalLevel = %v, want %v", got, tt.wantGlobal)
			}
		})
	}
}

func isRequiredFlag(flag *pflag.Flag) bool {
	annotations := flag.Annotations
	if annotations == nil {
		return false
	}
	values, ok := annotations[cobra.BashCompOneRequiredFlag]
	if !ok || len(values) == 0 {
		return false
	}
	return values[0] == "true"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
