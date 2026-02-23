package api

// GCPVMDiskQuery implements AssetQuery for GcpVmDisk assets.
type GCPVMDiskQuery struct{}

func (q *GCPVMDiskQuery) AssetType() string {
	return "GcpVmDisk"
}

func (q *GCPVMDiskQuery) BuildPayload(assetUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"GcpVmDisk"},
			"type":   "object_set",
			"with": map[string]any{
				"operator": "and",
				"type":     "operation",
				"values": []any{
					map[string]any{
						"key":      "SizeGb",
						"values":   []any{},
						"type":     "int",
						"operator": "exists",
					},
					map[string]any{
						"key":      "VolumeType",
						"values":   []any{},
						"type":     "str",
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
			"SizeGb",
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

func (q *GCPVMDiskQuery) MapResponse(node *OrcaAssetNode) (*AssetDetails, error) {
	d := node.Data
	return &AssetDetails{
		AssetUniqueID: node.AssetUniqueID,
		Type:          node.Type,
		Name:          node.Name,
		Region:        ExtractString(d, "Region"),
		State:         ExtractString(d, "State"),
		SizeGB:        ExtractFloat(d, "SizeGb"),
		UsedGB:        ExtractFloat(d, "UsedDiskSize"),
		VolumeType:    ExtractString(d, "VolumeType"),
	}, nil
}
