include ../common.mk

build :
	go build -o $(BIN_DIR)/test-suite

test-process-manager : build
	rm -rf $(WORK_DIR)
	mkdir -p $(WORK_DIR)
	$(BIN_DIR)/test-suite process_manager $(TEST_SUITE_DIR)/tests/reconfigure_backnet_ipfs.json
