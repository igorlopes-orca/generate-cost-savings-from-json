package api

import (
	"encoding/json"
	"testing"
)

func TestGCPVMInstanceQuery_AssetType(t *testing.T) {
	q := &GCPVMInstanceQuery{}
	if got := q.AssetType(); got != "GcpVmInstance" {
		t.Errorf("AssetType() = %q, want %q", got, "GcpVmInstance")
	}
}

func TestGCPVMInstanceQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "vm_project123_instance-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPVMInstanceQuery{}
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
			if m["start_at_index"] != float64(0) {
				t.Errorf("start_at_index = %v, want 0", m["start_at_index"])
			}
			if m["max_tier"] != float64(2) {
				t.Errorf("max_tier = %v, want 2", m["max_tier"])
			}

			// Check full_graph_fetch
			fgf, ok := m["full_graph_fetch"].(map[string]any)
			if !ok {
				t.Fatal("missing full_graph_fetch")
			}
			if fgf["enabled"] != true {
				t.Errorf("full_graph_fetch.enabled = %v, want true", fgf["enabled"])
			}

			// Check query models
			query, ok := m["query"].(map[string]any)
			if !ok {
				t.Fatal("missing query key")
			}
			models, ok := query["models"].([]any)
			if !ok || len(models) != 1 || models[0] != "GcpVmInstance" {
				t.Errorf("query.models = %v, want [GcpVmInstance]", models)
			}

			// Check select at top level includes InstanceDisks sub-fields
			sel, ok := m["select"].([]any)
			if !ok {
				t.Fatal("missing top-level select")
			}
			wantFields := map[string]bool{
				"Name":               true,
				"State":              true,
				"AssetUniqueId":      true,
				"InstanceDisks.Name": true,
			}
			for _, f := range sel {
				delete(wantFields, f.(string))
			}
			if len(wantFields) > 0 {
				t.Errorf("missing select fields: %v", wantFields)
			}

			// Check the with clause contains both the "has" (InstanceDisks) and AssetUniqueId filter
			with, ok := query["with"].(map[string]any)
			if !ok {
				t.Fatal("missing query.with")
			}
			values, ok := with["values"].([]any)
			if !ok {
				t.Fatal("missing query.with.values")
			}
			if len(values) != 2 {
				t.Fatalf("expected 2 with values, got %d", len(values))
			}

			// First value: "has" InstanceDisks
			hasFilter, ok := values[0].(map[string]any)
			if !ok {
				t.Fatal("first filter is not an object")
			}
			if hasFilter["operator"] != "has" {
				t.Errorf("first filter operator = %v, want has", hasFilter["operator"])
			}

			// Second value: AssetUniqueId filter
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

func TestGCPVMInstanceQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps instance with attached disks",
			node: &OrcaAssetNode{
				ID:            "inst-123",
				Name:          "gcp-web-server",
				Type:          "GcpVmInstance",
				AssetUniqueID: "vm_project123_instance-abc",
				Data: map[string]json.RawMessage{
					"State":         json.RawMessage(`{"value":"TERMINATED"}`),
					"AssetUniqueId": json.RawMessage(`{"value":"vm_project123_instance-abc"}`),
					"InstanceDisks": json.RawMessage(`{
						"count": 2,
						"value": [
							{
								"id": "disk-1",
								"name": "boot-disk",
								"type": "GcpVmInstanceDisk",
								"asset_unique_id": "GcpVmDisk_project123_disk-1",
								"data": {
									"Name": {"value": "boot-disk"}
								}
							},
							{
								"id": "disk-2",
								"name": "data-disk",
								"type": "GcpVmInstanceDisk",
								"asset_unique_id": "GcpVmDisk_project123_disk-2",
								"data": {
									"Name": {"value": "data-disk"}
								}
							}
						]
					}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "vm_project123_instance-abc",
				Type:          "GcpVmInstance",
				Name:          "gcp-web-server",
				State:         "TERMINATED",
				Disks: []DiskInfo{
					{
						ID:   "GcpVmDisk_project123_disk-1",
						Name: "boot-disk",
					},
					{
						ID:   "GcpVmDisk_project123_disk-2",
						Name: "data-disk",
					},
				},
			},
		},
		{
			name: "handles instance with no disks gracefully",
			node: &OrcaAssetNode{
				ID:            "inst-456",
				Name:          "no-disk-instance",
				Type:          "GcpVmInstance",
				AssetUniqueID: "vm_project456_instance-def",
				Data: map[string]json.RawMessage{
					"State": json.RawMessage(`{"value":"RUNNING"}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "vm_project456_instance-def",
				Type:          "GcpVmInstance",
				Name:          "no-disk-instance",
				State:         "RUNNING",
				Disks:         nil,
			},
		},
		{
			name: "handles missing data fields gracefully",
			node: &OrcaAssetNode{
				ID:            "inst-789",
				Name:          "minimal-instance",
				Type:          "GcpVmInstance",
				AssetUniqueID: "vm_project789_instance-ghi",
				Data:          map[string]json.RawMessage{},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "vm_project789_instance-ghi",
				Type:          "GcpVmInstance",
				Name:          "minimal-instance",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPVMInstanceQuery{}
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
			if got.State != tt.wantAsset.State {
				t.Errorf("State = %q, want %q", got.State, tt.wantAsset.State)
			}

			if len(got.Disks) != len(tt.wantAsset.Disks) {
				t.Fatalf("Disks count = %d, want %d", len(got.Disks), len(tt.wantAsset.Disks))
			}
			for i, wantDisk := range tt.wantAsset.Disks {
				gotDisk := got.Disks[i]
				if gotDisk.ID != wantDisk.ID {
					t.Errorf("Disk[%d].ID = %q, want %q", i, gotDisk.ID, wantDisk.ID)
				}
				if gotDisk.Name != wantDisk.Name {
					t.Errorf("Disk[%d].Name = %q, want %q", i, gotDisk.Name, wantDisk.Name)
				}
			}
		})
	}
}
