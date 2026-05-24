.PHONY: loadenv start-prod-landing seed-landing start-landing stop-landing start-espacios stop-espacios start-deu stop-deu

loadenv:
	@set -a && source .env && set +a && env | grep -E '^HTTP_'

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

refresh-prod: stop-PROD start-prod