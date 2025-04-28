#! /usr/bin/env sh
TEMPLATE_BASE=./config-template
OPERATIONS="refresh apply rm nop" # 탐색 실패를 nop로 표기한다. 굳이 별도의 환경변수로 탐색 실패 여부 탐색 안해도 됨.
HELP_MSG="[ERROR] USAGE: $0 [refresh/apply] [IMAGE_TAG(Mandatory)] [NAMESPACE(Optional; Default=default)] / rm [NAMESPACE]" 

NAMESPACE=default
TAG_NAME=
OPTION=

for OP in $OPERATIONS; do
	if [ $OP = "nop" ]; then
		echo $HELP_MSG
		exit 1
	elif [ $OP = $1 ]; then
		break
	fi	
done

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

