.PHONY: loadenv start-prod-landing seed-landing start-landing stop-landing start-espacios stop-espacios start-deu stop-deu migrate-up migrate-down migrate-down-all migrate-version migrate-force migrate-create migrate-up-prod

loadenv:
	@set -a && . .env && set +a && env | grep -E '^HTTP_'

start-backend:
	docker compose up --build -d

start-prod:
	git submodule update --init --recursive
	docker compose -f docker-compose.prod.yml up  --build -d

start-deu: start-prod start-prod-landing start-espacios

stop-deu:
	docker compose -f docker-compose.prod.yml down
	docker compose -f docker-compose.landing.yml down
	docker compose -f docker-compose.espacios.yml down

start-prod-landing:
	git submodule sync -- landing
	git submodule update --init --recursive landing
	docker compose -f docker-compose.landing.yml up -d --build

stop-landing:
	docker compose -f docker-compose.landing.yml down

seed-landing:
	docker compose -f docker-compose.landing.yml exec landing ./bin/rails db:seed

start-espacios:
	git submodule sync -- espacios-universitarios
	git submodule update --init --recursive espacios-universitarios
	docker compose -f docker-compose.espacios.yml up -d --build

stop-espacios:
	docker compose -f docker-compose.espacios.yml down

start-db:
	docker compose -f docker-compose.dev.yml up  --build -d

migrate-up:
	set -a && . ./.env && set +a && \
	docker compose -f docker-compose.yml run --rm migrate \
	  -path=/migrations -database="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@db:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up

migrate-down:
	set -a && . ./.env && set +a && \
	docker compose -f docker-compose.yml run --rm migrate \
	  -path=/migrations -database="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@db:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" down 1

migrate-down-all:
	set -a && . ./.env && set +a && \
	docker compose -f docker-compose.yml run --rm migrate \
	  -path=/migrations -database="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@db:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" down -all

migrate-version:
	set -a && . ./.env && set +a && \
	docker compose -f docker-compose.yml run --rm migrate \
	  -path=/migrations -database="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@db:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" version

migrate-force:
	set -a && . ./.env && set +a && \
	docker compose -f docker-compose.yml run --rm migrate \
	  -path=/migrations -database="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@db:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" force $(version)

migrate-create:
	docker compose -f docker-compose.yml run --rm migrate \
	  create -ext sql -dir /migrations -seq $(name)

migrate-up-prod:
	set -a && . ./.env && set +a && \
	docker compose -f docker-compose.prod.yml run --rm migrate \
	  -path=/migrations -database="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@db:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" up

stop-backend:
	docker compose down

stop-prod:
	docker compose -f docker-compose.prod.yml down -v

stop-db:
	docker compose -f docker-compose.dev.yml down -v

generate-mocks:
	docker pull vektra/mockery:v3.7.0 && docker run --user 1000:1000 --rm -v ${PWD}/backend:/src -w /src vektra/mockery:v3.7.0 --config .mockery.yaml

go-test:
	cd backend && go test -count=1 -short -cover ./...

git-pull:
	git pull

quality:
	cd backend && \
	go vet ./... && \
	go fmt ./... && \
	golangci-lint run && \
	go mod tidy && \
	go test -count=1 -race -short -cover ./...

refresh: stop-db start-db

refresh-prod: stop-prod start-prod
