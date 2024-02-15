![Tests](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/tests.yml/badge.svg)
![Linter](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/golangci-lint.yml/badge.svg)
![Build](https://github.com/Layr-Labs/eigenlayer-cli/actions/workflows/build.yml/badge.svg)

## EigenLayer CLI

EigenLayer CLI is used to manage core operator functionalities like local key management, operator registration and updates.

<!-- TOC -->
* [EigenLayer CLI](#eigenlayer-cli)
  * [Supported Operating Systems](#supported-operating-systems)
  * [Install `eigenlayer` CLI using a binary](#install-eigenlayer-cli-using-a-binary)
    * [Installing in a custom location](#installing-in-a-custom-location)
  * [Install `eigenlayer` CLI using Go](#install-eigenlayer-cli-using-go)
  * [Install `eigenlayer` CLI from source](#install-eigenlayer-cli-from-source)
  * [Documentation](#documentation)
<!-- TOC -->

## Supported Operating Systems
| Operating System | Architecture |
|------------------|--------------|
| Linux            | amd64        |
| Linux            | arm64        |
| Darwin           | amd64        |
| Darwin           | arm64        |

## Update Your Machine
```bash
sudo apt update && sudo apt upgrade -y
```
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

## Installing in a custom location
To download the binary in a custom location, run:
```bash
curl -sSfL https://raw.githubusercontent.com/layr-labs/eigenlayer-cli/master/scripts/install.sh | sh -s -- -b <custom_location>
```

## Install `eigenlayer` CLI using Go

First, install the Go programming language following the [official instructions](https://go.dev/doc/install). You need at least the `1.21` version.

> Eigenlayer is only supported on **Linux**. Make sure you install Go for Linux in a Linux environment (e.g. WSL2, Docker, etc.)
 ## Check Go Version
```bash
go version
```

This command will install the `eigenlayer` executable along with the library and its dependencies in your system:

> As the repository is private, you need to set the `GOPRIVATE` variable properly by running the following command: `export GOPRIVATE=github.com/Layr-Labs/eigenlayer-cli,$GOPRIVATE`. Git will automatically resolve the private access if your Git user has all the required permissions over the repository.

```bash
git clone https://github.com/Layr-Labs/eigenlayer-cli.git && mv eigenlayer-cli eigenlayer
```

The executable will be in your `$GOBIN` (`$GOPATH/bin`).

To check if the `GOBIN` is not in your PATH, you can execute `echo $GOBIN` from the Terminal. If it doesn't print anything, then it is not in your PATH. To add `GOBIN` to your PATH, add the following lines to your `$HOME/.profile`:

```bash
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH
```

> Changes made to a profile file may not apply until the next time you log into your computer. To apply the changes immediately, run the shell commands directly or execute them from the profile using a command such as `source $HOME/.profile`.

## Install `eigenlayer` CLI from source

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

Links to specific sections are provided below : 
* [Create Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#create-keys)
* [Import Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#import-keys)
* [List Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#list-keys)
* [Export Keys](https://docs.eigenlayer.xyz/operator-guides/operator-installation#export-keys)
* [Fund Wallet with ETH](https://docs.eigenlayer.xyz/operator-guides/operator-installation#fund-ecdsa-wallet)
* [Register Operator](https://docs.eigenlayer.xyz/operator-guides/operator-installation#registration)
* [Operator Status](https://docs.eigenlayer.xyz/operator-guides/operator-installation#checking-status-of-registration)
* [Metadata Updates](https://docs.eigenlayer.xyz/operator-guides/operator-installation#metadata-updates)
* [Frequently Asked Questions](https://docs.eigenlayer.xyz/operator-guides/faq)
* [Troubleshooting](https://docs.eigenlayer.xyz/operator-guides/troubleshooting)

If you see any issues in documentation please create an issue or PR [here](https://github.com/Layr-Labs/eigenlayer-docs)
