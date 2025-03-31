FROM golang:1.24.1-bookworm
LABEL maintainer="kyoungho2018@gist.ac.kr"
LABEL description="Sidecar-Container for eBPF Network Monitoring"

RUN apt update

# dependency for bpftool
RUN apt install -y libelf-dev zlib1g-dev

# Copy Current Directory to Docker Image
COPY . /go/src/executor
WORKDIR /go/src/executor

# Build
ENV SUDO=
RUN make bpf runner watcher
