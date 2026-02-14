package api

import (
	"encoding/json"
	"testing"
)

func TestEC2InstanceQuery_AssetType(t *testing.T) {
	q := &EC2InstanceQuery{}
	if got := q.AssetType(); got != "AwsEc2Instance" {
		t.Errorf("AssetType() = %q, want %q", got, "AwsEc2Instance")
	}
}

func TestEC2InstanceQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "vm_506464807365_i-008e811c26290be21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EC2InstanceQuery{}
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
			if !ok || len(models) != 1 || models[0] != "AwsEc2Instance" {
				t.Errorf("query.models = %v, want [AwsEc2Instance]", models)
			}

			// Check select at top level includes EBS volume sub-fields
			sel, ok := m["select"].([]any)
			if !ok {
				t.Fatal("missing top-level select")
			}
			wantFields := map[string]bool{
				"Name":                     true,
				"State":                    true,
				"AssetUniqueId":            true,
				"Ec2EbsVolumes.Name":       true,
				"Ec2EbsVolumes.VolumeSize": true,
				"Ec2EbsVolumes.VolumeType": true,
			}
			for _, f := range sel {
				delete(wantFields, f.(string))
			}
			if len(wantFields) > 0 {
				t.Errorf("missing select fields: %v", wantFields)
			}

			// Check the with clause contains both the "has" (Ec2EbsVolumes) and Name filter
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

			// First value: "has" Ec2EbsVolumes
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

func TestEC2InstanceQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps instance with attached EBS volumes",
			node: &OrcaAssetNode{
				ID:            "c46cb523-3281-7768-02d6-56440a0c2c74",
				Name:          "Web-Nginx",
				Type:          "AwsEc2Instance",
				AssetUniqueID: "vm_506464807365_i-008e811c26290be21",
				Data: map[string]json.RawMessage{
					"State":          json.RawMessage(`{"value":"stopped"}`),
					"AssetUniqueId":  json.RawMessage(`{"value":"vm_506464807365_i-008e811c26290be21"}`),
					"Ec2EbsVolumes": json.RawMessage(`{
						"count": 2,
						"value": [
							{
								"id": "vol-1",
								"name": "vol-058c04bfd7778fcfc",
								"type": "AwsEc2EbsVolume",
								"asset_unique_id": "AwsEc2EbsVolume_506464807365_vol1",
								"data": {
									"Name": {"value": "vol-058c04bfd7778fcfc"},
									"VolumeSize": {"value": 8},
									"VolumeType": {"value": "gp2"}
								}
							},
							{
								"id": "vol-2",
								"name": "vol-0abcdef1234567890",
								"type": "AwsEc2EbsVolume",
								"asset_unique_id": "AwsEc2EbsVolume_506464807365_vol2",
								"data": {
									"Name": {"value": "vol-0abcdef1234567890"},
									"VolumeSize": {"value": 100},
									"VolumeType": {"value": "gp3"}
								}
							}
						]
					}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "vm_506464807365_i-008e811c26290be21",
				Type:          "AwsEc2Instance",
				Name:          "Web-Nginx",
				State:         "stopped",
				Disks: []DiskInfo{
					{
						ID:         "AwsEc2EbsVolume_506464807365_vol1",
						Name:       "vol-058c04bfd7778fcfc",
						SizeGB:     8,
						VolumeType: "gp2",
					},
					{
						ID:         "AwsEc2EbsVolume_506464807365_vol2",
						Name:       "vol-0abcdef1234567890",
						SizeGB:     100,
						VolumeType: "gp3",
					},
				},
			},
		},
		{
			name: "handles instance with no EBS volumes gracefully",
			node: &OrcaAssetNode{
				ID:            "abc-123",
				Name:          "no-disk-instance",
				Type:          "AwsEc2Instance",
				AssetUniqueID: "vm_123_i-000",
				Data: map[string]json.RawMessage{
					"State": json.RawMessage(`{"value":"stopped"}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "vm_123_i-000",
				Type:          "AwsEc2Instance",
				Name:          "no-disk-instance",
				State:         "stopped",
				Disks:         nil,
			},
		},
		{
			name: "handles missing data fields gracefully",
			node: &OrcaAssetNode{
				ID:            "abc-456",
				Name:          "minimal-instance",
				Type:          "AwsEc2Instance",
				AssetUniqueID: "vm_456_i-111",
				Data:          map[string]json.RawMessage{},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "vm_456_i-111",
				Type:          "AwsEc2Instance",
				Name:          "minimal-instance",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &EC2InstanceQuery{}
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
				if gotDisk.SizeGB != wantDisk.SizeGB {
					t.Errorf("Disk[%d].SizeGB = %v, want %v", i, gotDisk.SizeGB, wantDisk.SizeGB)
				}
				if gotDisk.VolumeType != wantDisk.VolumeType {
					t.Errorf("Disk[%d].VolumeType = %q, want %q", i, gotDisk.VolumeType, wantDisk.VolumeType)
				}
			}
		})
	}
}
