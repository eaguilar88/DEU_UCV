.PHONY: loadenv

loadenv:
	@set -a && source .env && set +a && env | grep -E '^HTTP_'

start-backend:
	git pull --recursive-submodules
	docker compose up -d

start-prod:
	git pull --recursive-submodules
	docker compose -f docker-compose.prod.yml up -d --build

start-db:
	git pull --recursive-submodules
	docker compose -f docker-compose.dev.yml up -d