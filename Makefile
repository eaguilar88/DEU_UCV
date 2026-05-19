.PHONY: loadenv start-prod-landing seed-landing

loadenv:
	@set -a && source .env && set +a && env | grep -E '^HTTP_'

start-backend:
	docker compose up --build -d

start-prod:
	git submodule update --recursive
	docker compose -f docker-compose.prod.yml up  --build -d

start-prod-landing:
	git submodule sync -- landing
	git submodule update --init --recursive landing
	docker compose -f docker-compose.prod.yml up -d --build traefik db backend landing

seed-landing:
	docker compose -f docker-compose.prod.yml exec landing ./bin/rails db:seed

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