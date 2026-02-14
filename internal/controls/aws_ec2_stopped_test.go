package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSEC2Stopped_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *api.AssetDetails
		want    float64
		wantErr bool
	}{
		{
			name: "single gp2 disk",
			asset: &api.AssetDetails{
				AssetUniqueID: "i-123",
				Disks: []api.DiskInfo{
					{ID: "vol-1", SizeGB: 100, VolumeType: "gp2"},
				},
			},
			want: 10.0, // 100 * 0.10
		},
		{
			name: "multiple disks different types",
			asset: &api.AssetDetails{
				AssetUniqueID: "i-456",
				Disks: []api.DiskInfo{
					{ID: "vol-1", SizeGB: 100, VolumeType: "gp3"},
					{ID: "vol-2", SizeGB: 500, VolumeType: "st1"},
				},
			},
			want: 30.5, // (100 * 0.08) + (500 * 0.045)
		},
		{
			name: "no disks returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "i-789",
			},
			wantErr: true,
		},
		{
			name: "unknown volume type returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "i-000",
				Disks: []api.DiskInfo{
					{ID: "vol-1", SizeGB: 50, VolumeType: "unknown"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSEC2Stopped{}
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
