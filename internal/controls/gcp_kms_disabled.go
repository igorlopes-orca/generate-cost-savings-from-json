package controls

import (
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&GCPKMSDisabled{})
}

// GCPKMSDisabled returns the fixed monthly cost of a disabled GCP KMS key version.
type GCPKMSDisabled struct{}

func (c *GCPKMSDisabled) ControlName() string {
	return "Identify and remove any disabled GCP KMS primary key versions"
}

func (c *GCPKMSDisabled) Calculate(_ *api.AssetDetails) (float64, error) {
	return pricing.GCPKMSKeyVersionMonthly, nil
}
