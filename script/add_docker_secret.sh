if [ -z $1 ]; then
	NS=default
else
	NS=$1
fi

kubectl create secret generic regcred \
	--from-file=.dockerconfigjson=${HOME}/.docker/config.json \
	--type=kubernetes.io/dockerconfigjson \
	-n $NS
