include ./common.mk

all : build

build : clean	
	mkdir -p $(BIN_DIR)
	$(MAKE) -C orchestrator build
	$(MAKE) -C fs build
	$(MAKE) -C client build

clean :
	go clean -testcache
	rm -rf $(BIN_DIR)
	rm -rf $(WORK_DIR)
	rm -rf auth/

test : clean build
	go test -short ./...

run : build
	mkdir -p $(WORK_DIR)/ipfs
	rm -rf $(IPFS_DIR)
	$(BIN_DIR)/orchestrator

