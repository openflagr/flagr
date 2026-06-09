build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

down:
	docker compose down -v

up:
	docker network create flagr_integration_tests || true
	docker compose up -d

# Run Go integration tests against all 6 Docker Compose instances.
# The TestMain loop in integration_server.go iterates over FLAGR_SERVER_URLS
# and runs the full suite against each one.
test: down build-image up
	docker compose exec -T integration-runner \
		env FLAGR_SERVER_URLS="flagr_with_sqlite:18000,flagr_with_mysql:18000,flagr_with_mysql8:18000,flagr_with_postgres9:18000,flagr_with_postgres13:18000,checkr_flagr_with_sqlite:18000" \
		go test -tags=integration -v -count=1 .

retest: down test
