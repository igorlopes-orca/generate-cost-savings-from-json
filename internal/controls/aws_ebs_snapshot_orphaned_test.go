package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSEBSSnapshotOrphaned_Calculate(t *testing.T) {
	tests := []struct {
		name  string
		asset *api.AssetDetails
		want  float64
	}{
		{
			name: "10 GB snapshot",
			asset: &api.AssetDetails{
				AssetUniqueID: "snap-1",
				SizeGB:        10,
			},
			want: 0.50, // 10 * 0.05
		},
		{
			name: "100 GB snapshot",
			asset: &api.AssetDetails{
				AssetUniqueID: "snap-2",
				SizeGB:        100,
			},
			want: 5.0, // 100 * 0.05
		},
		{
			name: "zero size snapshot",
			asset: &api.AssetDetails{
				AssetUniqueID: "snap-3",
				SizeGB:        0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSEBSSnapshotOrphaned{}
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
