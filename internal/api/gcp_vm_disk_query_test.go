package api

import (
	"encoding/json"
	"testing"
)

func TestGCPVMDiskQuery_AssetType(t *testing.T) {
	q := &GCPVMDiskQuery{}
	if got := q.AssetType(); got != "GcpVmDisk" {
		t.Errorf("AssetType() = %q, want %q", got, "GcpVmDisk")
	}
}

func TestGCPVMDiskQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "GcpVmDisk_project123_disk-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPVMDiskQuery{}
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
				"Name": true, "SizeGb": true, "VolumeType": true,
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
			if !ok || len(models) != 1 || models[0] != "GcpVmDisk" {
				t.Errorf("query.models = %v, want [GcpVmDisk]", models)
			}

			// Check with clause: SizeGb exists (type=int) + VolumeType exists (type=str) + AssetUniqueId filter
			with, ok := query["with"].(map[string]any)
			if !ok {
				t.Fatal("missing query.with")
			}
			values, ok := with["values"].([]any)
			if !ok {
				t.Fatal("missing query.with.values")
			}
			if len(values) != 3 {
				t.Fatalf("expected 3 with values, got %d", len(values))
			}

			// First filter: SizeGb exists with type=int
			sizeFilter, ok := values[0].(map[string]any)
			if !ok {
				t.Fatal("first filter is not an object")
			}
			if sizeFilter["key"] != "SizeGb" {
				t.Errorf("first filter key = %v, want SizeGb", sizeFilter["key"])
			}
			if sizeFilter["type"] != "int" {
				t.Errorf("SizeGb filter type = %v, want int", sizeFilter["type"])
			}
			if sizeFilter["operator"] != "exists" {
				t.Errorf("SizeGb filter operator = %v, want exists", sizeFilter["operator"])
			}

			// Second filter: VolumeType exists with type=str
			typeFilter, ok := values[1].(map[string]any)
			if !ok {
				t.Fatal("second filter is not an object")
			}
			if typeFilter["key"] != "VolumeType" {
				t.Errorf("second filter key = %v, want VolumeType", typeFilter["key"])
			}
			if typeFilter["type"] != "str" {
				t.Errorf("VolumeType filter type = %v, want str", typeFilter["type"])
			}
			if typeFilter["operator"] != "exists" {
				t.Errorf("VolumeType filter operator = %v, want exists", typeFilter["operator"])
			}

			// Third filter: AssetUniqueId
			idFilter, ok := values[2].(map[string]any)
			if !ok {
				t.Fatal("third filter is not an object")
			}
			if idFilter["key"] != "AssetUniqueId" {
				t.Errorf("third filter key = %v, want AssetUniqueId", idFilter["key"])
			}
			filterValues, ok := idFilter["values"].([]any)
			if !ok || len(filterValues) != 1 || filterValues[0] != tt.assetUniqueID {
				t.Errorf("AssetUniqueId filter values = %v, want [%s]", filterValues, tt.assetUniqueID)
			}
		})
	}
}

func TestGCPVMDiskQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps all GCP disk fields",
			node: &OrcaAssetNode{
				ID:            "disk-123",
				Name:          "my-gcp-disk",
				Type:          "GcpVmDisk",
				AssetUniqueID: "GcpVmDisk_project123_disk-abc",
				Data: map[string]json.RawMessage{
					"Region":       json.RawMessage(`{"value":"us-central1"}`),
					"State":        json.RawMessage(`{"value":"READY"}`),
					"SizeGb":       json.RawMessage(`{"value":200}`),
					"VolumeType":   json.RawMessage(`{"value":"pd-ssd"}`),
					"UsedDiskSize": json.RawMessage(`{"value":85.5}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "GcpVmDisk_project123_disk-abc",
				Type:          "GcpVmDisk",
				Name:          "my-gcp-disk",
				Region:        "us-central1",
				State:         "READY",
				SizeGB:        200,
				UsedGB:        85.5,
				VolumeType:    "pd-ssd",
			},
		},
		{
			name: "handles missing optional fields gracefully",
			node: &OrcaAssetNode{
				ID:            "disk-456",
				Name:          "minimal-disk",
				Type:          "GcpVmDisk",
				AssetUniqueID: "GcpVmDisk_project456_disk-def",
				Data:          map[string]json.RawMessage{},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "GcpVmDisk_project456_disk-def",
				Type:          "GcpVmDisk",
				Name:          "minimal-disk",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPVMDiskQuery{}
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
