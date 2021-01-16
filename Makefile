include ./common.mk

all : build

build : clean	
	mkdir -p $(BIN_DIR)
	$(MAKE) -C orchestrator build

clean :
	go clean -testcache
	rm -rf $(BIN_DIR)

test : clean build
	go test -short ./...

run : build
	$(BIN_DIR)/orchestrator

