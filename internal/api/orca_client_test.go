package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrcaClient_FetchAsset(t *testing.T) {
	tests := []struct {
		name            string
		assetType       string
		assetUniqueID   string
		serverHandler   http.HandlerFunc
		wantErr         bool
		wantErrContains string
		wantAsset       *AssetDetails
	}{
		{
			name:          "successful EBS volume fetch",
			assetType:     "AwsEc2EbsVolume",
			assetUniqueID: "AwsEc2EbsVolume_506464807365_c46cb523-bf6f-32a1-9a67-4694eb2f159a",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// Verify request structure
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/serving-layer/query" {
					t.Errorf("expected /api/serving-layer/query, got %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "TOKEN test-token" {
					t.Errorf("expected TOKEN test-token, got %s", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Verify request body uses the EBS volume query
				body, _ := io.ReadAll(r.Body)
				var reqBody map[string]any
				json.Unmarshal(body, &reqBody)
				if reqBody["limit"] != float64(1) {
					t.Errorf("expected limit=1, got %v", reqBody["limit"])
				}
				query, _ := reqBody["query"].(map[string]any)
				models, _ := query["models"].([]any)
				if len(models) == 0 || models[0] != "AwsEc2EbsVolume" {
					t.Errorf("expected model AwsEc2EbsVolume, got %v", models)
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
					"data": []map[string]any{
						{
							"id":              "c46cb523-3232-8a09-b9db-9ff358d87e99",
							"name":            "vol-0e6eeace1b112dabb",
							"type":            "AwsEc2EbsVolume",
							"asset_unique_id": "AwsEc2EbsVolume_506464807365_c46cb523-bf6f-32a1-9a67-4694eb2f159a",
							"data": map[string]any{
								"Region":       map[string]any{"value": "us-east-1"},
								"State":        map[string]any{"value": "in-use"},
								"VolumeSize":   map[string]any{"value": float64(8)},
								"VolumeType":   map[string]any{"value": "gp2"},
								"UsedDiskSize": map[string]any{"value": float64(6)},
							},
						},
					},
				})
			},
			wantAsset: &AssetDetails{
				AssetUniqueID: "AwsEc2EbsVolume_506464807365_c46cb523-bf6f-32a1-9a67-4694eb2f159a",
				Type:          "AwsEc2EbsVolume",
				Name:          "vol-0e6eeace1b112dabb",
				Region:        "us-east-1",
				State:         "in-use",
				SizeGB:        8,
				UsedGB:        6,
				VolumeType:    "gp2",
			},
		},
		{
			name:          "unsupported asset type",
			assetType:     "UnknownType",
			assetUniqueID: "something",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called for unknown asset type")
			},
			wantErr:         true,
			wantErrContains: "unsupported asset type",
		},
		{
			name:      "no asset found",
			assetType: "AwsEc2EbsVolume",
			assetUniqueID: "AwsEc2EbsVolume_123_doesnotexist",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"status": "success",
					"data":   []map[string]any{},
				})
			},
			wantErr:         true,
			wantErrContains: "no asset found",
		},
		{
			name:      "API returns error status",
			assetType: "AwsEc2EbsVolume",
			assetUniqueID: "AwsEc2EbsVolume_123_xxx",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"status": "error",
					"data":   []map[string]any{},
				})
			},
			wantErr:         true,
			wantErrContains: "status \"error\"",
		},
		{
			name:      "HTTP 500",
			assetType: "AwsEc2EbsVolume",
			assetUniqueID: "AwsEc2EbsVolume_123_xxx",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:         true,
			wantErrContains: "unexpected status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.serverHandler)
			defer srv.Close()

			client := NewOrcaClient(srv.URL, "test-token")
			got, err := client.FetchAsset(context.Background(), tt.assetType, tt.assetUniqueID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContains != "" {
					if !contains(err.Error(), tt.wantErrContains) {
						t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContains)
					}
				}
				return
			}

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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
