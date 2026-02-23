# cloud-savings

A CLI tool that analyzes [Orca Security](https://orca.security) cost optimization reports and calculates potential monthly USD savings from non-compliant cloud assets.

**v1 supports AWS + GCP controls.** Azure controls are planned for a future release.

## Installation

```bash
go install github.com/igorlopes-orca/generate-cost-savings-from-json@latest
```

## Usage

```bash
cloud-savings --file cost_optimization_report.json --api-token $ORCA_API_TOKEN
```

### Flags

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--file` | `-f` | Yes | Path to the Orca cost optimization JSON report |
| `--api-token` | `-t` | Yes | Orca API token for fetching asset details |
| `--api-url` | | No | Orca API base URL (default: `https://app.us.orcasecurity.io`) |
| `--log-level` | `-l` | No | Log level: `debug`, `info`, `warn`, `error` (default: `info`) |

## Example Output

```
Cloud Savings Report
====================

  Control                                          Assets   Monthly Savings
  ──────────────────────────────────────────────────────────────────────────────
  Stopped EC2 instances not removed                     7        $1,234.56
  Unattached EBS volumes                                1           $12.50
  Orphaned EBS snapshots                                2            $3.40
  Disabled AWS KMS keys                                 1            $1.00
  ══════════════════════════════════════════════════════════════════════════════
  TOTAL POTENTIAL MONTHLY SAVINGS                      11        $1,251.46
```

## Pricing

v1 uses **hardcoded pricing tables**. There is no dependency on any cloud pricing API. AWS prices (us-east-1) are in `internal/pricing/aws.go`, GCP prices (us-central1) are in `internal/pricing/gcp.go`.

### EBS Price Table

| Volume Type | $/GB/month |
|-------------|-----------|
| gp3 | 0.08 |
| gp2 | 0.10 |
| io1 | 0.125 |
| io2 | 0.125 |
| st1 | 0.045 |
| sc1 | 0.015 |
| standard | 0.05 |

### Snapshot Pricing

- EBS snapshots: **$0.05/GB/month**

### KMS Pricing

- Disabled CMK: **$1.00/key/month** (fixed)

### GCP Disk Price Table

| Disk Type | $/GB/month |
|-----------|-----------|
| pd-standard | 0.04 |
| pd-balanced | 0.10 |
| pd-ssd | 0.17 |
| hyperdisk-balanced | 0.06 |

### GCP Snapshot Pricing

- Disk snapshots: **$0.026/GB/month**

### GCP KMS Pricing

- Disabled key version: **$0.06/key/month** (fixed)

---

## Supported Controls (AWS)

| # | Control Name (exact) | Asset Type | Formula |
|---|---|---|---|
| 1 | `Stopped Ec2 instance not removed` | `AwsEc2Instance` | `sum(vol.size_gb × ebs_price[vol.type])` for each attached volume |
| 2 | `Ensure EBS volume is attached to an EC2 instance` | `AwsEc2EbsVolume` | `size_gb × ebs_price[type]` |
| 3 | `Low disk space utilization in your AWS EC2 EBS volume` | `AwsEc2EbsVolume` | `(size_gb - used_gb) × ebs_price[type]` |
| 4 | `EBS snapshot's originating volume no longer exists` | `AwsEc2EbsSnapshot` | `size_gb × $0.05` |
| 5 | `Ensure EC2 instances with EBS volumes have only one updated snapshot` | `AwsEc2Instance` | `sum((disk.snapshot_count - 1) × avg_snapshot_size × $0.05)` per disk |
| 6 | `Ensure EC2 instances with EBS volumes with snapshots created less than 90 days ago` | `AwsEc2Instance` | `sum(snap.size_gb × $0.05)` for snapshots older than 90 days |
| 7 | `Identify and remove any disabled AWS Customer Master Keys (CMK)` | `AwsKmsKey` | `$1.00` fixed |

### Supported Orca API Asset Types

Each asset type has a dedicated query builder that fetches enriched details via the Orca serving-layer API (filtered by `AssetUniqueId`):

| Orca Model | Query File | Used By Controls | Key Fields |
|---|---|---|---|
| `AwsEc2Instance` | `ec2_instance_query.go` | #1, #5, #6 | State, nested Ec2EbsVolumes (Name, VolumeSize, VolumeType) |
| `AwsEc2EbsVolume` | `ebs_volume_query.go` | #2, #3 | VolumeSize, VolumeType, UsedDiskSize, Region, State |
| `AwsEc2EbsSnapshot` | `ebs_snapshot_query.go` | #4 | VolumeSize, UsedDiskSize, Region, State |
| `AwsKmsKey` | `kms_key_query.go` | #7 | State, Region |

---

## Supported Controls (GCP)

| # | Control Name (exact) | Asset Type | Formula |
|---|---|---|---|
| 1 | `Ensure GCP disk is attached to a virtual machine` | `GcpVmDisk` | `size_gb × disk_price[type]` |
| 2 | `Low disk space utilization in your GCP Disk` | `GcpVmDisk` | `(size_gb - used_gb) × disk_price[type]` |
| 3 | `Ensure gcp VM's disks have only one snapshot` | `GcpVmInstance` | `sum((disk.snapshot_count - 1) × avg_snapshot_size × $0.026)` per disk |
| 4 | `GCP VM's disks with snapshots created more than 90 days ago` | `GcpVmInstance` | `sum(snap.size_gb × $0.026)` for snapshots older than 90 days |
| 5 | `Identify and remove any disabled GCP KMS primary key versions` | `GcpKmsKey` | `$0.06` fixed |

### Supported Orca API Asset Types (GCP)

| Orca Model | Query File | Used By Controls | Key Fields |
|---|---|---|---|
| `GcpVmDisk` | `gcp_vm_disk_query.go` | #1, #2 | SizeGb, VolumeType, UsedDiskSize, Region, State |
| `GcpVmInstance` | `gcp_vm_instance_query.go` | #3, #4 | State, nested InstanceDisks (Name) |
| `GcpKmsKey` | `gcp_kms_key_query.go` | #5 | State, Region |

### Not Supported

| Control Name | Reason |
|---|---|
| `ELB with expired ACM certificate` | Security finding — no cost savings |
| `ELB with expired IAM certificate` | Security finding — no cost savings |
| `Ensure Ali Ecs disk is attached to a virtual machine` | AliCloud pricing not available |
| `Identify and remove any disabled AliCloud CMK (customer master keys)` | AliCloud pricing not available |

### Planned (Future)

Azure controls will be added in a future release.

---

## Development

```bash
# Build
go build -o cloud-savings .

# Run tests
go test ./...

# Run with debug logging
./cloud-savings -f report.json -t $ORCA_API_TOKEN -l debug
```

## License

Internal tool — Orca Security.
