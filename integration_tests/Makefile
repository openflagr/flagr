build-image:
	docker build -f ./Dockerfile-Integration-Test -t flagr_integration_tests ../

down:
	docker compose down -v

up:
	docker network create flagr_integration_tests || true
	docker compose up -d

test:
	docker cp ./test.sh flagr-shakedown:/app/test.sh
	docker compose exec -T shakedown bash ./test.sh

retest: down up test
