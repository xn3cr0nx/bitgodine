SERVER=./cmd/server
SERVER_BINARY=server

.PHONY: build_server
build_server: 
	$(GOBUILD) -o $(BUILD_PATH)/$(SERVER_BINARY) -v $(SERVER)

.PHONY: install_server
install_server:
	$(GOINSTALL) $(SERVER)

.PHONY: server-docs
server-docs:
	# more recent version not working (1.7.0). had to downgrade to 1.6.7 in order to make it work
	# swag init --parseDependency --parseInternal -g cmd/server/main.go
	swag init -g cmd/server/main.go

.PHONY: server
server: server-docs
	reflex -r '\.go$$' -R './docs/*.go' -s -- sh -c 'config="./config/local.json" $(GORUN) $(SERVER) serve --badger ~/.bitgodine/badger --analysis ~/.bitgodine/analysis'

docker-server:
	$(DC) $(DCUP) bitgodine_server
