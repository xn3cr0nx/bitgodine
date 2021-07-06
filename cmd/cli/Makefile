CLI=./cmd/cli
CLI_BINARY=cli

.PHONY: build_cli
build_cli: 
	$(GOBUILD) -o $(BUILD_PATH)/$(CLI_BINARY) -v $(CLI)

.PHONY: install_cli
install_cli:
	$(GOINSTALL) $(CLI)

.PHONY: cli
cli:
	$(GORUN) $(CLI)
