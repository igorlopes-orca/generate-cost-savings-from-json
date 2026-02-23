package api

import (
	"encoding/json"
	"time"
)

// AssetQuery defines a per-asset-type query builder and response mapper.
// Each supported Orca asset type (e.g. AwsEc2EbsVolume, AwsEc2Instance)
// implements this interface with its own payload structure and field mapping.
type AssetQuery interface {
	// AssetType returns the Orca model name this query handles (e.g. "AwsEc2EbsVolume").
	AssetType() string

	// BuildPayload returns the request body for the serving-layer query API.
	BuildPayload(assetUniqueID string) any

	// MapResponse converts an Orca API response node into an AssetDetails.
	MapResponse(node *OrcaAssetNode) (*AssetDetails, error)
}

// OrcaAssetNode represents a single asset node in the Orca serving-layer response.
type OrcaAssetNode struct {
	ID            string                     `json:"id"`
	Name          string                     `json:"name"`
	Type          string                     `json:"type"`
	AssetUniqueID string                     `json:"asset_unique_id"`
	Data          map[string]json.RawMessage `json:"data"`
}

// orcaValue wraps the {"value": ...} pattern used by Orca data fields.
type orcaValue[T any] struct {
	Value T `json:"value"`
}

// ExtractString reads a string field from the Orca data map.
func ExtractString(data map[string]json.RawMessage, key string) string {
	raw, ok := data[key]
	if !ok {
		return ""
	}
	var v orcaValue[string]
	if err := json.Unmarshal(raw, &v); err != nil {
		return ""
	}
	return v.Value
}

// ExtractFloat reads a numeric field from the Orca data map.
func ExtractFloat(data map[string]json.RawMessage, key string) float64 {
	raw, ok := data[key]
	if !ok {
		return 0
	}
	var v orcaValue[float64]
	if err := json.Unmarshal(raw, &v); err != nil {
		return 0
	}
	return v.Value
}

// orcaNestedObjects wraps the {"count": N, "value": [...]} pattern for related objects.
type orcaNestedObjects struct {
	Count int             `json:"count"`
	Value []OrcaAssetNode `json:"value"`
}

// ExtractNestedObjects reads a nested object array (e.g. Ec2EbsVolumes) from the Orca data map.
func ExtractNestedObjects(data map[string]json.RawMessage, key string) []OrcaAssetNode {
	raw, ok := data[key]
	if !ok {
		return nil
	}
	var nested orcaNestedObjects
	if err := json.Unmarshal(raw, &nested); err != nil {
		return nil
	}
	return nested.Value
}

// ExtractTime reads a time field from the Orca data map, parsing RFC3339-formatted strings.
func ExtractTime(data map[string]json.RawMessage, key string) time.Time {
	s := ExtractString(data, key)
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// DiskSnapshotQuery defines a per-disk-type query that fetches a disk with its nested snapshots.
type DiskSnapshotQuery interface {
	// DiskAssetType returns the Orca disk model this query handles (e.g. "AwsEc2EbsVolume").
	DiskAssetType() string

	// BuildPayload returns the request body to query a disk with nested snapshots.
	BuildPayload(diskUniqueID string) any

	// MapResponse extracts snapshot info from the response node.
	MapResponse(node *OrcaAssetNode) ([]SnapshotInfo, error)
}
