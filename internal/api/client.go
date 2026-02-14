package api

import (
	"context"
	"fmt"
	"time"
)

// AssetFetcher retrieves detailed asset information from the Orca API.
type AssetFetcher interface {
	FetchAsset(ctx context.Context, assetType, assetUniqueID string) (*AssetDetails, error)
}

// AssetDetails contains the enriched data returned by the Orca API for a single asset.
type AssetDetails struct {
	AssetUniqueID string `json:"asset_unique_id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Region        string `json:"region"`
	State         string `json:"state"`

	// Storage fields (EBS volumes)
	SizeGB     float64 `json:"size_gb"`
	UsedGB     float64 `json:"used_gb"`
	VolumeType string  `json:"volume_type"` // gp2, gp3, io1, io2, st1, sc1, standard

	// Attached disks (for VM-level controls like stopped EC2)
	Disks []DiskInfo `json:"disks,omitempty"`

	// Snapshots (for snapshot-level controls)
	Snapshots []SnapshotInfo `json:"snapshots,omitempty"`
}

// DiskInfo describes a disk attached to a VM.
type DiskInfo struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	SizeGB     float64        `json:"size_gb"`
	VolumeType string         `json:"volume_type"` // gp2, gp3, io1, etc.
	Snapshots  []SnapshotInfo `json:"snapshots,omitempty"`
}

// SnapshotInfo describes a single snapshot.
type SnapshotInfo struct {
	ID        string    `json:"id"`
	SizeGB    float64   `json:"size_gb"`
	CreatedAt time.Time `json:"created_at"`
}

// StubFetcher is a placeholder AssetFetcher that always returns an error.
type StubFetcher struct{}

func (s *StubFetcher) FetchAsset(_ context.Context, assetType, assetUniqueID string) (*AssetDetails, error) {
	return nil, fmt.Errorf("stub fetcher: Orca API not yet configured (type=%s asset=%s)", assetType, assetUniqueID)
}
