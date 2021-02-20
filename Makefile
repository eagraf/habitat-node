include ./common.mk

all : build

build : clean	
	mkdir -p $(BIN_DIR)
	$(MAKE) -C orchestrator build
	$(MAKE) -C test-suite build

clean :
	go clean -testcache
	rm -rf $(BIN_DIR)
	rm -rf $(WORK_DIR)

test : clean build
	go test -short ./...

run : build
	rm -rf $(IPFS_DIR)
	mkdir -p $(WORK_DIR)/ipfs
	$(BIN_DIR)/orchestrator

