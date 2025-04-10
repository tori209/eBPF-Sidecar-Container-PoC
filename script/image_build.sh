#! /usr/bin/env bash

BUILD_TAG=$1

IMAGE_NAME="tori209/ebpf-sidecar-poc:$BUILD_TAG"

EXIST_CNT=`docker images $IMAGE_NAME | wc -l`
if [ $EXIST_CNT -ge 2 ]; then
	echo "overwrite image ${IMAGE_NAME}..."
	docker rmi $IMAGE_NAME
fi

docker build . --tag ${IMAGE_NAME} &&
docker push ${IMAGE_NAME}
