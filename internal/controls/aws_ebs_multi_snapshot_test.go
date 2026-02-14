package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSEBSMultiSnapshot_Calculate(t *testing.T) {
	tests := []struct {
		name  string
		asset *api.AssetDetails
		want  float64
	}{
		{
			name: "one disk with 3 snapshots of 10 GB each",
			asset: &api.AssetDetails{
				Disks: []api.DiskInfo{
					{
						ID: "vol-1",
						Snapshots: []api.SnapshotInfo{
							{ID: "s1", SizeGB: 10},
							{ID: "s2", SizeGB: 10},
							{ID: "s3", SizeGB: 10},
						},
					},
				},
			},
			want: 1.0, // 2 extra * 10 avg * 0.05
		},
		{
			name: "one disk with 1 snapshot",
			asset: &api.AssetDetails{
				Disks: []api.DiskInfo{
					{
						ID: "vol-1",
						Snapshots: []api.SnapshotInfo{
							{ID: "s1", SizeGB: 50},
						},
					},
				},
			},
			want: 0,
		},
		{
			name: "no disks",
			asset: &api.AssetDetails{},
			want:  0,
		},
		{
			name: "multiple disks with snapshots",
			asset: &api.AssetDetails{
				Disks: []api.DiskInfo{
					{
						ID: "vol-1",
						Snapshots: []api.SnapshotInfo{
							{ID: "s1", SizeGB: 20},
							{ID: "s2", SizeGB: 20},
						},
					},
					{
						ID: "vol-2",
						Snapshots: []api.SnapshotInfo{
							{ID: "s3", SizeGB: 40},
							{ID: "s4", SizeGB: 40},
							{ID: "s5", SizeGB: 40},
						},
					},
				},
			},
			want: 5.0, // vol-1: 1*20*0.05=1.0, vol-2: 2*40*0.05=4.0
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSEBSMultiSnapshot{}
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
