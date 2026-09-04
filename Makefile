# ==============================================================================
# Lyostar - Ultra-lightweight, single-binary, self-hosted ebook server & reader
# ==============================================================================

BINARY_NAME   := lyostar
CMD_PATH      := ./cmd/lyostar
FRONTEND_DIR  := frontend
LDFLAGS       := -s -w

.PHONY: all build build-frontend build-backend run dev-backend dev-frontend test test-short clean sqlc help

# Default target
all: build

## help: Show available commands
help:
	@echo "Available commands in Lyostar:"
	@echo ""
	@echo "  make (or make build)  Build single binary embedding both BE & FE"
	@echo "  make run              Run the application (both BE & FE served on http://localhost:8080)"
	@echo "  make dev              Run directly from source (both BE & FE) without building binary"
	@echo "  make dev-live         Run Go backend + Vite HMR frontend simultaneously for UI coding"
	@echo "  make test             Run all unit tests"
	@echo "  make clean            Remove compiled binary"
	@echo ""

## build: Build frontend and then compile Go binary with embedded assets
build: build-frontend build-backend

## run: Run the built single-binary (serves both BE & FE at http://localhost:8080)
run:
	@if [ ! -f $(BINARY_NAME) ]; then $(MAKE) build; fi
	./$(BINARY_NAME)

## dev: Run both BE and embedded FE directly via go run (1 command)
dev:
	go run $(CMD_PATH)

## dev-live: Run backend + Vite HMR dev server concurrently (for active UI development)
dev-live:
	@echo "==> Starting backend and frontend live-reload concurrently..."
	@bash -c "trap 'kill 0' SIGINT SIGTERM EXIT; go run $(CMD_PATH) & (cd $(FRONTEND_DIR) && npm run dev)"

## dev-frontend: Start Vite dev server for frontend
dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

## test: Run all Go unit tests with verbose output
test:
	go test -v ./...

## test-short: Run all Go unit tests concisely
test-short:
	go test ./...

## clean: Remove compiled binary
clean:
	@echo "==> Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

## sqlc: Regenerate SQL code
sqlc:
	sqlc generate
