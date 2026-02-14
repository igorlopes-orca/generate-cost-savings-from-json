package api

// KMSKeyQuery implements AssetQuery for AwsKmsKey assets.
type KMSKeyQuery struct{}

func (q *KMSKeyQuery) AssetType() string {
	return "AwsKmsKey"
}

func (q *KMSKeyQuery) BuildPayload(assetUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"AwsKmsKey"},
			"type":   "object_set",
			"with": map[string]any{
				"key":      "AssetUniqueId",
				"values":   []string{assetUniqueID},
				"type":     "str",
				"operator": "in",
			},
		},
		"limit":          1,
		"start_at_index": 0,
		"select": []string{
			"Name",
			"State",
			"Region",
		},
		"full_graph_fetch": map[string]any{
			"enabled": true,
		},
		"max_tier": 2,
	}
}

func (q *KMSKeyQuery) MapResponse(node *OrcaAssetNode) (*AssetDetails, error) {
	d := node.Data
	return &AssetDetails{
		AssetUniqueID: node.AssetUniqueID,
		Type:          node.Type,
		Name:          node.Name,
		Region:        ExtractString(d, "Region"),
		State:         ExtractString(d, "State"),
	}, nil
}
