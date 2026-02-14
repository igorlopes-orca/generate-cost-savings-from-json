package controls

import (
	"fmt"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSEC2Stopped{})
}

// AWSEC2Stopped calculates savings for stopped EC2 instances by summing
// the cost of their attached EBS volumes.
type AWSEC2Stopped struct{}

func (c *AWSEC2Stopped) ControlName() string {
	return "Stopped Ec2 instance not removed"
}

func (c *AWSEC2Stopped) Calculate(asset *api.AssetDetails) (float64, error) {
	if len(asset.Disks) == 0 {
		return 0, fmt.Errorf("stopped EC2 %s: no attached disks", asset.AssetUniqueID)
	}

	var total float64
	for _, d := range asset.Disks {
		price, ok := pricing.EBSPricePerGB[d.VolumeType]
		if !ok {
			return 0, fmt.Errorf("stopped EC2 %s: unknown volume type %q on disk %s", asset.AssetUniqueID, d.VolumeType, d.ID)
		}
		total += d.SizeGB * price
	}
	return total, nil
}
