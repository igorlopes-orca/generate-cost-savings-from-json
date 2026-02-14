package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestAWSKMSDisabled_Calculate(t *testing.T) {
	tests := []struct {
		name  string
		asset *api.AssetDetails
		want  float64
	}{
		{
			name:  "returns fixed cost",
			asset: &api.AssetDetails{AssetUniqueID: "kms-1"},
			want:  1.0,
		},
		{
			name:  "empty asset still returns fixed cost",
			asset: &api.AssetDetails{},
			want:  1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AWSKMSDisabled{}
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
