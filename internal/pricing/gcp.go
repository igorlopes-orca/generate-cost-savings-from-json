package pricing

// GCPDiskPricePerGB maps GCP persistent disk types to their monthly cost per GB (us-central1).
var GCPDiskPricePerGB = map[string]float64{
	"pd-standard":        0.04,
	"pd-balanced":        0.10,
	"pd-ssd":             0.17,
	"hyperdisk-balanced": 0.06,
}

// GCPSnapshotPricePerGB is the monthly cost per GB for GCP disk snapshots.
const GCPSnapshotPricePerGB = 0.026

// GCPKMSKeyVersionMonthly is the fixed monthly cost for a GCP KMS key version.
const GCPKMSKeyVersionMonthly = 0.06
