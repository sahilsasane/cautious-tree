include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api \
		-mongo-uri="$${MONGO_URI}" \
		-jwt-secret="$${JWT_SECRET}"
# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo "Formatting code"
	go fmt ./...
	@echo "Vetting code"
	go vet ./...
	staticcheck ./...
	@echo "Running tests"
	go test -race -vet=off ./...

.PHONY: vendor
vendor:
	@echo "Tidying and verifying module dependencies"
	go mod tidy
	go mod verify
	@echo "Vendoring dependencies"
	go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #
current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'
## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo "Building api"
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/api-linux-amd64 ./cmd/api