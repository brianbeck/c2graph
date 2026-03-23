#!/bin/sh
# Wait for Neo4j to be ready, then apply constraints and seed data.
# Used by the neo4j-init container in docker-compose.

set -e

NEO4J_HOST="${NEO4J_HOST:-neo4j}"
NEO4J_PORT="${NEO4J_PORT:-7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-c2graph-dev-password}"

echo "Waiting for Neo4j at ${NEO4J_HOST}:${NEO4J_PORT}..."

# Poll until cypher-shell can connect
until cypher-shell -a "bolt://${NEO4J_HOST}:${NEO4J_PORT}" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" "RETURN 1" >/dev/null 2>&1; do
  sleep 2
done

echo "Neo4j is ready. Applying constraints..."
cypher-shell -a "bolt://${NEO4J_HOST}:${NEO4J_PORT}" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" --format plain < /scripts/neo4j-constraints.cypher || true

echo "Seeding programs..."
cypher-shell -a "bolt://${NEO4J_HOST}:${NEO4J_PORT}" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" --format plain < /scripts/seed-programs.cypher || true

echo "Neo4j initialization complete."
