package api

// EBSVolumeQuery implements AssetQuery for AwsEc2EbsVolume assets.
type EBSVolumeQuery struct{}

func (q *EBSVolumeQuery) AssetType() string {
	return "AwsEc2EbsVolume"
}

func (q *EBSVolumeQuery) BuildPayload(assetUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"AwsEc2EbsVolume"},
			"type":   "object_set",
			"with": map[string]any{
				"operator": "and",
				"type":     "operation",
				"values": []any{
					map[string]any{
						"key":      "VolumeSize",
						"values":   []any{},
						"type":     "int",
						"operator": "exists",
					},
					map[string]any{
						"key":      "AssetUniqueId",
						"values":   []string{assetUniqueID},
						"type":     "str",
						"operator": "in",
					},
				},
			},
		},
		"limit":          1,
		"start_at_index": 0,
		"select": []string{
			"Name",
			"VolumeSize",
			"VolumeType",
			"UsedDiskSize",
			"Region",
			"State",
		},
		"full_graph_fetch": map[string]any{
			"enabled": true,
		},
		"max_tier": 2,
	}
}

func (q *EBSVolumeQuery) MapResponse(node *OrcaAssetNode) (*AssetDetails, error) {
	d := node.Data
	return &AssetDetails{
		AssetUniqueID: node.AssetUniqueID,
		Type:          node.Type,
		Name:          node.Name,
		Region:        ExtractString(d, "Region"),
		State:         ExtractString(d, "State"),
		SizeGB:        ExtractFloat(d, "VolumeSize"),
		UsedGB:        ExtractFloat(d, "UsedDiskSize"),
		VolumeType:    ExtractString(d, "VolumeType"),
	}, nil
}
