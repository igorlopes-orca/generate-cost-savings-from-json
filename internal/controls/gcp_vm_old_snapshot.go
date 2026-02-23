package controls

import (
	"time"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&GCPVMOldSnapshot{})
}

// GCPVMOldSnapshot calculates savings from removing GCP disk snapshots older than 90 days.
type GCPVMOldSnapshot struct {
	// Now allows injecting a time for testing. If nil, time.Now() is used.
	Now func() time.Time
}

func (c *GCPVMOldSnapshot) ControlName() string {
	return "GCP VM's disks with snapshots created more than 90 days ago"
}

func (c *GCPVMOldSnapshot) NeedsSnapshotEnrichment() bool { return true }

func (c *GCPVMOldSnapshot) Calculate(asset *api.AssetDetails) (float64, error) {
	now := time.Now()
	if c.Now != nil {
		now = c.Now()
	}
	cutoff := now.AddDate(0, 0, -90)

	var total float64
	for _, d := range asset.Disks {
		for _, s := range d.Snapshots {
			if s.CreatedAt.Before(cutoff) {
				total += s.SizeGB * pricing.GCPSnapshotPricePerGB
			}
		}
	}
	return total, nil
}
