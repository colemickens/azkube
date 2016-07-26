#!/usr/bin/env bash

# ./util.sh kubectl get nodes  # invokes kubectl with your args
# ./util.sh curl api/v1/nodes  # invokes curl with your args trailing the master server prefix
# ./util.sh configure-kubectl  # reconfigure your users kubeconfig settings to point to this cluster
# ./util.sh deploy-addons       # deploy addons (ns/kube-system, svc+rc/kube-dashboard, svc+rc/skydns+kube2sky)
# ./util.sh copykey            # copys private key to master
# ./util.sh ssh                # ssh into the master

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

DEPLOYMENTNAME="{{.DeploymentName}}"
MASTERFQDN="{{.MasterFQDN}}"
USERNAME="{{.Username}}"

AZKUBE_BRANCH="v0.0.4"

KUBECTL="`which kubectl`"
dockerized_kubectl="${DIR}/../../../../_output/dockerized/bin/linux/amd64/kubectl"
if [[ -f "${dockerized_kubectl}" ]]; then
	KUBECTL="${dockerized_kubectl}"
	echo "Using _output's kubectl: ${KUBECTL}"
else
	echo "Using system's kubectl: ${KUBECTL}"
fi

cmd_kubectl() {
	${KUBECTL} \
		--cluster="${DEPLOYMENTNAME}" \
		--context="${DEPLOYMENTNAME}" \
		--client-certificate="${DIR}/client.crt" \
		--client-key="${DIR}/client.key" \
		--certificate-authority="${DIR}/ca.crt" \
		--server="https://${MASTERFQDN}:6443" \
		"${@}"
}

cmd_curl() {
	curl \
		--cert "${DIR}/client.crt" \
		--key "${DIR}/client.key" \
		--cacert "${DIR}/ca.crt" \
		https://${MASTERFQDN}:6443/"${@}"
}

cmd_configure-kubectl() {
	kubectl config set-cluster "${DEPLOYMENTNAME}" --server="https://${MASTERFQDN}:6443" --certificate-authority="${DIR}/ca.crt"
	kubectl config set-credentials "${USERNAME}_user" --client-certificate="${DIR}/client.crt" --client-key="${DIR}/client.key"
	kubectl config set-context "${DEPLOYMENTNAME}" --cluster="${DEPLOYMENTNAME}" --user="${USERNAME}_user"
	kubectl config use-context "${DEPLOYMENTNAME}"
}

cmd_deploy-addons() {
	cmd_kubectl create -f "https://raw.githubusercontent.com/colemickens/azkube/v0.0.4/templates/coreos/addons/kube-system.yaml"
	cmd_kubectl create -f "https://raw.githubusercontent.com/colemickens/azkube/v0.0.4/templates/coreos/addons/skydns.yaml"
	cmd_kubectl create -f "https://raw.githubusercontent.com/colemickens/azkube/v0.0.4/templates/coreos/addons/kube-dashboard.yaml"
}

cmd_copykey() {
	scp -i "${USERNAME}_rsa" "${USERNAME}_rsa" "${USERNAME}@${MASTERFQDN}":"/home/${USERNAME}/${USERNAME}_rsa"
}

cmd_ssh() {
	ssh -i "${USERNAME}_rsa" ${USERNAME}@${MASTERFQDN}
}

cmd="$1"
shift 1

"cmd_${cmd}" "${@}"
