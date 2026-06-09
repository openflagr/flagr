build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

down:
	docker compose down -v

up:
	docker network create flagr_integration_tests || true
	docker compose up -d

# Run Go integration tests against all 6 Docker Compose instances
test: down build-image up
	@errors=0; \
	for instance in flagr_with_sqlite flagr_with_mysql flagr_with_mysql8 \
	                flagr_with_postgres9 flagr_with_postgres13 \
	                checkr_flagr_with_sqlite; do \
		echo "=== $$instance ==="; \
		if docker compose exec -T integration-runner \
			sh -c "FLAGR_SERVER_URL=http://$$instance:18000 go test -tags=integration -v -count=1 ./integration_tests/"; then \
			:; \
		else \
			echo "FAILED: $$instance"; \
			errors=$$((errors+1)); \
		fi; \
	done; \
	[ $$errors -eq 0 ]

# Single-instance test (used by CI matrix)
test-instance: up
	docker compose exec -T integration-runner \
		sh -c "FLAGR_SERVER_URL=http://$(INSTANCE):18000 go test -tags=integration -v -count=1 ./integration_tests/"

INSTANCE ?= flagr_with_sqlite

retest: down test
