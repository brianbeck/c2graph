// =============================================================================
// C2Graph - Neo4j Schema Constraints & Indexes
// Run this after Neo4j starts to set up the schema.
// =============================================================================

// -- Uniqueness Constraints --
CREATE CONSTRAINT wallet_address IF NOT EXISTS FOR (w:Wallet) REQUIRE w.address IS UNIQUE;
CREATE CONSTRAINT tx_signature IF NOT EXISTS FOR (t:Transaction) REQUIRE t.signature IS UNIQUE;
CREATE CONSTRAINT token_mint IF NOT EXISTS FOR (t:Token) REQUIRE t.mint_address IS UNIQUE;
CREATE CONSTRAINT program_id IF NOT EXISTS FOR (p:Program) REQUIRE p.program_id IS UNIQUE;
CREATE CONSTRAINT lookup_table_address IF NOT EXISTS FOR (a:AddressLookupTable) REQUIRE a.address IS UNIQUE;
CREATE CONSTRAINT scan_job_id IF NOT EXISTS FOR (j:ScanJob) REQUIRE j.job_id IS UNIQUE;

// -- Performance Indexes --
CREATE INDEX wallet_risk IF NOT EXISTS FOR (w:Wallet) ON (w.risk_score);
CREATE INDEX wallet_last_scanned IF NOT EXISTS FOR (w:Wallet) ON (w.last_scanned);
CREATE INDEX wallet_tags IF NOT EXISTS FOR (w:Wallet) ON (w.tags);
CREATE INDEX tx_block_time IF NOT EXISTS FOR (t:Transaction) ON (t.block_time);
CREATE INDEX tx_slot IF NOT EXISTS FOR (t:Transaction) ON (t.slot);
CREATE INDEX program_category IF NOT EXISTS FOR (p:Program) ON (p.category);
CREATE INDEX scan_job_status IF NOT EXISTS FOR (j:ScanJob) ON (j.status);
CREATE INDEX scan_job_root IF NOT EXISTS FOR (j:ScanJob) ON (j.root_address);
