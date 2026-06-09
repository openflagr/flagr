build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

down:
	docker compose down -v

up:
	docker network create flagr_integration_tests || true
	docker compose up -d

# Run Go integration tests against all 6 Docker Compose instances
test: down build-image up
	@total=0; pass=0; rm -f integration.test.log; \
	for instance in flagr_with_sqlite flagr_with_mysql flagr_with_mysql8 \
	                flagr_with_postgres9 flagr_with_postgres13 \
	                checkr_flagr_with_sqlite; do \
		echo "=== $$instance ===" | tee -a integration.test.log; \
		total=$$((total+1)); \
		if docker compose exec -T -e FLAGR_SERVER_URL=http://$$instance:18000 integration-runner \
			go test -tags=integration -v -count=1 ./integration_tests/ >> integration.test.log 2>&1; then \
			pass=$$((pass+1)); \
		else \
			echo "FAILED: $$instance" | tee -a integration.test.log; \
		fi; \
	done; \
	echo "=== $$pass/$$total instances passed ===" | tee -a integration.test.log; \
	[ $$pass -eq $$total ]

# Single-instance test (used by CI matrix)
test-instance: up
	@mkdir -p /tmp/flagr-integration && rm -f /tmp/flagr-integration/integration.test.log
	docker compose exec -T -e FLAGR_SERVER_URL=http://$(INSTANCE):18000 integration-runner \
		go test -tags=integration -v -count=1 ./integration_tests/ 2>&1 | tee /tmp/flagr-integration/integration.test.log; \
		exit_code=$${PIPESTATUS[0]}; \
		cat /tmp/flagr-integration/integration.test.log >> integration.test.log 2>/dev/null || true; \
		exit $$exit_code

INSTANCE ?= flagr_with_sqlite

retest: down test
