package api

// EBSVolumeSnapshotsQuery implements DiskSnapshotQuery for AwsEc2EbsVolume.
// It fetches a single EBS volume along with its nested Ec2EbsSnapshots.
type EBSVolumeSnapshotsQuery struct{}

func (q *EBSVolumeSnapshotsQuery) DiskAssetType() string {
	return "AwsEc2EbsVolume"
}

func (q *EBSVolumeSnapshotsQuery) BuildPayload(diskUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"AwsEc2EbsVolume"},
			"type":   "object_set",
			"with": map[string]any{
				"operator": "and",
				"type":     "operation",
				"values": []any{
					map[string]any{
						"keys":     []string{"Ec2EbsSnapshots"},
						"models":   []string{"AwsEc2EbsSnapshot"},
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
			"VolumeSize",
			"VolumeType",
			"Ec2EbsSnapshots.Name",
			"Ec2EbsSnapshots.VolumeSize",
			"Ec2EbsSnapshots.CreationTime",
		},
		"full_graph_fetch": map[string]any{
			"enabled": true,
		},
		"max_tier": 2,
	}
}

func (q *EBSVolumeSnapshotsQuery) MapResponse(node *OrcaAssetNode) ([]SnapshotInfo, error) {
	snapNodes := ExtractNestedObjects(node.Data, "Ec2EbsSnapshots")
	var snapshots []SnapshotInfo
	for _, sn := range snapNodes {
		snapshots = append(snapshots, SnapshotInfo{
			ID:        sn.AssetUniqueID,
			SizeGB:    ExtractFloat(sn.Data, "VolumeSize"),
			CreatedAt: ExtractTime(sn.Data, "CreationTime"),
		})
	}
	return snapshots, nil
}
