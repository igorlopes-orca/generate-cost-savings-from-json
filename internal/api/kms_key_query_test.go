package api

import (
	"encoding/json"
	"testing"
)

func TestKMSKeyQuery_AssetType(t *testing.T) {
	q := &KMSKeyQuery{}
	if got := q.AssetType(); got != "AwsKmsKey" {
		t.Errorf("AssetType() = %q, want %q", got, "AwsKmsKey")
	}
}

func TestKMSKeyQuery_BuildPayload(t *testing.T) {
	tests := []struct {
		name          string
		assetUniqueID string
	}{
		{
			name:          "builds payload with asset unique ID",
			assetUniqueID: "AwsKmsKey_412128479389_abc-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &KMSKeyQuery{}
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
			if !ok || len(models) != 1 || models[0] != "AwsKmsKey" {
				t.Errorf("query.models = %v, want [AwsKmsKey]", models)
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

func TestKMSKeyQuery_MapResponse(t *testing.T) {
	tests := []struct {
		name      string
		node      *OrcaAssetNode
		wantAsset *AssetDetails
	}{
		{
			name: "maps KMS key fields",
			node: &OrcaAssetNode{
				ID:            "kms-123",
				Name:          "alias/eks/demo-agentic-412128479389",
				Type:          "AwsKmsKey",
				AssetUniqueID: "AwsKmsKey_123_abc",
				Data: map[string]json.RawMessage{
					"Region": json.RawMessage(`{"value":"us-east-1"}`),
					"State":  json.RawMessage(`{"value":"Disabled"}`),
				},
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "AwsKmsKey_123_abc",
				Type:          "AwsKmsKey",
				Name:          "alias/eks/demo-agentic-412128479389",
				Region:        "us-east-1",
				State:         "Disabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &KMSKeyQuery{}
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
