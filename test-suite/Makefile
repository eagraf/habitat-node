include ../common.mk

build :
	go build -o $(BIN_DIR)/test-suite

reset : build
	rm -rf $(WORK_DIR)
	mkdir -p $(WORK_DIR)

test-process-manager : reset 
	$(BIN_DIR)/test-suite process_manager $(TEST_SUITE_DIR)/tests/reconfigure_backnet_ipfs.json

test-community-state-machine: reset 
	$(BIN_DIR)/test-suite community_state_machine $(TEST_SUITE_DIR)/tests/snapshot_wal_test.json
