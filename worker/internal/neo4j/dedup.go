package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// DedupChecker provides methods to avoid re-fetching data already in Neo4j.
type DedupChecker struct {
	driver         neo4j.DriverWithContext
	freshnessHours int
}

// NewDedupChecker creates a new dedup checker.
func NewDedupChecker(driver neo4j.DriverWithContext, freshnessHours int) *DedupChecker {
	return &DedupChecker{
		driver:         driver,
		freshnessHours: freshnessHours,
	}
}

// IsWalletFresh checks if a wallet was recently scanned at the requested depth.
// Returns true if the wallet data is fresh and does not need re-scanning.
func (d *DedupChecker) IsWalletFresh(ctx context.Context, address string, depth int) (bool, error) {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (w:Wallet {address: $address})
			WHERE w.last_scanned IS NOT NULL
			  AND w.last_scanned > datetime() - duration({hours: $freshness_hours})
			  AND COALESCE(w.scan_depth, 0) >= $depth
			RETURN count(w) > 0 AS is_fresh
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address":         address,
			"freshness_hours": d.freshnessHours,
			"depth":           depth,
		})
		if err != nil {
			return false, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("is_fresh")
			if b, ok := val.(bool); ok {
				return b, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return false, fmt.Errorf("check wallet freshness: %w", err)
	}

	return result.(bool), nil
}

// FilterNewSignatures takes a list of transaction signatures and returns only
// those that are NOT already in the database. This is the core of Layer 2 dedup.
func (d *DedupChecker) FilterNewSignatures(ctx context.Context, signatures []string) ([]string, error) {
	if len(signatures) == 0 {
		return nil, nil
	}

	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Find which of the given signatures already exist
		query := `
			UNWIND $signatures AS sig
			MATCH (t:Transaction {signature: sig})
			RETURN collect(t.signature) AS existing
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"signatures": signatures,
		})
		if err != nil {
			return nil, err
		}

		existingSet := make(map[string]bool)
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("existing")
			if existing, ok := val.([]interface{}); ok {
				for _, s := range existing {
					if str, ok := s.(string); ok {
						existingSet[str] = true
					}
				}
			}
		}

		// Return only signatures NOT in the database
		var newSigs []string
		for _, sig := range signatures {
			if !existingSet[sig] {
				newSigs = append(newSigs, sig)
			}
		}
		return newSigs, nil
	})
	if err != nil {
		return nil, fmt.Errorf("filter new signatures: %w", err)
	}

	return result.([]string), nil
}

// MarkWalletScanned updates the wallet's last_scanned timestamp and scan_depth.
func (d *DedupChecker) MarkWalletScanned(ctx context.Context, address string, depth int) error {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MERGE (w:Wallet {address: $address})
			SET w.last_scanned = datetime(),
			    w.scan_depth = $depth
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
			"depth":   depth,
		})
		return nil, err
	})
	return err
}

// GetDiscoveredWallets finds all wallets that are connected to the given wallet
// through transactions (i.e., wallets that received SOL/tokens from or sent to
// wallets the target wallet transacted with). These are candidates for depth+1 traversal.
func (d *DedupChecker) GetDiscoveredWallets(ctx context.Context, address string) ([]string, error) {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Find wallets connected through transactions initiated by or involving this wallet
		query := `
			MATCH (w:Wallet {address: $address})-[:INITIATED]->(t:Transaction)-[:TRANSFERRED_SOL|TRANSFERRED_TOKEN]->(other:Wallet)
			WHERE other.address <> $address
			RETURN DISTINCT other.address AS address
			UNION
			MATCH (other:Wallet)-[:INITIATED]->(t:Transaction)-[:TRANSFERRED_SOL|TRANSFERRED_TOKEN]->(w:Wallet {address: $address})
			WHERE other.address <> $address
			RETURN DISTINCT other.address AS address
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return nil, err
		}

		var wallets []string
		for rec.Next(ctx) {
			val, _ := rec.Record().Get("address")
			if addr, ok := val.(string); ok {
				wallets = append(wallets, addr)
			}
		}
		return wallets, nil
	})
	if err != nil {
		return nil, fmt.Errorf("get discovered wallets: %w", err)
	}

	return result.([]string), nil
}

