package controls

import (
	"time"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSEBSOldSnapshot{})
}

// AWSEBSOldSnapshot calculates savings from removing snapshots older than 90 days.
type AWSEBSOldSnapshot struct {
	// Now allows injecting a time for testing. If nil, time.Now() is used.
	Now func() time.Time
}

func (c *AWSEBSOldSnapshot) ControlName() string {
	return "Ensure EC2 instances with EBS volumes with snapshots created less than 90 days ago"
}

func (c *AWSEBSOldSnapshot) Calculate(asset *api.AssetDetails) (float64, error) {
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	cutoff := now.AddDate(0, 0, -90)

	var total float64
	for _, d := range asset.Disks {
		for _, s := range d.Snapshots {
			if s.CreatedAt.Before(cutoff) {
				total += s.SizeGB * pricing.EBSSnapshotPricePerGB
			}
		}
	}
	return total, nil
}
