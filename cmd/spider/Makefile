SPIDER=./cmd/spider
SPIDER_BINARY=spider

.PHONY: spider
spider:
	$(GORUN) $(SPIDER) --cron=false

.PHONY: build_spider
build_spider: 
	$(GOBUILD) -o $(BUILD_PATH)/$(SPIDER_BINARY) -v $(SPIDER)

.PHONY: install_spider
install_spider:
	$(GOINSTALL) $(SPIDER)