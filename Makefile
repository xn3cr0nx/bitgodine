# go
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
MAKE=make

BUILD_PATH=build
SCRIPTS_PATH=scripts

# Docker
DOCKER=docker
DC=docker-compose
DCUP=up -d

LNX_BUILD=$(build)/$(BINARY_NAME)
WIN_BUILD=$(build)/$(BINARY_NAME).exe

export GO111MODULE=on

include cmd/parser/Makefile
include cmd/server/Makefile
include cmd/clusterizer/Makefile
include cmd/cli/Makefile
include cmd/spider/Makefile

default: server

.PHONY: all
all: test build

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BUILD_PATH)


.PHONY: build
build: build_parser build_server build_clusterizer build_spider

.PHONY: install
install: install_parser install_server install_clusterizer install_spider

docker-deps:
	$(DC) $(DCUP) postgres redis

docker-otel:
	$(DC) $(DCUP) jaeger prometheus grafana config-concat loki fluent-bit

# # deploy:
# # 	ansible-playbook -i deploy/inventory.txt deploy/deploy.yml