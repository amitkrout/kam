#!/bin/sh

# fail if some commands fails
set -e

# Do not show token in CI log
set +x
export GITHUB_TOKEN=`cat $KAM_GITHUB_TOKEN_FILE`
export KUBEADMIN_PASSWORD=`cat $KUBEADMIN_PASSWORD_FILE`

# show commands
set -x
export CI="prow"
export PRNO="$(jq .refs.pulls[0].number <<< $(echo $JOB_SPEC))"

mkdir -p $HOME/.ssh && chmod 0700 $HOME/.ssh
cp $KAM_SSH_PRIVATE_KEY_FILE $HOME/.ssh/
chmod 600 $HOME/.ssh/id_rsa
echo -e "Host github.com\n\tStrictHostKeyChecking no\n" > $HOME/.ssh/config
chmod 600 $HOME/.ssh/config
ls $HOME/.ssh/
# mkdir -p /.ssh && chmod 0700 /.ssh
# cp $KAM_SSH_PRIVATE_KEY_FILE /.ssh/
# chmod 600 /.ssh/id_rsa
# echo -e "Host github.com\n\tStrictHostKeyChecking no\n" > /.ssh/config
# chmod 600 /.ssh/config

make prepare-test-cluster
make bin

INSTALL_ARGOCD="./scripts/install-argocd.sh"
sh $INSTALL_ARGOCD

INSTALL_DOCKER="./scripts/install-docker-cli.sh"
sh $INSTALL_DOCKER

INSTALL_KUBECTL="./scripts/install-kubectl.sh"
sh $INSTALL_KUBECTL

export PATH="$PATH:$(pwd)/bin"

# Copy kubeconfig to temporary kubeconfig file and grant
# read and Write permission to temporary kubeconfig file
TMP_DIR=$(mktemp -d)
cp $KUBECONFIG $TMP_DIR/kubeconfig
chmod 640 $TMP_DIR/kubeconfig
export KUBECONFIG=$TMP_DIR/kubeconfig

# login as kube:admin
oc login -u kubeadmin -p $KUBEADMIN_PASSWORD

# Check login user name for debugging purpose
oc whoami
login_user=`oc whoami`
if [[ $login_user == *"admin"* ]]; then
    echo "Login to the cluster as a admin user"
else
    echo "Fail to login as a admin user"
    exit 1
fi

# assert that kam is on the path
kam version

# Run e2e test
make e2e
