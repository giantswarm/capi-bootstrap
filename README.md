[![CircleCI](https://circleci.com/gh/giantswarm/capi-bootstrap.svg?style=shield)](https://circleci.com/gh/giantswarm/capi-bootstrap)

# capi-bootstrap

`capi-bootstrap` was written to encapsulate complex logic previously written in bash
for `capo-mc-bootstrap`. It doesn't automate the full process but rather follows the
unix philosophy providing discrete commands which perform specific tasks.

## Design

`capi-bootstrap` was designed to be portable meaning that it can run both on developers'
machines and in a container for CI adapting based on flags and environment variables. It
was also designed to be idempotent wherever possible so that commands should be able to
be run multiple times during development without overwriting existing data.

## Usage

```
Tool for automating the configuration of CAPI management clusters

Usage:
  capi-bootstrap [flags]
  capi-bootstrap [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  config      Commands for managing management cluster app configuration
  help        Help about any command
  key         Commands for managing management cluster encryption keys
  secret      Commands for managing the encryption of Kubernetes secrets
  secrets     Commands for managing installation secrets

Flags:
  -h, --help   help for capi-bootstrap

Use "capi-bootstrap [command] --help" for more information about a command.
```

For a realistic example of using the tool in a pipeline, see `example-script.sh`.

## Environment variables

- SOPS_AGE_KEY - age private key, looks like `AGE-SECRET-KEY-XYZ...`
- SOPS_AGE_KEY_FILE - file containing the same content as `SOPS_AGE_KEY`
- SOPS_AGE_RECIPIENTS - age public key, looks like `age1xyz...`
- LASTPASS_USERNAME - lastpass username
- LASTPASS_PASSWORD - lastpass password
- LASTPASS_TOTP_SECRET - lastpass authenticator secret (can be viewed in Lastpass UI, remove spaces)

## Lastpass

`capi-bootstrap` can access Lastpass in one of two ways. The first way wraps `lastpass-cli` (`cli`)
which is already authenticated in the current user's session. This requires `lpass login` to have
been run before it will work. The second way (`web`) is to send HTTP requests directly to the Lastpass API.
This requires the three `LASTPASS_*` environment variables to be defined so the client can authenticate
automatically. This is intended for use in containers or CI where the `lastpass-cli` session cannot
be used. The second way partially circumvents 2FA and requires storing sensitive credentials in environment
variables which may be able to be accessed by other processes so the first way should be preferred when possible.

## Generators

The tool includes a system for generating arbitrary secrets. Each type of secret is generated by a "generator".
Generators are defined in the `pkg/generator/generators` package. The list of available generators is as follows:

- `awsiam` - Generates new IAM users in AWS with limited permissions for a particular operator.
- `ca` - Generates a new certificate authority (CA).
- `gituboauth` - Generates a new GitHub OAuth app. As there is no API for this, the generator is a wizard which guides 
  the user through creating it in the browser manually. As a result, it can't be used in an automated fashion.
- `lastpass` - Fetches a shared secret from Lastpass from a given secret reference. `format` input determines whether it 
  will be parsed as YAML (for multiple values) or a plain string.
- `taylorbot` - Generates a new GitHub token for the `taylorbot` user. As there is no API for this, the generator is a 
  wizard which guides the user through creating it in the browser manually. As a result, it can't be used in an automated 
  fashion.

All secrets conform to the following format:
```yaml
key: "<key which will be used to refer to the secret in templates>"
generator: "<generator which will generate the secret>"
<generator name>:
  # generator specific inputs
```

Given a secret in Lastpass called "Shared-Team Example/Shared Secrets/Example" with notes containing:

```yaml
key1: value1
key2: value2
```

```yaml
key: example
generator: lastpass
lastpass:
  secretRef:
    share: Shared-Team Example
    group: Shared Secrets
    name: Example
```

`capi-bootstrap secrets generate --encrypt=false` would generate the following Kubernetes secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: installation-secrets
  namespace: giantswarm
stringData:
  example: |-
    key1: value1
    key2: value2
```

## Future work

- (high priority) Integrate with `capo-mc-bootstrap`. Replaces most of `lpass-pull-secrets`, `lpass-update-kubeconfig`, and `setup-config-branch` steps.
- (high priority) Add a `Dockerfile` for `capi-bootstrap` (should include `sops` CLI).
- (medium priority) Add unit tests.
- (low priority) Implement `awsiam` generator and use it for `etcd-backup-operator` and `dns-operator-openstack` credentials.
- (low priority) Find an OAuth solution which doesn't require GitHub OAuth apps and replace `githuboauth` generator with it (see https://gigantic.slack.com/archives/C01BYMF6RN0/p1650962548844109?thread_ts=1650957477.425279&cid=C01BYMF6RN0).
- (low priority) Find an alternative to using Taylorbot tokens (see https://gigantic.slack.com/archives/C01BYMF6RN0/p1650957477425279 and https://github.com/giantswarm/giantswarm/issues/21842).
