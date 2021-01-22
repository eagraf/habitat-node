MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export BIN_DIR := $(MKFILE_PATH)bin
export WORK_DIR := $(MKFILE_PATH)work

export STATE_DIR := $(WORK_DIR)/state
export IPFS_DIR := $(WORK_DIR)/ipfs
