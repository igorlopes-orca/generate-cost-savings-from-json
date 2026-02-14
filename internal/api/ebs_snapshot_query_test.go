package api

import (
	"encoding/json"
	"testing"
)

func TestEBSSnapshotQuery_AssetType(t *testing.T) {
	q := &EBSSnapshotQuery{}
	if got := q.AssetType(); got != "AwsEc2EbsSnapshot" {
		t.Errorf("AssetType() = %q, want %q", got, "AwsEc2EbsSnapshot")
	}
}

func TestEBSSnapshotQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "AwsEc2EbsSnapshot_506464807365_c46cb523-b8d6-5797-0afc-4db1223eb3d3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EBSSnapshotQuery{}
			payload := q.BuildPayload(tt.assetUniqueID)

			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}

			// Check select is at top level
			sel, ok := m["select"].([]any)
			if !ok {
				t.Fatal("missing top-level select")
			}
			wantFields := map[string]bool{
				"Name": true, "VolumeSize": true, "UsedDiskSize": true,
				"AllocatedDiskSize": true, "Region": true, "State": true,
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
			if !ok || len(models) != 1 || models[0] != "AwsEc2EbsSnapshot" {
				t.Errorf("query.models = %v, want [AwsEc2EbsSnapshot]", models)
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
			if sizeFilter["type"] != "int" {
				t.Errorf("VolumeSize filter type = %v, want int", sizeFilter["type"])
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

func TestEBSSnapshotQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps snapshot fields from real response",
			node: &OrcaAssetNode{
				ID:            "c46cb523-3d78-e3ae-bf46-6f7f4fa54b2d",
				Name:          "Jenkins-01-prd-vol-snapshot",
				Type:          "AwsEc2EbsSnapshot",
				AssetUniqueID: "AwsEc2EbsSnapshot_506464807365_c46cb523-b8d6-5797-0afc-4db1223eb3d3",
				Data: map[string]json.RawMessage{
					"VolumeSize":   json.RawMessage(`{"value":8}`),
					"UsedDiskSize": json.RawMessage(`{"value":6}`),
					"Region":       json.RawMessage(`{"value":"us-east-1"}`),
					"State":        json.RawMessage(`{"value":"completed"}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "AwsEc2EbsSnapshot_506464807365_c46cb523-b8d6-5797-0afc-4db1223eb3d3",
				Type:          "AwsEc2EbsSnapshot",
				Name:          "Jenkins-01-prd-vol-snapshot",
				Region:        "us-east-1",
				State:         "completed",
				SizeGB:        8,
				UsedGB:        6,
			},
		},
		{
			name: "handles missing optional fields",
			node: &OrcaAssetNode{
				ID:            "abc-123",
				Name:          "snap-minimal",
				Type:          "AwsEc2EbsSnapshot",
				AssetUniqueID: "AwsEc2EbsSnapshot_123_abc",
				Data: map[string]json.RawMessage{
					"VolumeSize": json.RawMessage(`{"value":100}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "AwsEc2EbsSnapshot_123_abc",
				Type:          "AwsEc2EbsSnapshot",
				Name:          "snap-minimal",
				SizeGB:        100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EBSSnapshotQuery{}
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
			if got.SizeGB != tt.wantAsset.SizeGB {
				t.Errorf("SizeGB = %v, want %v", got.SizeGB, tt.wantAsset.SizeGB)
			}
			if got.UsedGB != tt.wantAsset.UsedGB {
				t.Errorf("UsedGB = %v, want %v", got.UsedGB, tt.wantAsset.UsedGB)
			}
			if got.Region != tt.wantAsset.Region {
				t.Errorf("Region = %q, want %q", got.Region, tt.wantAsset.Region)
			}
		})
	}
}
