package controls

import (
	"testing"
	"time"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSEBSOldSnapshot_Calculate(t *testing.T) {
	fixedNow := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	cutoff := fixedNow.AddDate(0, 0, -90) // 2025-03-03

	tests := []struct {
		name  string
		asset *api.AssetDetails
		want  float64
	}{
		{
			name: "one old snapshot one recent",
			asset: &api.AssetDetails{
				Disks: []api.DiskInfo{
					{
						ID: "vol-1",
						Snapshots: []api.SnapshotInfo{
							{ID: "s1", SizeGB: 100, CreatedAt: cutoff.AddDate(0, 0, -1)},  // old
							{ID: "s2", SizeGB: 50, CreatedAt: cutoff.AddDate(0, 0, 1)},     // recent
						},
					},
				},
			},
			want: 5.0, // 100 * 0.05
		},
		{
			name: "all snapshots recent",
			asset: &api.AssetDetails{
				Disks: []api.DiskInfo{
					{
						ID: "vol-1",
						Snapshots: []api.SnapshotInfo{
							{ID: "s1", SizeGB: 100, CreatedAt: fixedNow.AddDate(0, 0, -10)},
						},
					},
				},
			},
			want: 0,
		},
		{
			name: "multiple old snapshots across disks",
			asset: &api.AssetDetails{
				Disks: []api.DiskInfo{
					{
						ID: "vol-1",
						Snapshots: []api.SnapshotInfo{
							{ID: "s1", SizeGB: 20, CreatedAt: cutoff.AddDate(0, 0, -30)},
						},
					},
					{
						ID: "vol-2",
						Snapshots: []api.SnapshotInfo{
							{ID: "s2", SizeGB: 40, CreatedAt: cutoff.AddDate(0, 0, -60)},
						},
					},
				},
			},
			want: 3.0, // (20 + 40) * 0.05
		},
		{
			name:  "no disks",
			asset: &api.AssetDetails{},
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSEBSOldSnapshot{Now: func() time.Time { return fixedNow }}
			got, err := c.Calculate(tt.asset)
			if err != nil {
				t.Errorf("Calculate() unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
