package controls

import (
	"testing"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

func TestGCPKMSDisabled_Calculate(t *testing.T) {
	tests := []struct {
		name  string
		asset *api.AssetDetails
		want  float64
	}{
		{
			name:  "returns fixed GCP KMS key version monthly cost",
			asset: &api.AssetDetails{AssetUniqueID: "GcpKmsKey_project123_key-abc"},
			want:  0.06,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GCPKMSDisabled{}
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
