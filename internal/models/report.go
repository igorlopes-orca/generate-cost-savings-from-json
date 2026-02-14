package models

// ReportEntry represents a single row from the Orca cost optimization JSON report.
type ReportEntry struct {
	AssetStatus  string       `json:"AssetStatus"`
	Control      Control      `json:"Control"`
	CloudAccount CloudAccount `json:"CloudAccount"`
	Asset        Asset        `json:"Asset"`
}

type Control struct {
	Name                    string   `json:"Name"`
	Section                 string   `json:"Section"`
	ResultType              string   `json:"ResultType"`
	EffectiveStatus         string   `json:"EffectiveStatus"`
	Priority                string   `json:"Priority"`
	ControlId               string   `json:"ControlId"`
	ApplicableCloudProviders []string `json:"ApplicableCloudProviders"`
}

type CloudAccount struct {
	Name string `json:"Name"`
}

type Asset struct {
	State         string            `json:"State"`
	AssetUniqueId string            `json:"AssetUniqueId"`
	Type          string            `json:"Type"`
	Name          string            `json:"Name"`
	NewCategory   string            `json:"NewCategory"`
	Tags          map[string]string `json:"Tags,omitempty"`
}
