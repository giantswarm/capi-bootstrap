[![CircleCI](https://circleci.com/gh/giantswarm/capi-bootstrap.svg?style=shield)](https://circleci.com/gh/giantswarm/capi-bootstrap)

# capi-bootstrap

## commands

- key create --overwrite
- key get --public-key
- key delete --cluster-name
- secret encrypt --public-key
- secret decrypt --private-key
- config render --provider
- bootstrap create --cluster-name --fleet-repo
- bootstrap delete
- bootstrap pivot
- cluster wait-for-ready
- kubeconfig create --cluster-name --ttl 

## Environment variables
- GITHUB_TOKEN
- LASTPASS_USERNAME
- LASTPASS_PASSWORD
- LASTPASS_TOTP_SECRET