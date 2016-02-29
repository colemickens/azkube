#!/usr/bin/env sh

# ./util.sh curl api/v1/nodes  # invokes curl with your args trailing the master server prefix
# ./util.sh kubectl get nodes  # TODO: fix, fails to negotiate version?
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

cmd_kubectl() {
	kubectl \
		--cluster="${DEPLOYMENTNAME}" \
		--context="${DEPLOYMENTNAME}" \
		--client-certificate="${DIR}/client.crt" \
		--client-key="${DIR}/client.key" \
		--certificate-authority="${DIR}/ca.crt" \
		--server="https://${MASTERFQDN}:6443/" \
		"${@}"
}

cmd_curl() {
	curl \
		--cert "${DIR}/client.crt" \
		--key "${DIR}/client.key" \
		--cacert "${DIR}/ca.crt" \
		https://${MASTERFQDN}:6443/"${@}"
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
