PKGS := $(shell go list ./...)

lint: # Ignore installation recommendations and always download latest version of golangci-lint
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.31.0
	golangci-lint run

test:
	go test $(PKGS) -v -short -coverprofile=coverage.out -timeout 10s

integration:
	go test -coverpkg=$(shell echo "${PKGS}" | tr ' ' ',') -v -coverprofile=coverage.out -p=1 -timeout=60s ./...

coverage:
	go tool cover -html=coverage.out

dep:
	go mod download

bench:
	go test $(PKGS) -bench=. -benchmem

.PHONY: lint test integration coverage dep bench