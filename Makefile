PKGS := $(shell go list ./...| grep -v /mocks)

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(PKGS) -v -short -coverprofile=coverage.out -timeout 10s

.PHONY: integration
integration:
	go test -coverpkg=$(shell echo "${PKGS}" | tr ' ' ',') -v -coverprofile=coverage.out -p=1 -timeout=60s $(PKGS)

.PHONY: humantest
humantest:
ifndef run
	LOG_DEBUG=true LOG_HUMAN=true richgo test -v -p=1 -timeout=60s $(PKGS)
else
	LOG_DEBUG=true LOG_HUMAN=true richgo test -v -p=1 -timeout=60s $(PKGS) -run $(run)
endif

.PHONY: coverage
coverage:
	go tool cover -html=coverage.out

.PHONY: benchmark
benchmark:
	go test $(PKGS) -bench=. -benchmem

.PHONY: mock
mock: # Generate new mocks for all interfaces within this package, see https://github.com/vektra/mockery
	mockery --recursive --name=AppService