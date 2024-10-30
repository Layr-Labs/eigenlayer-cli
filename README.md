![Tests](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/tests.yml/badge.svg)
![Linter](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/golangci-lint.yml/badge.svg)
![Build](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/build.yml/badge.svg)

# EigenLayer CLI 

EigenLayer CLI is used to interact with EigenLayer core contracts.

<!-- TOC -->
* [EigenLayer CLI](#eigenlayer-cli)
  * [Supported Features](#supported-features)
  * [Supported Key Management Backends](#supported-key-management-backends)
  * [Supported Operating Systems](#supported-operating-systems)
  * [Install `eigenlayer` CLI using a binary](#install-eigenlayer-cli-using-a-binary)
    * [Installing in a custom location](#installing-in-a-custom-location)
  * [Install `eigenlayer` CLI using Go](#install-eigenlayer-cli-using-go)
  * [Install `eigenlayer` CLI from source](#install-eigenlayer-cli-from-source)
  * [Documentation](#documentation)
  * [Release Process](#release-process)
<!-- TOC -->

## Supported Features
* Operator Keys Creation and Management via local keystore (ECDSA and BLS over bn254 curve) - `eigenlayer keys --help`
* Operator Registration, Updates and Status check - `eigenlayer operator --help`
* Reward Claiming and Setting Claimers - `eigenlayer rewards --help`
  * [Detailed Command Documentation](pkg/rewards/README.md)

## Supported Key Management Backends
* Private Key Hex (not recommended for production use)
* [Local Keystore](https://ethereum.org/en/developers/docs/data-structures-and-encoding/web3-secret-storage/)
* [Fireblocks](https://www.fireblocks.com/) backed by AWS KMS for secret management
* [Web3Signer](https://docs.web3signer.consensys.io/)

## Supported Operating Systems
| Operating System | Architecture |
|------------------|--------------|
| Linux            | amd64        |
| Linux            | arm64        |
| Darwin           | amd64        |
| Darwin           | arm64        |


## Install `eigenlayer` CLI using a binary
To download a binary for the latest release, run:
```bash
curl -sSfL https://raw.githubusercontent.com/layr-labs/eigenlayer-cli/master/scripts/install.sh | sh -s
```
The binary will be installed inside the `~/bin` directory.

To add the binary to your path, run:
```bash
export PATH=$PATH:~/bin
```

### Installing in a custom location
To download the binary in a custom location, run:
```bash
curl -sSfL https://raw.githubusercontent.com/layr-labs/eigenlayer-cli/master/scripts/install.sh | sh -s -- -b <custom_location>
```
We collect anonymous usage data to improve the CLI. To disable telemetry, set the environment variable `EIGENLAYER_CLI_TELEMETRY_ENABLED` to `false`.

## Install `eigenlayer` CLI using Go
>Note: Some commands might not work as expected as we use some build time variables. We recommend using [binary installation](#install-eigenlayer-cli-using-a-binary) for best experience.

First, install the Go programming language following the [official instructions](https://go.dev/doc/install). You need at least the `1.21` version.

This command will install the `eigenlayer` executable along with the library and its dependencies in your system:

```bash
go install github.com/Layr-Labs/eigenlayer-cli/cmd/eigenlayer@latest
```

The executable will be in your `$GOBIN` (`$GOPATH/bin`).

To check if the `GOBIN` is not in your PATH, you can execute `echo $GOBIN` from the Terminal. If it doesn't print anything, then it is not in your PATH. To add `GOBIN` to your PATH, add the following lines to your `$HOME/.profile`:

```bash
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
```

> Changes made to a profile file may not apply until the next time you log into your computer. To apply the changes immediately, run the shell commands directly or execute them from the profile using a command such as `source $HOME/.profile`.

## Install `eigenlayer` CLI from source
>Note: Some commands might not work as expected as we use some build time variables. We recommend using [binary installation](#install-eigenlayer-cli-using-a-binary) for best experience.

With this method, you generate the binary manually (need Go installed), downloading and compiling the source code:

```bash
git clone https://github.com/Layr-Labs/eigenlayer-cli.git
cd eigenlayer-cli
mkdir -p build
go build -o build/eigenlayer cmd/eigenlayer/main.go
```

or if you have `make` installed:

```bash
git clone https://github.com/Layr-Labs/eigenlayer-cli.git
cd eigenlayer-cli
make build
```

The executable will be in the `build` folder.

---
In case you want the binary in your PATH (or if you used the [Using Go](#install-eigenlayer-cli-using-go) method and you don't have `$GOBIN` in your PATH), please copy the binary to `/usr/local/bin`:

```bash
# Using Go
sudo cp $GOPATH/bin/eigenlayer /usr/local/bin/

# Build from source
sudo cp eigenlayer-cli/build/eigenlayer /usr/local/bin/
```

## Documentation
Please refer to the full documentation [here](https://docs.eigenlayer.xyz/operator-guides/operator-installation).

Links to specific sections are provided below.
* [Create Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#create-keys)
* [Import Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#import-keys)
* [List Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#list-keys)
* [Export Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#export-keys)
* [Fund Wallet with ETH](https://docs.eigenlayer.xyz/operator-guides/operator-installation#fund-ecdsa-wallet)
* [Register Operator](https://docs.eigenlayer.xyz/operator-guides/operator-installation#registration)
* [Operator Status](https://docs.eigenlayer.xyz/operator-guides/operator-installation#checking-status-of-registration)
* [Metadata Updates](https://docs.eigenlayer.xyz/operator-guides/operator-installation#metadata-updates)
* [Frequently Asked Questions](https://docs.eigenlayer.xyz/operator-guides/operator-faq)
* [Troubleshooting](https://docs.eigenlayer.xyz/operator-guides/troubleshooting)

If you see any issues in documentation please create an issue or PR [here](https://github.com/Layr-Labs/eigenlayer-docs)

## Release Process
To release a new version of the CLI, follow the steps below:
> Note: You need to have write permission to this repo to release new version

1. Checkout the master branch and pull the latest changes:
    ```bash
    git checkout master
    git pull origin master
    ```
2. In your local clone, create a new release tag using the following command:
    ```bash
     git tag v<version> -m "Release v<version>"
    ```
3. Push the tag to the repository using the following command:
    ```bash
    git push origin v<version>
    ```
   
4. This will automatically start the release process in the [GitHub Actions](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/release.yml) and will create a draft release to the [GitHub Releases](https://github.com/Layr-Labs/eigenlayer-cli/releases) with all the required binaries and assets
5. Check the release notes and add any notable changes and publish the release
