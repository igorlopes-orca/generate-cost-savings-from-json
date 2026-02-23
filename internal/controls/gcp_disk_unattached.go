package controls

import (
	"fmt"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&GCPDiskUnattached{})
}

// GCPDiskUnattached calculates savings for unattached GCP persistent disks.
type GCPDiskUnattached struct{}

func (c *GCPDiskUnattached) ControlName() string {
	return "Ensure GCP disk is attached to a virtual machine"
}

func (c *GCPDiskUnattached) Calculate(asset *api.AssetDetails) (float64, error) {
	price, ok := pricing.GCPDiskPricePerGB[asset.VolumeType]
	if !ok {
		return 0, fmt.Errorf("unattached GCP disk %s: unknown volume type %q", asset.AssetUniqueID, asset.VolumeType)
	}
	return asset.SizeGB * price, nil
}
