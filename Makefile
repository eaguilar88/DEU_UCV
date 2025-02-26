.PHONY: loadenv

loadenv:
	@set -a && source .env && set +a && env | grep -E '^HTTP_'

start-backend:
	docker compose up -d

start-prod:
	docker compose -f docker-compose.prod.yml up -d --build

start-db:
	docker compose -f docker-compose.dev.yml up -d