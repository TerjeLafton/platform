mod todo "apps/todo"
mod id "apps/id"
mod web "apps/web"

generate-proto:
  @echo "Generating protobufs..."
  protoc \
    --proto_path=proto \
    --go_out=libs/proto-stubs \
    --go_opt=paths=source_relative \
    proto/common/v1/*.proto \
    proto/todo/v1/*.proto \
    proto/id/v1/*.proto

generate: generate-proto
  @echo "Generating sqlc..."
  cd apps/todo && sqlc generate
  cd apps/id && sqlc generate
  @echo "Generating templ..."
  cd apps/web && templ generate
  @echo "Generating CSS..."
  cd apps/web && tailwindcss -i static/css/input.css -o static/css/style.css --minify
  @echo "Done!"

migrate-diff:
  @docker exec postgres psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'atlas_dev'" | grep -q 1 || docker exec postgres psql -U postgres -c "CREATE DATABASE atlas_dev"
  atlas migrate diff --env local

migrate-apply:
  atlas migrate apply --env local

db-clean:
  @echo "Cleaning database (removing volumes)..."
  podman compose down -v
  podman compose up -d
  @echo "Database cleaned. Run 'just migrate-apply' to apply migrations."

compose-up:
  podman compose up -d

compose-down:
  podman compose down

run-all:
  #!/usr/bin/env bash
  trap 'kill 0' EXIT
  (cd apps/id && go run ./cmd/server) &
  (cd apps/todo && go run ./cmd/server) &
  (cd apps/web && go run ./cmd/server) &
  wait
