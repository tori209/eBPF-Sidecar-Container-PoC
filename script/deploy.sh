#! /usr/bin/env sh

TEMPLATE_BASE=./config-template
NAMESPACE=default
TAG_NAME=
OPTION=

if [ -z $1 ]; then 
	echo "[ERROR] USAGE: $0 [refresh/apply] [IMAGE_TAG(Mandatory)] [NAMESPACE(Optional; Default=default)] / rm [NAMESPACE]" > /dev/stderr
	exit 1
else
	OPTION=$1
fi

if [ $OPTION != "rm" ]; then
	if [ -z $2 ]; then
		echo "[ERROR] USAGE $0 [IMAGE_TAG(Mandatory)] [NAMESPACE(Optional; Default=default)]" > /dev/stderr
		exit 1
	elif [ -z $3 ]; then
		TAG_NAME=$2
	else 
		TAG_NAME=$2
		NAMESPACE=$3
	fi
else
	if [ ! -z $2 ]; then
		NAMESPACE=$2
	fi	
	TAG_NAME="dummy"
fi

if [ $(basename `pwd`) = "script" ]; then
	cd ..
fi

export NAMESPACE
export TAG_NAME
# Delete first
for TEMPLATE in `ls ${TEMPLATE_BASE}`; do
	echo ${TEMPLATE}
	if [ $OPTION != "apply" ]; then
		envsubst < "${TEMPLATE_BASE}/${TEMPLATE}" | kubectl delete -n ${NAMESPACE} -f -
	fi
	if [ $OPTION != "rm" ]; then
		envsubst < "${TEMPLATE_BASE}/${TEMPLATE}" | kubectl apply -n ${NAMESPACE} -f -
	fi
done

