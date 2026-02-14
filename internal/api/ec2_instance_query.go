package api

// EC2InstanceQuery implements AssetQuery for AwsEc2Instance assets.
// It fetches the instance along with its attached EBS volumes so that
// the stopped-EC2 calculator can sum disk costs.
type EC2InstanceQuery struct{}

func (q *EC2InstanceQuery) AssetType() string {
	return "AwsEc2Instance"
}

func (q *EC2InstanceQuery) BuildPayload(assetUniqueID string) any {
	return map[string]any{
		"query": map[string]any{
			"models": []string{"AwsEc2Instance"},
			"type":   "object_set",
			"with": map[string]any{
				"operator": "and",
				"type":     "operation",
				"values": []any{
					map[string]any{
						"keys":     []string{"Ec2EbsVolumes"},
						"models":   []string{"AwsEc2EbsVolume"},
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
			"Ec2EbsVolumes.Name",
			"Ec2EbsVolumes.VolumeSize",
			"Ec2EbsVolumes.VolumeType",
		},
		"full_graph_fetch": map[string]any{
			"enabled": true,
		},
		"max_tier": 2,
	}
}

func (q *EC2InstanceQuery) MapResponse(node *OrcaAssetNode) (*AssetDetails, error) {
	d := node.Data

	// Extract attached EBS volumes from the nested Ec2EbsVolumes field.
	volumes := ExtractNestedObjects(d, "Ec2EbsVolumes")
	var disks []DiskInfo
	for _, vol := range volumes {
		disks = append(disks, DiskInfo{
			ID:         vol.AssetUniqueID,
			Name:       vol.Name,
			SizeGB:     ExtractFloat(vol.Data, "VolumeSize"),
			VolumeType: ExtractString(vol.Data, "VolumeType"),
		})
	}

	return &AssetDetails{
		AssetUniqueID: node.AssetUniqueID,
		Type:          node.Type,
		Name:          node.Name,
		State:         ExtractString(d, "State"),
		Disks:         disks,
	}, nil
}