// UpdateScanJobProgress updates the ScanJob node with current progress.
func (d *DedupChecker) UpdateScanJobProgress(ctx context.Context, jobID string, walletsProcessed int, status string) error {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MERGE (j:ScanJob {job_id: $job_id})
			SET j.wallets_processed = $wallets_processed,
			    j.status = $status,
			    j.updated_at = datetime()
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"job_id":            jobID,
			"wallets_processed": walletsProcessed,
			"status":            status,
		})
		return nil, err
	})
	return err
}

// SetScanJobExpected sets the total number of wallets expected for a job.
// Called after the root wallet enqueues its children.
func (d *DedupChecker) SetScanJobExpected(ctx context.Context, jobID string, walletsExpected int) error {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (j:ScanJob {job_id: $job_id})
			SET j.wallets_expected = $wallets_expected,
			    j.updated_at = datetime()
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"job_id":           jobID,
			"wallets_expected": walletsExpected,
		})
		return nil, err
	})
	return err
}

// CreateScanJob creates a new ScanJob node in Neo4j.
func (d *DedupChecker) CreateScanJob(ctx context.Context, jobID, rootAddress string, depth, walletsCap int) error {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (j:ScanJob {
				job_id: $job_id,
				root_address: $root_address,
				requested_depth: $depth,
				status: 'queued',
				wallets_processed: 0,
				wallets_cap: $wallets_cap,
				created_at: datetime(),
				updated_at: datetime()
			})
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"job_id":       jobID,
			"root_address": rootAddress,
			"depth":        depth,
			"wallets_cap":  walletsCap,
		})
		return nil, err
	})
	return err
}

// GetScanJob retrieves a ScanJob by ID.
func (d *DedupChecker) GetScanJob(ctx context.Context, jobID string) (*ScanJobInfo, error) {
	session := d.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (j:ScanJob {job_id: $job_id})
			RETURN j.job_id AS job_id,
			       j.root_address AS root_address,
			       j.requested_depth AS requested_depth,
			       j.status AS status,
			       j.wallets_processed AS wallets_processed,
			       j.wallets_cap AS wallets_cap,
			       j.created_at AS created_at,
			       j.updated_at AS updated_at
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"job_id": jobID,
		})
		if err != nil {
			return nil, err
		}
		if rec.Next(ctx) {
			record := rec.Record()
			info := &ScanJobInfo{}
			if v, ok := record.Get("job_id"); ok && v != nil {
				info.JobID, _ = v.(string)
			}
			if v, ok := record.Get("root_address"); ok && v != nil {
				info.RootAddress, _ = v.(string)
			}
			if v, ok := record.Get("requested_depth"); ok && v != nil {
				info.RequestedDepth, _ = v.(int64)
			}
			if v, ok := record.Get("status"); ok && v != nil {
				info.Status, _ = v.(string)
			}
			if v, ok := record.Get("wallets_processed"); ok && v != nil {
				info.WalletsProcessed, _ = v.(int64)
			}
			if v, ok := record.Get("wallets_cap"); ok && v != nil {
				info.WalletsCap, _ = v.(int64)
			}
			if v, ok := record.Get("created_at"); ok && v != nil {
				// Neo4j returns time as neo4j.Time or similar
				if t, ok := v.(time.Time); ok {
					info.CreatedAt = t
				}
			}
			return info, nil
		}
		return nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("get scan job: %w", err)
	}
	if result == nil {
		return nil, nil
	}
	return result.(*ScanJobInfo), nil
}

// ScanJobInfo represents the state of a scan job.
type ScanJobInfo struct {
	JobID            string
	RootAddress      string
	RequestedDepth   int64
	Status           string
	WalletsProcessed int64
	WalletsCap       int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
