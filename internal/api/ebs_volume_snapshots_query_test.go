package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEBSVolumeSnapshotsQuery_DiskAssetType(t *testing.T) {
	q := &EBSVolumeSnapshotsQuery{}
	if got := q.DiskAssetType(); got != "AwsEc2EbsVolume" {
		t.Errorf("DiskAssetType() = %q, want %q", got, "AwsEc2EbsVolume")
	}
}

func TestEBSVolumeSnapshotsQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name         string
		diskUniqueID string
	}{
		{
			name:         "builds payload with disk unique ID",
			diskUniqueID: "AwsEc2EbsVolume_506464807365_vol-058c04bfd7778fcfc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EBSVolumeSnapshotsQuery{}
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
			if !ok || len(models) != 1 || models[0] != "AwsEc2EbsVolume" {
				t.Errorf("query.models = %v, want [AwsEc2EbsVolume]", models)
			}

			// Check the with clause contains both "has" (Ec2EbsSnapshots) and AssetUniqueId filter
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

			// First value: "has" Ec2EbsSnapshots
			hasFilter, ok := values[0].(map[string]any)
			if !ok {
				t.Fatal("first filter is not an object")
			}
			if hasFilter["operator"] != "has" {
				t.Errorf("first filter operator = %v, want has", hasFilter["operator"])
			}
			keys, ok := hasFilter["keys"].([]any)
			if !ok || len(keys) != 1 || keys[0] != "Ec2EbsSnapshots" {
				t.Errorf("has keys = %v, want [Ec2EbsSnapshots]", keys)
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
				"Name":                          true,
				"VolumeSize":                    true,
				"VolumeType":                    true,
				"Ec2EbsSnapshots.Name":          true,
				"Ec2EbsSnapshots.VolumeSize":    true,
				"Ec2EbsSnapshots.CreationTime":  true,
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

func TestEBSVolumeSnapshotsQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name          string
		node          *OrcaAssetNode
		wantSnapshots []SnapshotInfo
	}{
		{
			name: "maps volume with nested snapshots",
			node: &OrcaAssetNode{
				ID:            "vol-1",
				Name:          "vol-058c04bfd7778fcfc",
				Type:          "AwsEc2EbsVolume",
				AssetUniqueID: "AwsEc2EbsVolume_506464807365_vol1",
				Data: map[string]json.RawMessage{
					"VolumeSize": json.RawMessage(`{"value": 100}`),
					"VolumeType": json.RawMessage(`{"value": "gp2"}`),
					"Ec2EbsSnapshots": json.RawMessage(`{
						"count": 2,
						"value": [
							{
								"id": "snap-1",
								"name": "snap-0aaa",
								"type": "AwsEc2EbsSnapshot",
								"asset_unique_id": "AwsEc2EbsSnapshot_506464807365_snap1",
								"data": {
									"Name": {"value": "snap-0aaa"},
									"VolumeSize": {"value": 100},
									"CreationTime": {"value": "2020-09-02T09:09:34+00:00"}
								}
							},
							{
								"id": "snap-2",
								"name": "snap-0bbb",
								"type": "AwsEc2EbsSnapshot",
								"asset_unique_id": "AwsEc2EbsSnapshot_506464807365_snap2",
								"data": {
									"Name": {"value": "snap-0bbb"},
									"VolumeSize": {"value": 50},
									"CreationTime": {"value": "2024-01-15T12:30:00+00:00"}
								}
							}
						]
					}`),
				},
			},
			wantSnapshots: []SnapshotInfo{
				{
					ID:        "AwsEc2EbsSnapshot_506464807365_snap1",
					SizeGB:    100,
					CreatedAt: time.Date(2020, 9, 2, 9, 9, 34, 0, time.FixedZone("", 0)),
				},
				{
					ID:        "AwsEc2EbsSnapshot_506464807365_snap2",
					SizeGB:    50,
					CreatedAt: time.Date(2024, 1, 15, 12, 30, 0, 0, time.FixedZone("", 0)),
				},
			},
		},
		{
			name: "handles volume with no snapshots",
			node: &OrcaAssetNode{
				ID:            "vol-2",
				Name:          "empty-vol",
				Type:          "AwsEc2EbsVolume",
				AssetUniqueID: "AwsEc2EbsVolume_123_vol2",
				Data: map[string]json.RawMessage{
					"VolumeSize": json.RawMessage(`{"value": 50}`),
				},
			},
			wantSnapshots: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EBSVolumeSnapshotsQuery{}
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
