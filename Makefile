.PHONY: all build build-backend build-frontend test test-backend test-frontend clean lint dev dev-db dev-api dev-worker setup-mac

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
setup-mac:
	@echo "Installing PostgreSQL via Homebrew..."
	brew install postgresql@15
	@echo "Starting PostgreSQL service..."
	brew services start postgresql@15
	@echo "Creating database 'image_play'..."
	createdb image_play || echo "Database may already exist"
	@echo "Running migrations..."
	psql postgres://postgres:postgres@localhost:5432/image_play -f $(BACKEND_DIR)/migrations/0001_init.sql
	psql postgres://postgres:postgres@localhost:5432/image_play -f $(BACKEND_DIR)/migrations/0002_scene_templates.sql
	psql postgres://postgres:postgres@localhost:5432/image_play -f $(INFRA_DIR)/sql/seed_scene_templates.sql
	@echo "Setup complete. PostgreSQL is running on localhost:5432"

dev-db:
	@echo "Trying to start PostgreSQL..."
	@which docker >/dev/null 2>&1 && cd $(INFRA_DIR) && docker compose up -d postgres && echo "PostgreSQL started via Docker." && exit 0; \
	which pg_ctl >/dev/null 2>&1 && pg_ctl -D $(shell brew --prefix)/var/postgresql@15 start 2>/dev/null && echo "PostgreSQL started via pg_ctl." && exit 0; \
	which brew >/dev/null 2>&1 && brew services start postgresql@15 2>/dev/null && echo "PostgreSQL started via brew services." && exit 0; \
	echo "ERROR: Could not start PostgreSQL."; \
	echo "  - If you have Docker, start Docker Desktop and try again."; \
	echo "  - Otherwise, run 'make setup-mac' to install PostgreSQL via Homebrew."; \
	exit 1

dev-api:
	@echo "Starting API server..."
	cd $(BACKEND_DIR) && JWT_SECRET=dev-secret DATABASE_URL=postgres://postgres:postgres@localhost:5432/image_play?sslmode=disable go run ./cmd/api

dev-worker:
	@echo "Starting worker..."
	cd $(BACKEND_DIR) && DATABASE_URL=postgres://postgres:postgres@localhost:5432/image_play?sslmode=disable go run ./cmd/worker

dev-frontend:
	@echo "Building frontend for WeChat MP (dev)..."
	cd $(FRONTEND_DIR) && npm run build -- --platform mp-weixin
	@echo "Import $(FRONTEND_DIR)/dist/build/mp-weixin into WeChat Developer Tools"

$(DIST_DIR):
	mkdir -p $(DIST_DIR)
