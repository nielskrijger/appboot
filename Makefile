PKGS := $(shell go list ./...| grep -v /mocks)

lint:
	golangci-lint run

test:
	go test $(PKGS) -v -short -coverprofile=coverage.out -timeout 10s

integration:
	go test -coverpkg=$(shell echo "${PKGS}" | tr ' ' ',') -v -coverprofile=coverage.out -p=1 -timeout=60s $(PKGS)

humantest:
	LOG_HUMAN=true richgo test -v -p=1 -timeout=60s $(PKGS)

coverage:
	go tool cover -html=coverage.out

dep:
	go mod download

benchmark:
	go test $(PKGS) -bench=. -benchmem

mock: # Generate new mocks for all interfaces within this package, see https://github.com/vektra/mockery
	mockery --recursive --name=AppService

.PHONY: lint test integration humantest coverage dep benchmark mock