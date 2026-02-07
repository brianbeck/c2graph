.PHONY: dev-infra dev-api dev-worker dev-frontend stop-infra neo4j-constraints neo4j-seed build test clean

# --- Infrastructure ---

dev-infra:
	docker compose up -d
	@echo "Waiting for Neo4j to be healthy..."
	@until docker exec sentinel-neo4j cypher-shell -u neo4j -p sentinel-dev-password "RETURN 1" > /dev/null 2>&1; do sleep 2; done
	@echo "Neo4j is ready. Applying constraints..."
	$(MAKE) neo4j-constraints
	$(MAKE) neo4j-seed
	@echo "Infrastructure ready!"
	@echo "  Neo4j Browser: http://localhost:7474"
	@echo "  RabbitMQ Mgmt: http://localhost:15672"

stop-infra:
	docker compose down

neo4j-constraints:
	@cat scripts/neo4j-constraints.cypher | docker exec -i sentinel-neo4j cypher-shell -u neo4j -p sentinel-dev-password --format plain 2>/dev/null || true

neo4j-seed:
	@cat scripts/seed-programs.cypher | docker exec -i sentinel-neo4j cypher-shell -u neo4j -p sentinel-dev-password --format plain 2>/dev/null || true

# --- Development Servers ---

dev-api:
	cd api-gateway && go run ./cmd/server/

dev-worker:
	cd worker && go run ./cmd/worker/

dev-frontend:
	cd frontend && npm run dev

# --- Build ---

build: build-api build-worker build-frontend

build-api:
	cd api-gateway && go build -o sentinel-api ./cmd/server/

build-worker:
	cd worker && go build -o sentinel-worker ./cmd/worker/

build-frontend:
	cd frontend && npm run build

# --- Test ---

test: test-api test-worker

test-api:
	cd api-gateway && go test ./...

test-worker:
	cd worker && go test ./...

# --- Clean ---

clean:
	rm -f api-gateway/sentinel-api worker/sentinel-worker
	rm -rf frontend/dist
