package api

// GCPVMDiskSnapshotsQuery implements DiskSnapshotQuery for GcpVmDisk.
// It fetches a single GCP disk along with its nested Snapshots.
type GCPVMDiskSnapshotsQuery struct{}

func (q *GCPVMDiskSnapshotsQuery) DiskAssetType() string {
	return "GcpVmDisk"
}

func (q *GCPVMDiskSnapshotsQuery) BuildPayload(diskUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"GcpVmDisk"},
			"type":   "object_set",
			"with": map[string]any{
				"operator": "and",
				"type":     "operation",
				"values": []any{
					map[string]any{
						"keys":     []string{"Snapshots"},
						"models":   []string{"GcpVmSnapshot"},
						"type":     "object_set",
						"operator": "has",
					},
					map[string]any{
						"key":      "AssetUniqueId",
						"values":   []string{diskUniqueID},
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
			"Snapshots.Name",
			"Snapshots.SizeGb",
			"Snapshots.CreationTime",
		},
		"full_graph_fetch": map[string]any{
			"enabled": true,
		},
		"max_tier": 2,
	}
}

func (q *GCPVMDiskSnapshotsQuery) MapResponse(node *OrcaAssetNode) ([]SnapshotInfo, error) {
	snapNodes := ExtractNestedObjects(node.Data, "Snapshots")
	var snapshots []SnapshotInfo
	for _, sn := range snapNodes {
		snapshots = append(snapshots, SnapshotInfo{
			ID:        sn.AssetUniqueID,
			SizeGB:    ExtractFloat(sn.Data, "SizeGb"),
			CreatedAt: ExtractTime(sn.Data, "CreationTime"),
		})
	}
	return snapshots, nil
}
