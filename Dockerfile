FROM golang:1.24.1-bookworm
LABEL maintainer="kyoungho2018@gist.ac.kr"
LABEL description="Sidecar-Container for eBPF Network Monitoring"

RUN apt update

# dependency for bpftool
RUN apt install -y libelf-dev zlib1g-dev clang libcap-dev libbfd-dev llvm libbpf-dev

# Copy Current Directory to Docker Image
COPY . /go/src/executor
WORKDIR /go/src/executor

# bpftool
RUN wget https://github.com/libbpf/bpftool/releases/download/v7.4.0/bpftool-v7.4.0-amd64.tar.gz \
    && tar -zxvf bpftool-v7.4.0-amd64.tar.gz  \
    && rm bpftool-v7.4.0-amd64.tar.gz  \
	&& chmod u+x bpftool
ENV BPFTOOL=/go/src/executor/bpftool

# Build
ENV SUDO=
RUN make bpf
RUN make user
