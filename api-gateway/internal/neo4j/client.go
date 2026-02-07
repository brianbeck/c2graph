package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Client wraps Neo4j operations for the API gateway.
type Client struct {
	driver         neo4j.DriverWithContext
	freshnessHours int
}

// NewClient creates a new Neo4j client.
func NewClient(driver neo4j.DriverWithContext, freshnessHours int) *Client {
	return &Client{driver: driver, freshnessHours: freshnessHours}
}

// GraphNode represents a node in the graph response.
type GraphNode struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Props     map[string]interface{} `json:"props"`
}

// GraphLink represents an edge in the graph response.
type GraphLink struct {
	Source string                 `json:"source"`
	Target string                 `json:"target"`
	Type   string                 `json:"type"`
	Props  map[string]interface{} `json:"props,omitempty"`
}

// GraphResponse is the format expected by react-force-graph.
type GraphResponse struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// IsWalletFresh checks if a wallet was recently scanned at the given depth.
func (c *Client) IsWalletFresh(ctx context.Context, address string, depth int) (bool, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
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
			"freshness_hours": c.freshnessHours,
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
		return false, err
	}
	return result.(bool), nil
}

// GetGraph fetches the graph around a wallet address up to the given depth.
func (c *Client) GetGraph(ctx context.Context, address string, depth int) (*GraphResponse, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Variable-length path query. Each "hop" in the user's mental model is
		// Wallet -> Transaction -> Wallet, which is 2 relationships in the graph.
		maxRels := depth * 2
		query := fmt.Sprintf(`
			MATCH path = (root:Wallet {address: $address})-[*1..%d]-(n)
			WHERE n:Wallet OR n:Transaction
			UNWIND relationships(path) AS rel
			WITH DISTINCT rel,
			     startNode(rel) AS src,
			     endNode(rel) AS tgt
			RETURN
				CASE WHEN src:Wallet THEN 'wallet'
				     WHEN src:Transaction THEN 'transaction'
				     ELSE labels(src)[0] END AS src_type,
				CASE WHEN tgt:Wallet THEN 'wallet'
				     WHEN tgt:Transaction THEN 'transaction'
				     ELSE labels(tgt)[0] END AS tgt_type,
				CASE WHEN src:Wallet THEN src.address
				     WHEN src:Transaction THEN src.signature
				     ELSE toString(id(src)) END AS src_id,
				CASE WHEN tgt:Wallet THEN tgt.address
				     WHEN tgt:Transaction THEN tgt.signature
				     ELSE toString(id(tgt)) END AS tgt_id,
				type(rel) AS rel_type,
				properties(rel) AS rel_props,
				properties(src) AS src_props,
				properties(tgt) AS tgt_props
		`, maxRels)

		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"address": address,
		})
		if err != nil {
			return nil, err
		}

		nodeMap := make(map[string]GraphNode)
		var links []GraphLink

		for rec.Next(ctx) {
			record := rec.Record()

			srcID, _ := record.Get("src_id")
			tgtID, _ := record.Get("tgt_id")
			srcType, _ := record.Get("src_type")
			tgtType, _ := record.Get("tgt_type")
			relType, _ := record.Get("rel_type")
			relProps, _ := record.Get("rel_props")
			srcProps, _ := record.Get("src_props")
			tgtProps, _ := record.Get("tgt_props")

			srcIDStr := fmt.Sprintf("%v", srcID)
			tgtIDStr := fmt.Sprintf("%v", tgtID)

			// Add source node if not seen
			if _, ok := nodeMap[srcIDStr]; !ok {
				props := toMapStringInterface(srcProps)
				props["type"] = fmt.Sprintf("%v", srcType)
				nodeMap[srcIDStr] = GraphNode{
					ID:    srcIDStr,
					Type:  fmt.Sprintf("%v", srcType),
					Props: props,
				}
			}

			// Add target node if not seen
			if _, ok := nodeMap[tgtIDStr]; !ok {
				props := toMapStringInterface(tgtProps)
				props["type"] = fmt.Sprintf("%v", tgtType)
				nodeMap[tgtIDStr] = GraphNode{
					ID:    tgtIDStr,
					Type:  fmt.Sprintf("%v", tgtType),
					Props: props,
				}
			}

			// Add link
			links = append(links, GraphLink{
				Source: srcIDStr,
				Target: tgtIDStr,
				Type:   fmt.Sprintf("%v", relType),
				Props:  toMapStringInterface(relProps),
			})
		}

		nodes := make([]GraphNode, 0, len(nodeMap))
		for _, n := range nodeMap {
			nodes = append(nodes, n)
		}

		return &GraphResponse{Nodes: nodes, Links: links}, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*GraphResponse), nil
}

// GetScanJob retrieves a ScanJob by ID.
func (c *Client) GetScanJob(ctx context.Context, jobID string) (map[string]interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (j:ScanJob {job_id: $job_id})
			RETURN properties(j) AS props
		`
		rec, err := tx.Run(ctx, query, map[string]interface{}{
			"job_id": jobID,
		})
		if err != nil {
			return nil, err
		}
		if rec.Next(ctx) {
			val, _ := rec.Record().Get("props")
			return toMapStringInterface(val), nil
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(map[string]interface{}), nil
}

// CreateScanJob creates a new ScanJob node.
func (c *Client) CreateScanJob(ctx context.Context, jobID, address string, depth, walletsCap int) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (j:ScanJob {
				job_id: $job_id,
				root_address: $address,
				requested_depth: $depth,
				status: 'queued',
				wallets_processed: 0,
				wallets_cap: $wallets_cap,
				created_at: datetime(),
				updated_at: datetime()
			})
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"job_id":      jobID,
			"address":     address,
			"depth":       depth,
			"wallets_cap": walletsCap,
		})
		return nil, err
	})
	return err
}

func toMapStringInterface(val interface{}) map[string]interface{} {
	if val == nil {
		return map[string]interface{}{}
	}
	if m, ok := val.(map[string]interface{}); ok {
		// Convert any neo4j time types to strings
		result := make(map[string]interface{}, len(m))
		for k, v := range m {
			switch t := v.(type) {
			case time.Time:
				result[k] = t.Format(time.RFC3339)
			default:
				result[k] = v
			}
		}
		return result
	}
	return map[string]interface{}{}
}
