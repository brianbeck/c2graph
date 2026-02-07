// =============================================================================
// Sentinel - Seed Known Solana Programs
// These are well-known programs on Solana with risk categorizations.
// =============================================================================

// -- System Programs --
MERGE (p:Program {program_id: "11111111111111111111111111111111"})
SET p.name = "System Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"})
SET p.name = "Token Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"})
SET p.name = "Token-2022 Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"})
SET p.name = "Associated Token Account Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "ComputeBudget111111111111111111111111111111"})
SET p.name = "Compute Budget Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "Vote111111111111111111111111111111111111111"})
SET p.name = "Vote Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "Stake11111111111111111111111111111111111111"})
SET p.name = "Stake Program", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"})
SET p.name = "Memo Program v2", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "Memo1UhkJBfCR6MNB5bUcqwaFT7q6FhMzfshwJpp6i5"})
SET p.name = "Memo Program v1", p.category = "System", p.risk_level = "safe";

MERGE (p:Program {program_id: "namesLPneVptA9Z5rqUDD9tMTWEJwofgaYwp8cawRkX"})
SET p.name = "Name Service Program", p.category = "System", p.risk_level = "safe";

// -- DEX / AMM Programs --
MERGE (p:Program {program_id: "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4"})
SET p.name = "Jupiter Aggregator v6", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "jupoNjAxXgZ4rjzxzPMP4oxduvQsQtZzyknqvzYNrNu"})
SET p.name = "Jupiter Limit Order", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"})
SET p.name = "Raydium AMM v4", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK"})
SET p.name = "Raydium CLMM", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"})
SET p.name = "Orca Whirlpools", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"})
SET p.name = "Orca Token Swap v2", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "SSwpkEEcbUqx4vtoEByFjSkhKdCT862DNVb52nZg1UZ"})
SET p.name = "Saber Stable Swap", p.category = "DEX", p.risk_level = "safe";

MERGE (p:Program {program_id: "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"})
SET p.name = "Meteora DLMM", p.category = "DEX", p.risk_level = "safe";

// -- Lending / Borrowing --
MERGE (p:Program {program_id: "MFv2hWf31Z9kbCa1snEPYctwafyhdvnV7FZnsebVacA"})
SET p.name = "Marginfi v2", p.category = "Lending", p.risk_level = "safe";

MERGE (p:Program {program_id: "So1endDq2YkqhipRh3WViPa8hFMiGBMQi8aDbpq1STLL"})
SET p.name = "Solend", p.category = "Lending", p.risk_level = "safe";

MERGE (p:Program {program_id: "KLend2g3cP87ber8pVKvFMeHZjDmbJzB4rBiVN7FhFHa"})
SET p.name = "Kamino Lending", p.category = "Lending", p.risk_level = "safe";

// -- NFT / Metaplex --
MERGE (p:Program {program_id: "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"})
SET p.name = "Metaplex Token Metadata", p.category = "NFT", p.risk_level = "safe";

MERGE (p:Program {program_id: "M2mx93ekt1fmXSVkTrUL9xVFHkmME8HTUi5Cyc5aF7K"})
SET p.name = "Magic Eden v2", p.category = "NFT Marketplace", p.risk_level = "safe";

// -- Bridges --
MERGE (p:Program {program_id: "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth"})
SET p.name = "Wormhole Core Bridge", p.category = "Bridge", p.risk_level = "medium";

MERGE (p:Program {program_id: "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb"})
SET p.name = "Wormhole Token Bridge", p.category = "Bridge", p.risk_level = "medium";

// -- Known High-Risk / Mixer Programs --
// These are flagged for enhanced monitoring. The list should be updated
// regularly based on threat intelligence feeds.

// Placeholder: Add known mixer and tumbler program IDs here as they are identified.
// Example:
// MERGE (p:Program {program_id: "KNOWN_MIXER_PROGRAM_ID_HERE"})
// SET p.name = "Known Mixer", p.category = "Mixer", p.risk_level = "high";

// -- Governance --
MERGE (p:Program {program_id: "GovER5Lthms3bLBqWub97yVrMmEogzX7xNjdXpPPCVZw"})
SET p.name = "SPL Governance", p.category = "Governance", p.risk_level = "safe";

MERGE (p:Program {program_id: "pytS9TjG1qyAZypk7n8rw8gfW9sUaqqYyMhJQ4E7JCQ"})
SET p.name = "Pyth Oracle", p.category = "Oracle", p.risk_level = "safe";

MERGE (p:Program {program_id: "SW1TCH7qEPTdLsDHRgPuMQjbQxKdH2aBStViMFnt64f"})
SET p.name = "Switchboard Oracle", p.category = "Oracle", p.risk_level = "safe";
