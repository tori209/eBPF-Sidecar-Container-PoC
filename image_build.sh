#! /usr/bin/env bash

IMAGE_NAME="tori209/ebpf-sidecar-poc:$1"

EXIST_CNT=`docker images $IMAGE_NAME | wc -l`

if [ $EXIST_CNT -ge 2 ]; then
	docker rmi $IMAGE_NAME
fi

docker build . --tag tori209/ebpf-sidecar-poc:$1 &&
docker push tori209/ebpf-sidecar-poc:$1
