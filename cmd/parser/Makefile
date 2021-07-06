PARSER=./cmd/parser
PARSER_BINARY=parser

.PHONY: parser
parser:
	$(GORUN) $(PARSER) --debug

.PHONY: build_parser
build_parser: 
	$(GOBUILD) -o $(BUILD_PATH)/$(PARSER_BINARY) -v $(PARSER)

.PHONY: install_parser
install_parser:
	$(GOINSTALL) $(PARSER)