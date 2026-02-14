package controls

import (
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&AWSEBSSnapshotOrphaned{})
}

// AWSEBSSnapshotOrphaned calculates savings for EBS snapshots whose originating volume no longer exists.
type AWSEBSSnapshotOrphaned struct{}

func (c *AWSEBSSnapshotOrphaned) ControlName() string {
	return "EBS snapshot's originating volume no longer exists"
}

func (c *AWSEBSSnapshotOrphaned) Calculate(asset *api.AssetDetails) (float64, error) {
	return asset.SizeGB * pricing.EBSSnapshotPricePerGB, nil
}
