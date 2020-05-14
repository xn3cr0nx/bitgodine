GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
MAKE=make

BUILD_PATH=build
ENTRY=./cmd/bitgodine
SPIDER=./cmd/spider
BINARY_NAME=bitgodine
SCRIPTS_PATH=scripts

LNX_BUILD=$(build)/$(BINARY_NAME)
WIN_BUILD=$(build)/$(BINARY_NAME).exe

export GO111MODULE=on


.PHONY: all
all: test build linux

.PHONY: build
build: 
	$(GOBUILD) -o $(BUILD_PATH)/$(BINARY_NAME) -v $(ENTRY)

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BUILD_PATH)

.PHONY: run
run:
	$(GORUN) $(ENTRY)

.PHONY: install
# Interacting with bitgodine cli
install:
	$(GOINSTALL) $(ENTRY)

.PHONY: run
run:
	reflex -r '\.go$$' -s -- sh -c "$(GORUN) $(ENTRY) serve --badger ~/.bitgodine/badger --analysis ~/.bitgodine/analysis"

.PHONY: spider
spider:
	$(GORUN) $(SPIDER) crawl --cron=false



# Cross compilation
linux: $(LNX_BUILD)
windows: $(WIN_BUILD)
# deploy:
# 	ansible-playbook -i deploy/inventory.txt deploy/deploy.yml

$(LNX_BUILD):
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/$(BINARY_NAME) -v $(ENTRY)
$(WIN_BUILD):
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/$(BINARY_NAME).exe -v $(ENTRY)