build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

build-integration-test:
	go test -tags=integration -count=1 -c -o integration.test .

down:
	docker compose down -v

up:
	docker network create flagr_integration_tests || true
	docker compose up -d

# Run Go integration tests against all 6 Docker Compose instances
test: build-image build-integration-test up
	@total=0; pass=0; \
	for instance in flagr_with_sqlite flagr_with_mysql flagr_with_mysql8 \
	                flagr_with_postgres9 flagr_with_postgres13 \
	                checkr_flagr_with_sqlite; do \
		echo "=== $$instance ==="; \
		total=$$((total+1)); \
		if docker compose exec -T integration-runner \
			sh -c "FLAGR_SERVER_URL=http://$$instance:18000 /app/integration.test -test.v -test.count=1"; then \
			pass=$$((pass+1)); \
		else \
			echo "FAILED: $$instance"; \
		fi; \
	done; \
	echo "=== $$pass/$$total instances passed ==="; \
	[ $$pass -eq $$total ]

# Shell test fallback (original behavior)
test-shell: build-image up
	docker run --rm --network flagr_integration_tests \
		-v $(PWD)/test.sh:/app/test.sh:ro \
		zhouzhuojie/docker-shakedown \
		sh /app/test.sh

# Single-instance test
test-instance: build-image build-integration-test up
	docker compose exec -T integration-runner \
		sh -c "FLAGR_SERVER_URL=http://$(INSTANCE):18000 /app/integration.test -test.v -test.count=1"

# Benchmark against a single instance
bench-instance: build-image build-integration-test up
	docker compose exec -T integration-runner \
		sh -c "FLAGR_SERVER_URL=http://$(INSTANCE):18000 /app/integration.test -test.bench=. -test.benchmem -test.count=1 -test.run=^$$"

retest: down test
