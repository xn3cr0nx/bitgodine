CLUSTERIZER=./cmd/clusterizer
CLUSTERIZER_BINARY=clusterizer

.PHONY: clusterizer
clusterizer:
	$(GORUN) $(CLUSTERIZER)

.PHONY: build_clusterizer
build_clusterizer: 
	$(GOBUILD) -o $(BUILD_PATH)/$(CLUSTERIZER_BINARY) -v $(CLUSTERIZER)

.PHONY: install_clusterizer
install_clusterizer:
	$(GOINSTALL) $(CLUSTERIZER)
