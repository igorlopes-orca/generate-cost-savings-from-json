package controls

import (
	"fmt"

	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/pricing"
)

func init() {
	Register(&GCPDiskLowUtil{})
}

// GCPDiskLowUtil calculates savings from the unused portion of low-utilization GCP disks.
type GCPDiskLowUtil struct{}

func (c *GCPDiskLowUtil) ControlName() string {
	return "Low disk space utilization in your GCP Disk"
}

func (c *GCPDiskLowUtil) Calculate(asset *api.AssetDetails) (float64, error) {
	if asset.UsedGB == 0 {
		return 0, fmt.Errorf("low-util GCP disk %s: used disk size unavailable", asset.AssetUniqueID)
	}
	price, ok := pricing.GCPDiskPricePerGB[asset.VolumeType]
	if !ok {
		return 0, fmt.Errorf("low-util GCP disk %s: unknown volume type %q", asset.AssetUniqueID, asset.VolumeType)
	}
	unused := asset.SizeGB - asset.UsedGB
	if unused < 0 {
		unused = 0
	}
	return unused * price, nil
}
