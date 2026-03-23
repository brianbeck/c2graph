// Graph data types matching the API response format for react-force-graph.

export interface GraphNode {
  id: string;
  type: "wallet" | "transaction";
  props: {
    type: string;
    address?: string;
    signature?: string;
    risk_score?: number;
    tags?: string[];
    lamports?: number;
    first_seen?: string;
    last_seen?: string;
    tx_count?: number;
    owner?: string;
    executable?: boolean;
    fee?: number;
    block_time?: string;
    slot?: number;
    version?: string;
    compute_units_consumed?: number;
    err?: string;
    risk_factors?: string[];
    bot_likelihood?: number;
    [key: string]: unknown;
  };
}

export interface GraphLink {
  source: string;
  target: string;
  type: string;
  props?: {
    amount_lamports?: number;
    direction?: "send" | "receive";
    amount?: string;
    ui_amount?: string;
    mint?: string;
    decimals?: number;
    [key: string]: unknown;
  };
}

export interface GraphResponse {
  nodes: GraphNode[];
  links: GraphLink[];
}

export interface ScanRequest {
  address: string;
  depth: number;
}

export interface ScanResponse {
  status: "ready" | "queued";
  job_id?: string;
  address: string;
}

export interface JobStatus {
  job_id: string;
  root_address: string;
  requested_depth: number;
  status: "queued" | "processing" | "complete" | "capped" | "failed";
  wallets_processed: number;
  wallets_cap: number;
  created_at: string;
  updated_at: string;
}
