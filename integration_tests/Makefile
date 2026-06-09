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
		if docker compose exec -T integration-runner \
			sh -c "FLAGR_SERVER_URL=http://$$instance:18000 go test -tags=integration -v -count=1 ./integration_tests/" >> integration.test.log 2>&1; then \
			pass=$$((pass+1)); \
		else \
			echo "FAILED: $$instance" | tee -a integration.test.log; \
		fi; \
	done; \
	echo "=== $$pass/$$total instances passed ===" | tee -a integration.test.log; \
	[ $$pass -eq $$total ]

# Single-instance test
test-instance: build-image up
	docker compose exec -T integration-runner \
		sh -c "FLAGR_SERVER_URL=http://$(INSTANCE):18000 go test -tags=integration -v -count=1 ./integration_tests/"

# Benchmark against a single instance (200ms per benchmark for faster feedback)
bench-instance: build-image up
	docker compose exec -T integration-runner \
		sh -c "FLAGR_SERVER_URL=http://$(INSTANCE):18000 go test -tags=integration -bench=. -benchmem -benchtime=200ms -count=1 -run=^$$ ./integration_tests/"

INSTANCE ?= flagr_with_sqlite

retest: down test
