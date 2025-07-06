#! /usr/bin/env sh
TEMPLATE_BASE=./config-template
OPERATIONS="refresh apply rm nop" # 탐색 실패를 nop로 표기한다. 굳이 별도의 환경변수로 탐색 실패 여부 탐색 안해도 됨.
HELP_MSG="[ERROR] USAGE: $0 [refresh/apply] [IMAGE_TAG(Mandatory)] [NAMESPACE(Optional; Default=default)] [EXECUTOR_CNT(OPTIONAL; Default=1)]/ rm [NAMESPACE]" 

NAMESPACE=default
TAG_NAME=
EXECUTOR_CNT=1

for OP in $OPERATIONS; do
	if [ $OP = "nop" ]; then
		echo $HELP_MSG
		exit 1
	elif [ $OP = $1 ]; then
		break
	fi	
done

if [ $OP != "rm" ]; then
	if [ -z $2 ]; then
		echo "[ERROR] USAGE $0 [IMAGE_TAG(Mandatory)] [NAMESPACE(Optional; Default=default)]" > /dev/stderr
		exit 1
	fi
	TAG_NAME=$2

	if [ ! -z $3 ]; then
		NAMESPACE=$3
		if [ ! -z $4 ]; then
			EXECUTOR_CNT=$4
		fi
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
export EXECUTOR_CNT
export POSTGRES_ADMIN_PASSWORD=$(kubectl get secret --namespace postgresql postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
export POSTGRES_URL="postgresql.postgresql.svc.cluster.local:5432"
export POSTGRES_DB="tasklist"
export COLLECTOR_INFLUXDB_ORG="data-executor"
export COLLECTOR_INFLUXDB_BUCKET="collector"
export INFLUXDB_URL="http://influxdb-influxdb2.influxdb.svc.cluster.local"
export MAKE_ANOMALY="False"
export RUNNER_MODE_IS_SIMPLE="True"

for secret_env in `cat ./script/endpoint.secret`; do
	export $secret_env
done

# Delete first
for TEMPLATE in `ls ${TEMPLATE_BASE}`; do
	echo ${TEMPLATE}
	if [ $OP != "apply" ]; then
		envsubst < "${TEMPLATE_BASE}/${TEMPLATE}" | kubectl delete -n ${NAMESPACE} -f -
#		if [ $? -ne 0 ]; then
#			echo "Error Occured. Terminate."
#			exit 1
#		fi
	fi
	if [ $OP != "rm" ]; then
		envsubst < "${TEMPLATE_BASE}/${TEMPLATE}" | kubectl apply -n ${NAMESPACE} -f -
#		if [ $? -ne 0 ]; then
#			echo "Error Occured. Terminate."
#			exit 1
#		fi
	fi
done

