.PHONY: loadenv

loadenv:
	@set -a && source .env && set +a && env | grep -E '^HTTP_'

start-backend:
	git submodule update --recursive
	docker compose up -d

start-prod:
	git submodule update --recursive
	docker compose -f docker-compose.prod.yml up -d --build

start-db:
	git submodule update --recursive
	docker compose -f docker-compose.dev.yml up -d

stop:
	docker compose down

stop-prod:
	docker compose -f docker-compose.prod.yml down

stop-db:
	docker compose -f docker-compose.dev.yml down

generate-mocks:
	docker run --user 1000:1000 --rm -v ${PWD}/backend:/src -w /src vektra/mockery --config .mockery.yaml

