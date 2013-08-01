SHELL := /bin/bash
GOBOT_PACKAGE := github.com/bonnefoa/gobot
BUILD_SRC := build_src
BUILD_PATH := ${BUILD_SRC}/src/${GOBOT_PACKAGE}

BUILD_DIR := $(CURDIR)/.gopath

GOPATH ?= $(BUILD_DIR)
export GOPATH

GO_OPTIONS ?= -ldflags='-w'
ifeq ($(VERBOSE), 1)
GO_OPTIONS += -v
endif

BUILD_OPTIONS = -a 

SRC_DIR := $(GOPATH)/src

GOBOT_DIR := $(SRC_DIR)/$(GOBOT_PACKAGE)
GOBOT_MAIN := $(GOBOT_DIR)/gobot

GOBOT_BIN_RELATIVE := bin/gobot
GOBOT_BIN := $(CURDIR)/$(GOBOT_BIN_RELATIVE)

IRC_PARSER_GO=$(GOBOT_DIR)/message/parser.go
IRC_PARSER_YACC=$(GOBOT_DIR)/message/parser.y

.PHONY: all clean $(GOBOT_BIN) $(GOBOT_DIR)

all: $(GOBOT_BIN)

$(GOBOT_BIN): $(GOBOT_DIR) $(IRC_PARSER_GO)
	@mkdir -p  $(dir $@)
	@(cd $(GOBOT_MAIN); go build $(GO_OPTIONS) $(BUILD_OPTIONS) -o $@)
	@echo $(GOBOT_BIN_RELATIVE) is created.

$(GOBOT_DIR):
	@mkdir -p $(dir $@)
	@if [ -h $@ ]; then rm -f $@; fi; ln -sf $(CURDIR)/ $@
	@(cd $(GOBOT_MAIN); go get -d $(GO_OPTIONS))

$(IRC_PARSER_GO):
	@go tool yacc -o $(IRC_PARSER_GO) $(IRC_PARSER_YACC)

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

test: 
	@go test $(GO_OPTIONS) $(shell find . -iname '*_test.go' | xargs -I{} dirname {} | sort | uniq | sed -rn 's`\.`$(GOBOT_PACKAGE)`p') 2>&1 | tee >( grep test\ \-i | `sed -rn "s/.*(go\ test[^']+).*/ \1 /p"` )
