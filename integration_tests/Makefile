build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

down:
	docker compose down -v

up: down build-image
	docker network create flagr_integration_tests || true
	docker compose up -d

SERVER_URLS = localhost:18001,localhost:18002,localhost:18003,localhost:18004,localhost:18005,localhost:18006

# Run Go integration tests against all 6 Docker Compose instances.
test: up
	FLAGR_SERVER_URLS="$(SERVER_URLS)" \
	go test -tags=integration -v -count=1 .

# Run Go integration benchmarks against all 6 Docker Compose instances.
bench: up
	@echo ""
	@echo "========================================="
	@echo "  INTEGRATION BENCHMARKS"
	@echo "========================================="
	FLAGR_SERVER_URLS="$(SERVER_URLS)" \
	go test -tags=integration -bench=. -benchmem -count=1 -run='^$$' .
	@echo "========================================="
	@echo "  END OF BENCHMARKS"
	@echo "========================================="
	@echo ""

# Run tests then benchmarks against the same servers.
test-and-bench: up
	FLAGR_SERVER_URLS="$(SERVER_URLS)" \
	go test -tags=integration -v -count=1 .
	@echo ""
	@echo "========================================="
	@echo "  INTEGRATION BENCHMARKS"
	@echo "========================================="
	FLAGR_SERVER_URLS="$(SERVER_URLS)" \
	go test -tags=integration -bench=. -benchmem -count=1 -run='^$$' .
	@echo "========================================="
	@echo "  END OF BENCHMARKS"
	@echo "========================================="
	@echo ""

retest: down test
