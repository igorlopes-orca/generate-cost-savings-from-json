package pricing

// EBSPricePerGB maps EBS volume types to their monthly cost per GB (us-east-1).
var EBSPricePerGB = map[string]float64{
	"gp3":      0.08,
	"gp2":      0.10,
	"io1":      0.125,
	"io2":      0.125,
	"st1":      0.045,
	"sc1":      0.015,
	"standard": 0.05,
}

// EBSSnapshotPricePerGB is the monthly cost per GB for EBS snapshots.
const EBSSnapshotPricePerGB = 0.05

// KMSKeyMonthly is the fixed monthly cost for an AWS KMS key.
const KMSKeyMonthly = 1.00
