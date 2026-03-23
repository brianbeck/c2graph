.PHONY: dev-infra dev-api dev-worker dev-frontend stop-infra neo4j-constraints neo4j-seed build test clean up down

# --- Full Stack (Docker Compose) ---

up:
	@docker rm -f c2graph-neo4j c2graph-rabbitmq c2graph-neo4j-init c2graph-api c2graph-worker c2graph-frontend 2>/dev/null || true
	docker compose up --build -d
	@echo "C2Graph is starting..."
	@echo "  Frontend:      http://localhost:5173"
	@echo "  API Gateway:   http://localhost:8080/health"
	@echo "  Neo4j Browser: http://localhost:7474"
	@echo "  RabbitMQ Mgmt: http://localhost:15672"

down:
	docker compose down
	@docker rm -f c2graph-neo4j c2graph-rabbitmq c2graph-neo4j-init c2graph-api c2graph-worker c2graph-frontend 2>/dev/null || true

# --- Infrastructure Only (for local dev) ---

dev-infra:
	docker compose up -d neo4j rabbitmq neo4j-init
	@echo "Waiting for Neo4j init to complete..."
	@docker compose wait neo4j-init 2>/dev/null || until [ "$$(docker inspect -f '{{.State.Status}}' c2graph-neo4j-init 2>/dev/null)" = "exited" ]; do sleep 2; done
	@echo "Infrastructure ready!"
	@echo "  Neo4j Browser: http://localhost:7474"
	@echo "  RabbitMQ Mgmt: http://localhost:15672"

stop-infra:
	docker compose down

neo4j-constraints:
	@cat scripts/neo4j-constraints.cypher | docker exec -i c2graph-neo4j cypher-shell -u neo4j -p c2graph-dev-password --format plain 2>/dev/null || true

neo4j-seed:
	@cat scripts/seed-programs.cypher | docker exec -i c2graph-neo4j cypher-shell -u neo4j -p c2graph-dev-password --format plain 2>/dev/null || true

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
	cd api-gateway && go build -o c2graph-api ./cmd/server/

build-worker:
	cd worker && go build -o c2graph-worker ./cmd/worker/

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
	rm -f api-gateway/c2graph-api worker/c2graph-worker
	rm -rf frontend/dist
