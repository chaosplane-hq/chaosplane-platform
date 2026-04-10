.PHONY: build-api test-api lint-api

build-api:
	cd apps/api && go build ./cmd/server/

test-api:
	cd apps/api && go test ./...

lint-api:
	cd apps/api && go vet ./...
