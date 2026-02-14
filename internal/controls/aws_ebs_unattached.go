package controls

import (
	"fmt"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSEBSUnattached{})
}

// AWSEBSUnattached calculates savings for unattached EBS volumes.
type AWSEBSUnattached struct{}

func (c *AWSEBSUnattached) ControlName() string {
	return "Ensure EBS volume is attached to an EC2 instance"
}

func (c *AWSEBSUnattached) Calculate(asset *api.AssetDetails) (float64, error) {
	price, ok := pricing.EBSPricePerGB[asset.VolumeType]
	if !ok {
		return 0, fmt.Errorf("unattached EBS %s: unknown volume type %q", asset.AssetUniqueID, asset.VolumeType)
	}
	return asset.SizeGB * price, nil
}
