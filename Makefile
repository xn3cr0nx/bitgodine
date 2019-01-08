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
BINARY_NAME=bitgodine

LNX_BUILD=$(build)/$(BINARY_NAME)_lnx
WIN_BUILD=$(build)/$(BINARY_NAME).exe

.PHONY: deploy

all: test build linux
build: deps 
	$(GOBUILD) -o $(BUILD_PATH)/$(BINARY_NAME) -v $(ENTRY)
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BUILD_PATH)/$(BINARY_NAME)
	rm -f $(BUILD_PATH)/$(BINARY_NAME)_lnx
	rm -f $(BUILD_PATH)/$(BINARY_NAME).exe
run:
	$(GORUN) $(ENTRY)
deps:
	dep ensure
deps-win:
	$(GOGET) github.com/inconshreveable/mousetrap

# Interacting with bitgodine cli
sync:
	$(GOINSTALL) $(ENTRY) && $(BINARY_NAME) -n regtest sync

# Cross compilation
linux: $(LNX_BUILD)
windows: $(WIN_BUILD)
# deploy:
# 	ansible-playbook -i deploy/inventory.txt deploy/deploy.yml

$(LNX_BUILD):
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/$(BINARY_NAME)_lnx -v $(ENTRY)
$(WIN_BUILD): deps-win
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_PATH)/$(BINARY_NAME).exe -v $(ENTRY)
