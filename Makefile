BUILD_DIR=./bin
BPF_DIR=./c
VMLINUX=${BPF_DIR}/vmlinux.h

# UserProgram ========================================

watcher: bin/watcher

runner: bin/runner

collector: bin/collector


bin/watcher: watcher/main.go
	go build -o bin/watcher ./watcher

bin/runner: runner/main.go
	go build -o bin/runner ./runner

bin/collector: collector/main.go
	go build -o bin/collector ./collector

# BPF Program ========================================

bpf: tc_capture

tc_capture: ${BPF_DIR}/tc_capture.bpf.c ${VMLINUX}
	clang -O2 -g -target bpf -I${VMLINUX} -c ${BPF_DIR}/tc_capture.bpf.c -o ${BUILD_DIR}/tc_capture.bpf.o

${VMLINUX}:
	sudo bpftool btf dump file /sys/kernel/btf/vmlinux format c > ${VMLINUX}
