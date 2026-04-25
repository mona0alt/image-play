.PHONY: all build build-backend build-frontend test test-backend test-frontend clean lint dev dev-db dev-api dev-worker

DIST_DIR := dist
BACKEND_DIR := backend
FRONTEND_DIR := frontend
INFRA_DIR := infra

all: build

build: build-backend build-frontend
	@echo "All builds complete. Binaries in $(DIST_DIR)/"

build-backend: $(DIST_DIR)
	@echo "Building backend API server..."
	cd $(BACKEND_DIR) && go build -o ../$(DIST_DIR)/api ./cmd/api
	@echo "Building backend worker..."
	cd $(BACKEND_DIR) && go build -o ../$(DIST_DIR)/worker ./cmd/worker
	@echo "Backend binaries built: $(DIST_DIR)/api, $(DIST_DIR)/worker"

build-frontend:
	@echo "Building frontend for WeChat Mini Program..."
	cd $(FRONTEND_DIR) && npm install --legacy-peer-deps && npm run build -- --platform mp-weixin
	@echo "Frontend build complete."
	@echo "Open WeChat Developer Tools and import: $(FRONTEND_DIR)/dist/build/mp-weixin"

test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	cd $(BACKEND_DIR) && go test ./...

test-frontend:
	@echo "Running frontend tests..."
	cd $(FRONTEND_DIR) && npm test -- --run
	@echo "Running frontend type check..."
	cd $(FRONTEND_DIR) && npm run type-check

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(DIST_DIR)
	cd $(FRONTEND_DIR) && rm -rf dist node_modules
	cd $(BACKEND_DIR) && go clean -cache

lint: lint-backend lint-frontend

lint-backend:
	cd $(BACKEND_DIR) && go vet ./...

lint-frontend:
	cd $(FRONTEND_DIR) && npx vue-tsc --noEmit

# Development helpers
dev-db:
	@echo "Starting PostgreSQL..."
	cd $(INFRA_DIR) && docker compose up -d postgres

dev-api:
	@echo "Starting API server..."
	JWT_SECRET=dev-secret DATABASE_URL=postgres://postgres:postgres@localhost:5432/image_play?sslmode=disable cd $(BACKEND_DIR) && go run ./cmd/api

dev-worker:
	@echo "Starting worker..."
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/image_play?sslmode=disable cd $(BACKEND_DIR) && go run ./cmd/worker

dev-frontend:
	@echo "Building frontend for WeChat MP (dev)..."
	cd $(FRONTEND_DIR) && npm run build -- --platform mp-weixin
	@echo "Import $(FRONTEND_DIR)/dist/build/mp-weixin into WeChat Developer Tools"

$(DIST_DIR):
	mkdir -p $(DIST_DIR)
