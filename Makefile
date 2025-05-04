SUDO ?= sudo # Docker에서 sudo 쓰지 않게.
BPFTOOL ?= bpftool
BUILD_DIR ?= ./bin
BPF_DIR ?= ./c
VMLINUX ?= $(BUILD_DIR)/vmlinux.h
GOBUILD_FLAG ?= -trimpath -ldflags="-s -w"

define log_run
	@printf "%-6s %s\n" "$(1)" "$(2)"
endef

.PHONY: all
all: bpf user

# Image Build =========================================

#.PHONY: build
#build: script/image_build.sh
#	$(call log_run, BUILD, ./script/image_build.sh)
#	@sh script/image_build.sh


# User Program ========================================

USER_PROGRAM := watcher runner collector driver

.PHONY: user
user: $(USER_PROGRAM)

.PHONY: $(USER_PROGRAM)
$(USER_PROGRAM): %: $(BUILD_DIR)/% 

# Build Go File

$(BUILD_DIR)/%: %/main.go $(wildcard ./log/format/*.go)
	$(call log_run, USER, $<)
	@go mod download
	@go build $(GOBUILD_FLAG) -o $@ ./$*
	@go clean -modcache

# BPF Program ========================================

BPF_PROGRAM := tc_capture

.PHONY: bpf $(BPF_PROGRAM)
bpf: $(BPF_PROGRAM)

$(BPF_PROGRAM): %: $(BUILD_DIR)/%.bpf.o 

$(BUILD_DIR)/%.bpf.o: $(BPF_DIR)/%.bpf.c $(VMLINUX)
	$(call log_run, BPF, $<)
	@clang -O2 -g -target bpf -I$(BUILD_DIR) -c $< -o $@

$(VMLINUX):
	$(call log_run, VMLINUX, $<)
	@mkdir $(BUILD_DIR)
	@$(SUDO) $(BPFTOOL) btf dump file /sys/kernel/btf/vmlinux format c > $(VMLINUX)

# Clean ==============================================

clean:
	$(call log_run, CLEAN, rm -rf $(BUILD_DIR))
	@rm -rf $(BUILD_DIR)
