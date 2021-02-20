MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export BIN_DIR := $(MKFILE_PATH)bin
export WORK_DIR := $(MKFILE_PATH)work
export TEST_SUITE_DIR := $(MKFILE_PATH)test-suite

export STATE_DIR := $(WORK_DIR)/state
export IPFS_DIR := $(WORK_DIR)/ipfs
export CONFIG_DIR := $(WORK_DIR)/config
