package api

// GCPVMInstanceQuery implements AssetQuery for GcpVmInstance assets.
// It fetches the instance along with its attached disks so that
// snapshot-related calculators can identify per-disk snapshots.
type GCPVMInstanceQuery struct{}

func (q *GCPVMInstanceQuery) AssetType() string {
	return "GcpVmInstance"
}

func (q *GCPVMInstanceQuery) BuildPayload(assetUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"GcpVmInstance"},
			"type":   "object_set",
			"with": map[string]any{
				"operator": "and",
				"type":     "operation",
				"values": []any{
					map[string]any{
						"keys":     []string{"InstanceDisks"},
						"models":   []string{"GcpVmInstanceDisk"},
						"type":     "object_set",
						"operator": "has",
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
			"State",
			"AssetUniqueId",
			"InstanceDisks.Name",
		},
		"full_graph_fetch": map[string]any{
			"enabled": true,
		},
		"max_tier": 2,
	}
}

func (q *GCPVMInstanceQuery) MapResponse(node *OrcaAssetNode) (*AssetDetails, error) {
	d := node.Data

	disks := ExtractNestedObjects(d, "InstanceDisks")
	var diskInfos []DiskInfo
	for _, disk := range disks {
		diskInfos = append(diskInfos, DiskInfo{
			ID:   disk.AssetUniqueID,
			Name: disk.Name,
		})
	}

	return &AssetDetails{
		AssetUniqueID: node.AssetUniqueID,
		Type:          node.Type,
		Name:          node.Name,
		State:         ExtractString(d, "State"),
		Disks:         diskInfos,
	}, nil
}
