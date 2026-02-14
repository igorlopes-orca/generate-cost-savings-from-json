package controls

import (
	"github.com/igorlopes-orca/generate-cost-savings-from-json/internal/api"
)

// SavingsCalculator computes the estimated monthly USD savings for a single non-compliant asset.
type SavingsCalculator interface {
	// ControlName returns the exact Control.Name string from the JSON report
	// that this calculator handles.
	ControlName() string

	// Calculate returns the estimated monthly USD savings for a single asset.
	// The asset parameter contains enriched details fetched from the Orca API.
	Calculate(asset *api.AssetDetails) (float64, error)
}
