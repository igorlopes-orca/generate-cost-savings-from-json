package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestGCPDiskLowUtil_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *api.AssetDetails
		want    float64
		wantErr bool
	}{
		{
			name: "pd-standard with 80% unused",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-1",
				SizeGB:        100,
				UsedGB:        20,
				VolumeType:    "pd-standard",
			},
			want: 3.2, // (100-20) * 0.04
		},
		{
			name: "pd-ssd fully used returns zero",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-2",
				SizeGB:        50,
				UsedGB:        50,
				VolumeType:    "pd-ssd",
			},
			want: 0,
		},
		{
			name: "used exceeds size clamps to zero",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-3",
				SizeGB:        100,
				UsedGB:        120,
				VolumeType:    "pd-balanced",
			},
			want: 0,
		},
		{
			name: "unknown volume type returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-4",
				SizeGB:        100,
				UsedGB:        10,
				VolumeType:    "local-ssd",
			},
			wantErr: true,
		},
		{
			name: "used disk size unavailable returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-5",
				SizeGB:        100,
				UsedGB:        0,
				VolumeType:    "pd-standard",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GCPDiskLowUtil{}
			got, err := c.Calculate(tt.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
