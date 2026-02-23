package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrUnsupportedAssetType indicates that no query is registered for the given asset type.
// This is an expected gap (not all asset types are implemented yet), not a system failure.
var ErrUnsupportedAssetType = errors.New("unsupported asset type")

// OrcaClient implements AssetFetcher by querying the Orca serving-layer API.
type OrcaClient struct {
	baseURL         string
	token           string
	httpClient      *http.Client
	queries         map[string]AssetQuery
	snapshotQueries map[string]DiskSnapshotQuery
}

// NewOrcaClient creates a new OrcaClient with registered asset-type queries.
func NewOrcaClient(baseURL, token string) *OrcaClient {
	c := &OrcaClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		queries:         make(map[string]AssetQuery),
		snapshotQueries: make(map[string]DiskSnapshotQuery),
	}

	// Register known queries — AWS
	c.registerQuery(&EC2InstanceQuery{})
	c.registerQuery(&EBSVolumeQuery{})
	c.registerQuery(&EBSSnapshotQuery{})
	c.registerQuery(&KMSKeyQuery{})

	// Register known queries — GCP
	c.registerQuery(&GCPVMDiskQuery{})
	c.registerQuery(&GCPVMInstanceQuery{})
	c.registerQuery(&GCPKMSKeyQuery{})

	// Register disk-snapshot queries (two-pass enrichment)
	c.registerSnapshotQuery(&EBSVolumeSnapshotsQuery{})
	c.registerSnapshotQuery(&GCPVMDiskSnapshotsQuery{})

	return c
}

func (c *OrcaClient) registerQuery(q AssetQuery) {
	c.queries[q.AssetType()] = q
}

func (c *OrcaClient) registerSnapshotQuery(q DiskSnapshotQuery) {
	c.snapshotQueries[q.DiskAssetType()] = q
}

// FetchAsset dispatches to the correct query builder based on assetType,
// executes the query, and maps the response.
func (c *OrcaClient) FetchAsset(ctx context.Context, assetType, assetUniqueID string) (*AssetDetails, error) {
	query, ok := c.queries[assetType]
	if !ok {
		return nil, fmt.Errorf("%w %q: no query registered", ErrUnsupportedAssetType, assetType)
	}

	payload := query.BuildPayload(assetUniqueID)

	node, err := c.doQuery(ctx, payload, assetType, assetUniqueID)
	if err != nil {
		return nil, err
	}

	return query.MapResponse(node)
}

// FetchDiskSnapshots fetches all snapshots for a given disk by querying the disk
// with nested snapshot objects (two-pass enrichment).
func (c *OrcaClient) FetchDiskSnapshots(ctx context.Context, diskAssetType, diskUniqueID string) ([]SnapshotInfo, error) {
	query, ok := c.snapshotQueries[diskAssetType]
	if !ok {
		return nil, fmt.Errorf("%w %q: no snapshot query registered", ErrUnsupportedAssetType, diskAssetType)
	}

	payload := query.BuildPayload(diskUniqueID)

	node, err := c.doQuery(ctx, payload, diskAssetType, diskUniqueID)
	if err != nil {
		return nil, err
	}

	return query.MapResponse(node)
}

// doQuery handles the shared HTTP transport: marshalling, POST, auth, and response envelope parsing.
func (c *OrcaClient) doQuery(ctx context.Context, payload any, assetType, assetUniqueID string) (*OrcaAssetNode, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	url := c.baseURL + "/api/serving-layer/query"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "TOKEN "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from Orca API", resp.StatusCode)
	}

	var orcaResp orcaResponse
	if err := json.NewDecoder(resp.Body).Decode(&orcaResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if orcaResp.Status != "success" {
		return nil, fmt.Errorf("orca API returned status %q", orcaResp.Status)
	}

	if len(orcaResp.Data) == 0 {
		return nil, fmt.Errorf("no asset found for type=%s id=%s", assetType, assetUniqueID)
	}

	return &orcaResp.Data[0], nil
}

// orcaResponse is the top-level envelope returned by the serving-layer API.
type orcaResponse struct {
	Status string          `json:"status"`
	Data   []OrcaAssetNode `json:"data"`
}
