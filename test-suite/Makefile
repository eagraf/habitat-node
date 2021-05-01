include ../common.mk

reset:
	go build -o $(BIN_DIR)/test-suite
	rm -rf $(WORK_DIR)
	mkdir -p $(WORK_DIR)

test-process-manager : reset 
	$(BIN_DIR)/test-suite process_manager $(TEST_SUITE_DIR)/tests/reconfigure_backnet_ipfs.json
