# Utility plugin for interacting with Hashicorp Vault service broker

The service broker for Vault provided by Hashicorp (<https://github.com/hashicorp/vault-service-broker>) allows CloudFoundry users to easily create a Vault service instance (like a database) and bind this to running application.

Additionally the service broker sets up shared areas for secrets between applications in the same environment and/or organization.

This plugin makes it easy for developers to invoke the `vault` command using the service keys created by the broker to read/write such values.

## Pre-requisites

1. You have access to a CloudFoundry installation.
1. The Hashicorp Vault service broker is installed and registered in the marketplace.
1. You have the `vault` CLI tool installed.

## Installation

From source:

```bash
go get https://github.com/govau/cf-vault/cmd/cf-vault
cf install-plugin -f $GOPATH/bin/cf-vault
```

Or from a pre-built [release](./releases/):

```bash
cf install-plugin -f https://github.com/govau/cf-vault/releases/download/0.1/cf-vault.osx
```

## Invocation

### Pre-work

```bash
# Create service
cf create-service hashicorp-vault shared my-vault

# Create a key so that we can access the service (the name doesn't matter)
cf create-service-key my-vault my-key

# Show the key
cf service-key my-vault my-key
```

### `cf vault`

Now we use `cf vault` to use the service key created above to interact with our Vault instance. This does the following:

1. Lists the service keys for the selected Vault instance, and picks the first one.
1. Automatically sets `VAULT_ADDR` and `VAULT_TOKEN` per the data issued by the service broker and configured in the service key.
1. Look for any arguments beginning with `cf_i/`, `cf_s/` and `cf_os/` and replace them with the path prefixes for the instance, space or organization (respectively) as configured by the Vault service broker.
1. Execute the `vault` command with remaining arguments.

For example:

```bash
# Write a secret
cf vault my-service write cf_i/email username=foo password=bar

# Read a secret
cf vault my-service read cf_i/email
```

## Building a release

```bash
PLUGIN_PATH=$GOPATH/src/github.com/govau/cf-vault/cmd/cf-vault
PLUGIN_NAME=$(basename $PLUGIN_PATH)
cd $PLUGIN_PATH

GOOS=linux GOARCH=amd64 go build -o ${PLUGIN_NAME}.linux64
GOOS=linux GOARCH=386 go build -o ${PLUGIN_NAME}.linux32
GOOS=windows GOARCH=amd64 go build -o ${PLUGIN_NAME}.win64
GOOS=windows GOARCH=386 go build -o ${PLUGIN_NAME}.win32
GOOS=darwin GOARCH=amd64 go build -o ${PLUGIN_NAME}.osx

shasum -a 1 ${PLUGIN_NAME}.linux64
shasum -a 1 ${PLUGIN_NAME}.linux32
shasum -a 1 ${PLUGIN_NAME}.win64
shasum -a 1 ${PLUGIN_NAME}.win32
shasum -a 1 ${PLUGIN_NAME}.osx
```