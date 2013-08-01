GOBOT_PACKAGE := github.com/bonnefoa/gobot
BUILD_SRC := build_src
BUILD_PATH := ${BUILD_SRC}/src/${GOBOT_PACKAGE}

BUILD_DIR := $(CURDIR)/.gopath

GOPATH ?= $(BUILD_DIR)
export GOPATH

GO_OPTIONS ?= -a -ldflags='-w'
ifeq ($(VERBOSE), 1)
GO_OPTIONS += -v
endif

BUILD_OPTIONS = -a 

SRC_DIR := $(GOPATH)/src

GOBOT_DIR := $(SRC_DIR)/$(GOBOT_PACKAGE)
GOBOT_MAIN := $(GOBOT_DIR)/gobot

GOBOT_BIN_RELATIVE := bin/gobot
GOBOT_BIN := $(CURDIR)/$(GOBOT_BIN_RELATIVE)

.PHONY: all clean $(GOBOT_BIN) $(GOBOT_DIR)

all: $(GOBOT_BIN)

$(GOBOT_BIN): $(GOBOT_DIR)
	@mkdir -p  $(dir $@)
	@(cd $(GOBOT_MAIN); go build $(GO_OPTIONS) $(BUILD_OPTIONS) -o $@)
	@echo $(GOBOT_BIN_RELATIVE) is created.

$(GOBOT_DIR):
	@mkdir -p $(dir $@)
	@if [ -h $@ ]; then rm -f $@; fi; ln -sf $(CURDIR)/ $@

deps: $(GOBOT_DIR)

clean:
	@rm -rf $(dir $(GOBOT_BIN))
ifeq ($(GOPATH), $(BUILD_DIR))
	@rm -rf $(BUILD_DIR)
else ifneq ($(GOBOT_DIR), $(realpath $(GOBOT_DIR)))
	@rm -f $(GOBOT_DIR)
endif

fmt:
	@gofmt -s -l -w .
