package api

import (
	"encoding/json"
	"testing"
)

func TestGCPKMSKeyQuery_AssetType(t *testing.T) {
	q := &GCPKMSKeyQuery{}
	if got := q.AssetType(); got != "GcpKmsKey" {
		t.Errorf("AssetType() = %q, want %q", got, "GcpKmsKey")
	}
}

func TestGCPKMSKeyQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "GcpKmsKey_project123_key-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPKMSKeyQuery{}
			payload := q.BuildPayload(tt.assetUniqueID)

			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}

			query, ok := m["query"].(map[string]any)
			if !ok {
				t.Fatal("missing query key")
			}
			models, ok := query["models"].([]any)
			if !ok || len(models) != 1 || models[0] != "GcpKmsKey" {
				t.Errorf("query.models = %v, want [GcpKmsKey]", models)
			}

			with, ok := query["with"].(map[string]any)
			if !ok {
				t.Fatal("missing query.with")
			}
			if with["key"] != "AssetUniqueId" {
				t.Errorf("filter key = %v, want AssetUniqueId", with["key"])
			}
			filterValues, ok := with["values"].([]any)
			if !ok || len(filterValues) != 1 || filterValues[0] != tt.assetUniqueID {
				t.Errorf("AssetUniqueId filter values = %v, want [%s]", filterValues, tt.assetUniqueID)
			}
		})
	}
}

func TestGCPKMSKeyQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps GCP KMS key fields",
			node: &OrcaAssetNode{
				ID:            "kms-123",
				Name:          "projects/my-project/locations/global/keyRings/my-ring/cryptoKeys/my-key",
				Type:          "GcpKmsKey",
				AssetUniqueID: "GcpKmsKey_project123_key-abc",
				Data: map[string]json.RawMessage{
					"Region": json.RawMessage(`{"value":"global"}`),
					"State":  json.RawMessage(`{"value":"DISABLED"}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "GcpKmsKey_project123_key-abc",
				Type:          "GcpKmsKey",
				Name:          "projects/my-project/locations/global/keyRings/my-ring/cryptoKeys/my-key",
				Region:        "global",
				State:         "DISABLED",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &GCPKMSKeyQuery{}
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
		})
	}
}
