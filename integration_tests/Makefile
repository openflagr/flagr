build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

down:
	docker compose down -v

up:
	docker network create flagr_integration_tests || true
	docker compose up -d

# Run Go integration tests against all 6 Docker Compose instances.
# Tests run on the host (not inside Docker) and connect via port mappings.
test: down build-image up
	FLAGR_SERVER_URLS="localhost:18001,localhost:18002,localhost:18003,localhost:18004,localhost:18005,localhost:18006" \
	go test -tags=integration -v -count=1 .

retest: down test
