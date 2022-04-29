#!/bin/bash
set -euo pipefail

HELM_VERSION=3.5.0
OPSCTL_VERSION=2.14.0
KIND_VERSION=0.12.0
SOPS_VERSION=3.7.2

GITHUB_TOKEN=$(cat /run/secrets/github_token)
export GITHUB_TOKEN

ASSET_ID=$(curl -H "Authorization: token ${GITHUB_TOKEN}" -sSL https://api.github.com/repos/giantswarm/opsctl/releases/tags/v${OPSCTL_VERSION} | jq -r '.assets[] | select( .name | contains("linux-amd64")) | .id')
export ASSET_ID

curl -H 'Accept: application/octet-stream' -H "Authorization: token ${GITHUB_TOKEN}" -sSL https://api.github.com/repos/giantswarm/opsctl/releases/assets/"${ASSET_ID}"  | tar --strip-components=1 --gzip --extract --file - opsctl-v${OPSCTL_VERSION}-linux-amd64/opsctl
curl --output /binaries/helm -sSL https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz
curl --output /binaries/kind -sSL https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-linux-amd64
curl --output /binaries/sops -sSL https://github.com/mozilla/sops/releases/download/v${SOPS_VERSION}/sops-v{$SOPS_VERSION}.linux.amd64
chmod +x opsctl helm kind sops
