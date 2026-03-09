package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSEBSLowUtil_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *api.AssetDetails
		want    float64
		wantErr bool
	}{
		{
			name: "50% utilization gp2",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-1",
				SizeGB:        100,
				UsedGB:        50,
				VolumeType:    "gp2",
			},
			want: 5.0, // (100 - 50) * 0.10
		},
		{
			name: "nearly full volume",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-2",
				SizeGB:        100,
				UsedGB:        99,
				VolumeType:    "gp3",
			},
			want: 0.08, // (100 - 99) * 0.08
		},
		{
			name: "used exceeds size clamps to zero",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-3",
				SizeGB:        100,
				UsedGB:        110,
				VolumeType:    "gp2",
			},
			want: 0,
		},
		{
			name: "unknown volume type returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-4",
				SizeGB:        100,
				UsedGB:        50,
				VolumeType:    "bad",
			},
			wantErr: true,
		},
		{
			name: "used disk size unavailable returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-5",
				SizeGB:        100,
				UsedGB:        0,
				VolumeType:    "gp2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSEBSLowUtil{}
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
