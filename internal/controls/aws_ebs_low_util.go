package controls

import (
	"fmt"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSEBSLowUtil{})
}

// AWSEBSLowUtil calculates savings from the unused portion of low-utilization EBS volumes.
type AWSEBSLowUtil struct{}

func (c *AWSEBSLowUtil) ControlName() string {
	return "Low disk space utilization in your AWS EC2 EBS volume"
}

func (c *AWSEBSLowUtil) Calculate(asset *api.AssetDetails) (float64, error) {
	if asset.UsedGB == 0 {
		return 0, fmt.Errorf("low-util EBS %s: used disk size unavailable", asset.AssetUniqueID)
	}
	price, ok := pricing.EBSPricePerGB[asset.VolumeType]
	if !ok {
		return 0, fmt.Errorf("low-util EBS %s: unknown volume type %q", asset.AssetUniqueID, asset.VolumeType)
	}
	unused := asset.SizeGB - asset.UsedGB
	if unused < 0 {
		unused = 0
	}
	return unused * price, nil
}
