package controls

import (
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSKMSDisabled{})
}

// AWSKMSDisabled returns the fixed monthly cost of a disabled AWS KMS key.
type AWSKMSDisabled struct{}

func (c *AWSKMSDisabled) ControlName() string {
	return "Identify and remove any disabled AWS Customer Master Keys (CMK)"
}

func (c *AWSKMSDisabled) Calculate(_ *api.AssetDetails) (float64, error) {
	return pricing.KMSKeyMonthly, nil
}
