package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSEBSUnattached_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		asset   *api.AssetDetails
		want    float64
		wantErr bool
	}{
		{
			name: "gp2 volume",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-1",
				SizeGB:        100,
				VolumeType:    "gp2",
			},
			want: 10.0, // 100 * 0.10
		},
		{
			name: "io1 volume",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-2",
				SizeGB:        200,
				VolumeType:    "io1",
			},
			want: 25.0, // 200 * 0.125
		},
		{
			name: "sc1 volume",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-3",
				SizeGB:        1000,
				VolumeType:    "sc1",
			},
			want: 15.0, // 1000 * 0.015
		},
		{
			name: "unknown volume type returns error",
			asset: &api.AssetDetails{
				AssetUniqueID: "vol-4",
				SizeGB:        50,
				VolumeType:    "magnetic",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSEBSUnattached{}
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
