GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
MAKE=make

BUILD_PATH=build
SERVER=./cmd/server
SERVER_BINARY=server
PARSER=./cmd/parser
PARSER_BINARY=parser
CLUSTERIZER=./cmd/clusterizer
CLUSTERIZER_BINARY=clusterizer
SPIDER=./cmd/spider
SPIDER_BINARY=spider
SCRIPTS_PATH=scripts

LNX_BUILD=$(build)/$(BINARY_NAME)
WIN_BUILD=$(build)/$(BINARY_NAME).exe

export GO111MODULE=on

.PHONY: all
all: test build linux

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BUILD_PATH)

# server
.PHONY: build_server
build_server: 
	$(GOBUILD) -o $(BUILD_PATH)/$(SERVER_BINARY) -v $(SERVER)

.PHONY: install_server
install_server:
	$(GOINSTALL) $(SERVER)

.PHONY: server
server:
	reflex -r '\.go$$' -s -- sh -c "config="./config/local.json" $(GORUN) $(SERVER) serve --badger ~/.bitgodine/badger --analysis ~/.bitgodine/analysis"

# parser
.PHONY: parser
parser:
	$(GORUN) $(PARSER) start --debug --skipped 300000 --file 0 --restored 20000000

.PHONY: build_parser
build_parser: 
	$(GOBUILD) -o $(BUILD_PATH)/$(PARSER_BINARY) -v $(PARSER)

.PHONY: install_parser
install_parser:
	$(GOINSTALL) $(PARSER)

# clusterizer
.PHONY: clusterizer
clusterizer:
	$(GORUN) $(CLUSTERIZER) start

.PHONY: build_clusterizer
build_clusterizer: 
	$(GOBUILD) -o $(BUILD_PATH)/$(CLUSTERIZER_BINARY) -v $(CLUSTERIZER)

.PHONY: install_clusterizer
install_clusterizer:
	$(GOINSTALL) $(CLUSTERIZER)

# spider
.PHONY: spider
spider:
	$(GORUN) $(SPIDER) crawl --cron=false

.PHONY: build_spider
build_spider: 
	$(GOBUILD) -o $(BUILD_PATH)/$(SPIDER_BINARY) -v $(SPIDER)

.PHONY: install_spider
install_spider:
	$(GOINSTALL) $(SPIDER)

.PHONY: build
build: build_parser build_server build_clusterizer build_spider

.PHONY: install
install: install_parser install_server install_clusterizer install_spider


# # Cross compilation
# linux: $(LNX_BUILD)
# windows: $(WIN_BUILD)
# # deploy:
# # 	ansible-playbook -i deploy/inventory.txt deploy/deploy.yml

# $(LNX_BUILD):
# 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/$(SERVER_BINARY) -v $(SERVER)
# $(WIN_BUILD):
# 	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/$(SERVER_BINARY).exe -v $(SERVER)