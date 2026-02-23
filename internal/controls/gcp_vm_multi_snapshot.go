package controls

import (
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&GCPVMMultiSnapshot{})
}

// GCPVMMultiSnapshot calculates savings from keeping only one snapshot per GCP disk.
// For each disk, extra snapshots beyond the first are considered waste.
type GCPVMMultiSnapshot struct{}

func (c *GCPVMMultiSnapshot) ControlName() string {
	return "Ensure gcp VM's disks have only one snapshot"
}

func (c *GCPVMMultiSnapshot) NeedsSnapshotEnrichment() bool { return true }

func (c *GCPVMMultiSnapshot) Calculate(asset *api.AssetDetails) (float64, error) {
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
		total += float64(extra) * avgSize * pricing.GCPSnapshotPricePerGB
	}
	return total, nil
}
