#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

echo "this script is just an example, don't run it"
exit 1

# Set environment variables required by commands
export CLUSTER_NAME=guppy
export BASE_DOMAIN=test.gigantic.io
export PROVIDER=openstack
export INSTALLATION_SECRETS_FILE="installation-secrets.yaml"
export SOPS_CONFIG_FILE=".sops.yaml"

# Generate the sops encryption key and save in Lastpass (or fetch existing) and store
# private and public key returned by the command as environment variables
# (SOPS_AGE_KEY and SOPS_AGE_RECIPIENTS respectively)
ENCRYPTION_KEY_ENV=$(capi-bootstrap key create --cluster-name "$CLUSTER_NAME")
eval "$ENCRYPTION_KEY_ENV"

# Clone management-clusters-fleet-openstack repo and set up branch and directories
git clone git@github.com:giantswarm/management-clusters-fleet-openstack
cd management-clusters-fleet-openstack || exit
git checkout -b "bootstrap-$CLUSTER_NAME"
mkdir -p "clusters/$CLUSTER_NAME"

# Generate installation secrets. These are used as inputs to the following command.
capi-bootstrap secrets generate \
  --base-domain "$BASE_DOMAIN" \
  --cluster-name "$CLUSTER_NAME" \
  --provider "$PROVIDER" \
  > "clusters/$CLUSTER_NAME/$INSTALLATION_SECRETS_FILE"

# Write sops config
capi-bootstrap key write-sops-config > "clusters/$CLUSTER_NAME/SOPS_CONFIG_FILE"

# Commit and push
git add -u
git commit -m "generate installation secrets and sops config for $CLUSTER_NAME"
git push -f
cd .. || exit

# Clone config repo and set up branch and directories
git clone git@github.com:giantswarm/config
cd config || exit
git checkout -b "bootstrap-$CLUSTER_NAME"
mkdir -p "installations/$CLUSTER_NAME"

# Generate management cluster app configuration files which are stored in the config repo.
capi-bootstrap config generate \
  --installation-secrets-file "../management-clusters-fleet-openstack/clusters/$CLUSTER_NAME/$INSTALLATION_SECRETS_FILE" \
  --provider "$PROVIDER" \
  --base-domain "$BASE_DOMAIN" \
  --customer "$CUSTOMER" \
  --cluster-name "$CLUSTER_NAME" \
  --pipeline "$PIPELINE" \
  --output-directory "installations/$CLUSTER_NAME"

# Commit and push
git add -u
git commit -m "generate config for $CLUSTER_NAME"
git push -f
cd "$BASE_DIRECTORY" || exit
