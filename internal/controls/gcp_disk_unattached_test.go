package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestGCPDiskUnattached_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *api.AssetDetails
		want    float64
		wantErr bool
	}{
		{
			name: "pd-standard disk",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-1",
				SizeGB:        100,
				VolumeType:    "pd-standard",
			},
			want: 4.0, // 100 * 0.04
		},
		{
			name: "pd-ssd disk",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-2",
				SizeGB:        200,
				VolumeType:    "pd-ssd",
			},
			want: 34.0, // 200 * 0.17
		},
		{
			name: "pd-balanced disk",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-3",
				SizeGB:        500,
				VolumeType:    "pd-balanced",
			},
			want: 50.0, // 500 * 0.10
		},
		{
			name: "hyperdisk-balanced disk",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-4",
				SizeGB:        1000,
				VolumeType:    "hyperdisk-balanced",
			},
			want: 60.0, // 1000 * 0.06
		},
		{
			name: "unknown volume type returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "disk-5",
				SizeGB:        50,
				VolumeType:    "local-ssd",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GCPDiskUnattached{}
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
