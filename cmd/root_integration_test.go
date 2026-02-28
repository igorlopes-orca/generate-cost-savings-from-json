package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/models"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/output"
)

type fixtureFetcher struct {
	assets map[string]*api.AssetDetails
	calls  []string
}

func (f *fixtureFetcher) FetchAsset(_ context.Context, _ string, assetUniqueID string) (*api.AssetDetails, error) {
	f.calls = append(f.calls, assetUniqueID)
	asset, ok := f.assets[assetUniqueID]
	if !ok {
		return nil, fmt.Errorf("missing asset %s", assetUniqueID)
	}
	return asset, nil
}

func (f *fixtureFetcher) FetchDiskSnapshots(_ context.Context, diskAssetType, diskUniqueID string) ([]api.SnapshotInfo, error) {
	return nil, fmt.Errorf("unexpected snapshot fetch for %s %s", diskAssetType, diskUniqueID)
}

func TestIntegrationFixtures(t *testing.T) {
	tests := []struct {
		name          string
		assets        map[string]*api.AssetDetails
		expectedCalls []string
	}{
		{
			name: "aws-basic",
			assets: map[string]*api.AssetDetails{
				"aws-ebs-1": {
					AssetUniqueID: "aws-ebs-1",
					Type:          "AwsEc2EbsVolume",
					VolumeType:    "gp2",
					SizeGB:        100,
				},
				"aws-ebs-2": {
					AssetUniqueID: "aws-ebs-2",
					Type:          "AwsEc2EbsVolume",
					VolumeType:    "gp3",
					SizeGB:        50,
					UsedGB:        10,
				},
			},
			expectedCalls: []string{"aws-ebs-1", "aws-ebs-2"},
		},
		{
			name: "gcp-basic",
			assets: map[string]*api.AssetDetails{
				"gcp-disk-1": {
					AssetUniqueID: "gcp-disk-1",
					Type:          "GcpVmDisk",
					VolumeType:    "pd-ssd",
					SizeGB:        100,
					UsedGB:        20,
				},
				"gcp-disk-2": {
					AssetUniqueID: "gcp-disk-2",
					Type:          "GcpVmDisk",
					VolumeType:    "pd-standard",
					SizeGB:        200,
				},
			},
			expectedCalls: []string{"gcp-disk-1", "gcp-disk-2"},
		},
		{
			name: "mixed-aws-gcp",
			assets: map[string]*api.AssetDetails{
				"gcp-disk-3": {
					AssetUniqueID: "gcp-disk-3",
					Type:          "GcpVmDisk",
					VolumeType:    "pd-balanced",
					SizeGB:        60,
					UsedGB:        10,
				},
				"aws-ebs-3": {
					AssetUniqueID: "aws-ebs-3",
					Type:          "AwsEc2EbsVolume",
					VolumeType:    "gp2",
					SizeGB:        40,
				},
				"aws-kms-1": {
					AssetUniqueID: "aws-kms-1",
					Type:          "AwsKmsKey",
				},
				"aws-kms-2": {
					AssetUniqueID: "aws-kms-2",
					Type:          "AwsKmsKey",
				},
			},
			expectedCalls: []string{"aws-ebs-3", "aws-kms-1", "aws-kms-2", "gcp-disk-3"},
		},
		{
			name: "unsupported-control",
			assets: map[string]*api.AssetDetails{
				"aws-kms-3": {
					AssetUniqueID: "aws-kms-3",
					Type:          "AwsKmsKey",
				},
			},
			expectedCalls: []string{"aws-kms-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportPath := filepath.Join("/Users/igorlopes/Documents/orca_internal/generate-cost-savings-from-json/cmd/"+"testdata", "integration", tt.name, "report.json")
			expectedPath := filepath.Join("/Users/igorlopes/Documents/orca_internal/generate-cost-savings-from-json/cmd/"+"testdata", "integration", tt.name, "expected.txt")

			entries, err := readReport(reportPath)
			if err != nil {
				t.Fatalf("read report: %v", err)
			}

			var findings []models.ReportEntry
			for _, entry := range entries {
				if entry.AssetStatus == "Non Compliant" {
					findings = append(findings, entry)
				}
			}

			fetcher := &fixtureFetcher{assets: tt.assets}
			results := processFindingsByControl(context.Background(), fetcher, findings)

			var buf bytes.Buffer
			output.NewFormatter(&buf).Render(results)

			got := normalizeOutput(buf.String())
			wantBytes, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("read expected output: %v", err)
			}
			want := normalizeOutput(string(wantBytes))

			if got != want {
				t.Fatalf("output mismatch\n--- got\n%s\n--- want\n%s", got, want)
			}

			sort.Strings(fetcher.calls)
			expectedCalls := append([]string(nil), tt.expectedCalls...)
			sort.Strings(expectedCalls)
			if strings.Join(fetcher.calls, ",") != strings.Join(expectedCalls, ",") {
				t.Fatalf("fetch calls = %v, want %v", fetcher.calls, expectedCalls)
			}
		})
	}
}

func normalizeOutput(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
