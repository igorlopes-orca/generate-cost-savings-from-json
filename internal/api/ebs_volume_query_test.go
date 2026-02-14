package api

import (
	"encoding/json"
	"testing"
)

func TestEBSVolumeQuery_AssetType(t *testing.T) {
	q := &EBSVolumeQuery{}
	if got := q.AssetType(); got != "AwsEc2EbsVolume" {
		t.Errorf("AssetType() = %q, want %q", got, "AwsEc2EbsVolume")
	}
}

func TestEBSVolumeQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "AwsEc2EbsVolume_506464807365_c46cb523-bf6f-32a1-9a67-4694eb2f159a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EBSVolumeQuery{}
			payload := q.BuildPayload(tt.assetUniqueID)

			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}

			// Check top-level fields
			if m["limit"] != float64(1) {
				t.Errorf("limit = %v, want 1", m["limit"])
			}
			if m["max_tier"] != float64(2) {
				t.Errorf("max_tier = %v, want 2", m["max_tier"])
			}

			// Check select is at top level
			sel, ok := m["select"].([]any)
			if !ok {
				t.Fatal("missing top-level select")
			}
			wantFields := map[string]bool{
				"Name": true, "VolumeSize": true, "VolumeType": true,
				"UsedDiskSize": true, "Region": true, "State": true,
			}
			for _, f := range sel {
				delete(wantFields, f.(string))
			}
			if len(wantFields) > 0 {
				t.Errorf("missing select fields: %v", wantFields)
			}

			// Check query models
			query, ok := m["query"].(map[string]any)
			if !ok {
				t.Fatal("missing query key")
			}
			models, ok := query["models"].([]any)
			if !ok || len(models) != 1 || models[0] != "AwsEc2EbsVolume" {
				t.Errorf("query.models = %v, want [AwsEc2EbsVolume]", models)
			}

			// Check with clause: VolumeSize exists (type=int) + AssetUniqueId filter
			with, ok := query["with"].(map[string]any)
			if !ok {
				t.Fatal("missing query.with")
			}
			values, ok := with["values"].([]any)
			if !ok {
				t.Fatal("missing query.with.values")
			}

			// First filter: VolumeSize exists with type=int
			sizeFilter, ok := values[0].(map[string]any)
			if !ok {
				t.Fatal("first filter is not an object")
			}
			if sizeFilter["key"] != "VolumeSize" {
				t.Errorf("first filter key = %v, want VolumeSize", sizeFilter["key"])
			}
			if sizeFilter["type"] != "int" {
				t.Errorf("VolumeSize filter type = %v, want int", sizeFilter["type"])
			}
			if sizeFilter["operator"] != "exists" {
				t.Errorf("VolumeSize filter operator = %v, want exists", sizeFilter["operator"])
			}

			// Second filter: AssetUniqueId
			idFilter, ok := values[1].(map[string]any)
			if !ok {
				t.Fatal("second filter is not an object")
			}
			if idFilter["key"] != "AssetUniqueId" {
				t.Errorf("second filter key = %v, want AssetUniqueId", idFilter["key"])
			}
			filterValues, ok := idFilter["values"].([]any)
			if !ok || len(filterValues) != 1 || filterValues[0] != tt.assetUniqueID {
				t.Errorf("AssetUniqueId filter values = %v, want [%s]", filterValues, tt.assetUniqueID)
			}
		})
	}
}

func TestEBSVolumeQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps all EBS fields",
			node: &OrcaAssetNode{
				ID:            "abc-123",
				Name:          "vol-test",
				Type:          "AwsEc2EbsVolume",
				AssetUniqueID: "AwsEc2EbsVolume_123_abc",
				Data: map[string]json.RawMessage{
					"Region":       json.RawMessage(`{"value":"eu-west-1"}`),
					"State":        json.RawMessage(`{"value":"available"}`),
					"VolumeSize":   json.RawMessage(`{"value":100}`),
					"VolumeType":   json.RawMessage(`{"value":"gp3"}`),
					"UsedDiskSize": json.RawMessage(`{"value":42.5}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "AwsEc2EbsVolume_123_abc",
				Type:          "AwsEc2EbsVolume",
				Name:          "vol-test",
				Region:        "eu-west-1",
				State:         "available",
				SizeGB:        100,
				UsedGB:        42.5,
				VolumeType:    "gp3",
			},
		},
		{
			name: "handles missing optional fields gracefully",
			node: &OrcaAssetNode{
				ID:            "abc-456",
				Name:          "vol-minimal",
				Type:          "AwsEc2EbsVolume",
				AssetUniqueID: "AwsEc2EbsVolume_456_def",
				Data:          map[string]json.RawMessage{},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "AwsEc2EbsVolume_456_def",
				Type:          "AwsEc2EbsVolume",
				Name:          "vol-minimal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EBSVolumeQuery{}
			got, err := q.MapResponse(tt.node)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.AssetUniqueID != tt.wantAsset.AssetUniqueID {
				t.Errorf("AssetUniqueID = %q, want %q", got.AssetUniqueID, tt.wantAsset.AssetUniqueID)
			}
			if got.Type != tt.wantAsset.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantAsset.Type)
			}
			if got.Name != tt.wantAsset.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantAsset.Name)
			}
			if got.Region != tt.wantAsset.Region {
				t.Errorf("Region = %q, want %q", got.Region, tt.wantAsset.Region)
			}
			if got.State != tt.wantAsset.State {
				t.Errorf("State = %q, want %q", got.State, tt.wantAsset.State)
			}
			if got.SizeGB != tt.wantAsset.SizeGB {
				t.Errorf("SizeGB = %v, want %v", got.SizeGB, tt.wantAsset.SizeGB)
			}
			if got.UsedGB != tt.wantAsset.UsedGB {
				t.Errorf("UsedGB = %v, want %v", got.UsedGB, tt.wantAsset.UsedGB)
			}
			if got.VolumeType != tt.wantAsset.VolumeType {
				t.Errorf("VolumeType = %q, want %q", got.VolumeType, tt.wantAsset.VolumeType)
			}
		})
	}
}
