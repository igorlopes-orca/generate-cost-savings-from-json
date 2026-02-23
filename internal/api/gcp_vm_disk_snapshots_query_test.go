package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGCPVMDiskSnapshotsQuery_DiskAssetType(t *testing.T) {
	q := &GCPVMDiskSnapshotsQuery{}
	if got := q.DiskAssetType(); got != "GcpVmDisk" {
		t.Errorf("DiskAssetType() = %q, want %q", got, "GcpVmDisk")
	}
}

func TestGCPVMDiskSnapshotsQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name         string
		diskUniqueID string
	}{
		{
			name:         "builds payload with disk unique ID",
			diskUniqueID: "GcpVmDisk_proj123_disk-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPVMDiskSnapshotsQuery{}
			payload := q.BuildPayload(tt.diskUniqueID)

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
			if !ok || len(models) != 1 || models[0] != "GcpVmDisk" {
				t.Errorf("query.models = %v, want [GcpVmDisk]", models)
			}

			// Check the with clause contains both "has" (Snapshots) and AssetUniqueId filter
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

			// First value: "has" Snapshots
			hasFilter, ok := values[0].(map[string]any)
			if !ok {
				t.Fatal("first filter is not an object")
			}
			if hasFilter["operator"] != "has" {
				t.Errorf("first filter operator = %v, want has", hasFilter["operator"])
			}
			keys, ok := hasFilter["keys"].([]any)
			if !ok || len(keys) != 1 || keys[0] != "Snapshots" {
				t.Errorf("has keys = %v, want [Snapshots]", keys)
			}
			hasModels, ok := hasFilter["models"].([]any)
			if !ok || len(hasModels) != 1 || hasModels[0] != "GcpVmSnapshot" {
				t.Errorf("has models = %v, want [GcpVmSnapshot]", hasModels)
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
			if !ok || len(filterValues) != 1 || filterValues[0] != tt.diskUniqueID {
				t.Errorf("AssetUniqueId filter values = %v, want [%s]", filterValues, tt.diskUniqueID)
			}

			// Check select includes snapshot sub-fields
			sel, ok := m["select"].([]any)
			if !ok {
				t.Fatal("missing top-level select")
			}
			wantFields := map[string]bool{
				"Name":                     true,
				"SizeGb":                   true,
				"VolumeType":               true,
				"Snapshots.Name":           true,
				"Snapshots.SizeGb":         true,
				"Snapshots.CreationTime":   true,
			}
			for _, f := range sel {
				delete(wantFields, f.(string))
			}
			if len(wantFields) > 0 {
				t.Errorf("missing select fields: %v", wantFields)
			}
		})
	}
}

func TestGCPVMDiskSnapshotsQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name          string
		node          *OrcaAssetNode
		wantSnapshots []SnapshotInfo
	}{
		{
			name: "maps disk with nested snapshots",
			node: &OrcaAssetNode{
				ID:            "disk-1",
				Name:          "my-disk",
				Type:          "GcpVmDisk",
				AssetUniqueID: "GcpVmDisk_proj123_disk1",
				Data: map[string]json.RawMessage{
					"SizeGb":     json.RawMessage(`{"value": 200}`),
					"VolumeType": json.RawMessage(`{"value": "pd-standard"}`),
					"Snapshots": json.RawMessage(`{
						"count": 2,
						"value": [
							{
								"id": "snap-1",
								"name": "snap-old",
								"type": "GcpVmSnapshot",
								"asset_unique_id": "GcpVmSnapshot_proj123_snap1",
								"data": {
									"Name": {"value": "snap-old"},
									"SizeGb": {"value": 200},
									"CreationTime": {"value": "2020-06-15T08:00:00+00:00"}
								}
							},
							{
								"id": "snap-2",
								"name": "snap-new",
								"type": "GcpVmSnapshot",
								"asset_unique_id": "GcpVmSnapshot_proj123_snap2",
								"data": {
									"Name": {"value": "snap-new"},
									"SizeGb": {"value": 100},
									"CreationTime": {"value": "2025-12-01T10:00:00+00:00"}
								}
							}
						]
					}`),
				},
			},
			wantSnapshots: []SnapshotInfo{
				{
					ID:        "GcpVmSnapshot_proj123_snap1",
					SizeGB:    200,
					CreatedAt: time.Date(2020, 6, 15, 8, 0, 0, 0, time.FixedZone("", 0)),
				},
				{
					ID:        "GcpVmSnapshot_proj123_snap2",
					SizeGB:    100,
					CreatedAt: time.Date(2025, 12, 1, 10, 0, 0, 0, time.FixedZone("", 0)),
				},
			},
		},
		{
			name: "handles disk with no snapshots",
			node: &OrcaAssetNode{
				ID:            "disk-2",
				Name:          "empty-disk",
				Type:          "GcpVmDisk",
				AssetUniqueID: "GcpVmDisk_proj123_disk2",
				Data: map[string]json.RawMessage{
					"SizeGb": json.RawMessage(`{"value": 50}`),
				},
			},
			wantSnapshots: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPVMDiskSnapshotsQuery{}
			got, err := q.MapResponse(tt.node)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.wantSnapshots) {
				t.Fatalf("snapshot count = %d, want %d", len(got), len(tt.wantSnapshots))
			}
			for i, want := range tt.wantSnapshots {
				g := got[i]
				if g.ID != want.ID {
					t.Errorf("Snapshot[%d].ID = %q, want %q", i, g.ID, want.ID)
				}
				if g.SizeGB != want.SizeGB {
					t.Errorf("Snapshot[%d].SizeGB = %v, want %v", i, g.SizeGB, want.SizeGB)
				}
				if !g.CreatedAt.Equal(want.CreatedAt) {
					t.Errorf("Snapshot[%d].CreatedAt = %v, want %v", i, g.CreatedAt, want.CreatedAt)
				}
			}
		})
	}
}
