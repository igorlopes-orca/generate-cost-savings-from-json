package controls

import (
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSEBSMultiSnapshot{})
}

// AWSEBSMultiSnapshot calculates savings from keeping only one snapshot per EBS volume.
// For each disk, extra snapshots beyond the first are considered waste.
type AWSEBSMultiSnapshot struct{}

func (c *AWSEBSMultiSnapshot) ControlName() string {
	return "Ensure EC2 instances with EBS volumes have only one updated snapshot"
}

func (c *AWSEBSMultiSnapshot) Calculate(asset *api.AssetDetails) (float64, error) {
	var total float64
	for _, d := range asset.Disks {
		count := len(d.Snapshots)
		if count <= 1 {
			continue
		}
		extra := count - 1
		var sumSize float64
		for _, s := range d.Snapshots {
			sumSize += s.SizeGB
		}
		avgSize := sumSize / float64(count)
		total += float64(extra) * avgSize * pricing.EBSSnapshotPricePerGB
	}
	return total, nil
}
