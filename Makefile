export PATH := $(HOME)/go/bin:$(PATH)

.PHONY: help db-up db-down db-status migrate-up migrate-down migrate-status sqlc-gen api-test api-run web-install web-dev web-build check

help:
	@echo "GeevheitNeural - Comandos disponíveis:"
	@echo "  make db-up          - Inicia os serviços locais do Supabase (Postgres, Auth)"
	@echo "  make db-down        - Para os serviços do Supabase"
	@echo "  make migrate-up     - Executa migrations goose no banco local"
	@echo "  make migrate-down   - Reverte a última migration goose"
	@echo "  make migrate-status - Exibe o status das migrations goose"
	@echo "  make sqlc-gen       - Gera código Go tipado a partir de queries SQL via sqlc"
	@echo "  make api-test       - Executa testes unitários e de integração da API Go (-race)"
	@echo "  make api-run        - Inicia a API Go localmente"
	@echo "  make web-install    - Instala dependências do frontend"
	@echo "  make web-dev        - Inicia dev server do Next.js"
	@echo "  make web-build      - Gera o static export da PWA para Cloudflare Pages"
	@echo "  make check          - Roda linters e testes da API e Web"

db-up:
	cd db && supabase start

db-down:
	cd db && supabase stop

db-status:
	cd db && supabase status

MIGRATIONS_DIR := db/migrations
DB_URL ?= postgresql://postgres:postgres@127.0.0.1:54332/postgres?sslmode=disable

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

sqlc-gen:
	cd api && sqlc generate

api-test:
	cd api && go test -v -race ./...

api-run:
	cd api && go run ./cmd/geevheit-api

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

check: api-test
	cd web && npm run lint && npm run test --if-present
