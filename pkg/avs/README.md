# Unified AVS Registration

This package contains an extension to the [EigenLayer CLI](https://github.com/Layr-Labs/eigenlayer-cli) that allows it to support managing registration of operators with different Actively Validated Services (AVS) in a unified manner. It is intended to replace the practice of AVS developers providing an AVS specific tool for the management of operator registration for each AVS.

## Design Objectives

* Allow EigenLayer operators to manage AVS registration workflows (register, opt-in, opt-out, deregister and status check) using the EigenLayer CLI instead of having to use AVS specific tooling.
* Allow AVS developers to provide support for AVS registration workflows by enrolling a specification with the EigenLayer CLI, instead of having to build AVS specific tooling or having to make code changes to the EigenLayer CLI.
* Provide support for remote signers during AVS registration for enhanced security of keys.

## Architecture

The extension design includes the following components.

* A specification manager that stores, updates and loads AVS specifications.
* A configuration provider that loads, overlays and validates configuration parameters specified via the operator configuration and AVS configuration files.
* A built-in coordinator that provide support for managing registration workflows for AVSs built using EigenLayer provided contracts and middleware.
* A built-in coordinator for dynamic contract function invocation that provides support for managing AVS registration workflows using direct or delegated contract function invocation.
* Support for extending coordinators via a plugin framework.
* Transaction signing support via local keystore and remote signers.

### Command Package

The `pkg/avs` command package contains multiple sub commands that perform different functions including management of specifications, generating AVS specific configuration files, and the actual registration workflows (register, opt-in, opt-out, deregister and status check).

These commands are summarized below.

| Command                                                   | Description                                               |
|-----------------------------------------------------------|-----------------------------------------------------------|
| `avs specs list`                                          | List all available AVS specifications.                    |
| `avs specs reset`                                         | Reset the local specification storage to the default.     |
| `avs specs update`                                        | Download the latest set of AVS specifications.            |
| `avs config create <avs-id> <avs-config-file>`            | Generate a configuration template for an AVS.             |
| `avs register <operator-config-file> <avs-config-file>`   | Register an operator with an AVS.                         |
| `avs opt-in <operator-config-file> <avs-config-file>`     | Opt-in an operator with an AVS.                           |
| `avs opt-out <operator-config-file> <avs-config-file>`    | Opt-out an operator with an AVS.                          |
| `avs deregister <operator-config-file> <avs-config-file>` | Deregister an operator with an AVS.                       |
| `avs status <operator-config-file> <avs-config-file>`     | Check the registration status of an operator with an AVS. |

Parameters for above command are as follows.

* `avs-id` refers to a unique identifier for an AVS (see Specifications for more details).
* `operator-config-file` refers to a YALM file that contains operator specific configuration details (also used for operator registration with EigenLayer).
* `avs-config-file` refers to a YAML file that contains AVS specific configuration details.

## Specifications

One of the objectives of the extension design is to be able to support registration workflows on new AVS implementations without changes to the CLI code. This is achieved by describing details about each AVS using a specification, which is a set of JSON (and possibly other) files.

Each specification has a unique identifier for naming it. For example, the identifier for EigenDA specification is `mainnet/eigenda` for the main-net and `holesky/eigenda` for the Holesky test-net. This identifier is used for organizing the specifications as well as in CLI commands when referring to a specification.

Each specification consists of the following.

* An `avs.json` file which describes the details about the AVS and the registration flows.
* An optional set of smart contract ABI definition JSON files which describe the ABI of the contracts with which the CLI is to interact.
* A `config.yaml` file which describes the configuration template to be used during registration workflows with the AVS. The operator must fill-up this template to specify AVS specific configuration parameters.

In addition the specification may also include other files as required.

### Managing Specifications

Specification management includes the following functionality.

* Local storage of specifications and listing them.
* Resetting the local storage of specifications to the default.
* Updating the local storage by downloading the latest set of specifications from a remote location (which is the `master` branch of the GitHub repository for EigenLayer CLI).

### Local Storage of Specifications

All AVS specifications are stored under the user’s home directory, inside the `.eigenlayer/avs/specs` directory. Each specification is stored under a sub-directory that match its unique identifier. CLI commands work off of AVS specifications in this storage.

The CLI extension includes an embedded set of specifications at compile time. On first use of the CLI the local storage would be initialized with this embedded set of specifications if it is empty.

### Listing Specifications

A CLI user will be able to list the available set of AVS specifications by running the `avs specs list` command.

```
$ eigenlayer avs specs list
holesky/altlayer-mach         ALT Layer MACH on Holesky Testnet
holesky/arpa                  ARPA Network on Holesky Testnet
holesky/automata-mp           Automata Multi-Prover AVS on Holesky Testnet
holesky/ava-protocol          AVA Protocol AVS on Holesky Testnet
holesky/brevis-cochain        Brevis coChain AVS on Holesky Testnet
holesky/eigenda               EigenDA on Holesky Testnet
holesky/eoracle               eOracle on Holesky Testnet
holesky/hyperlane             Hyperlane AVS on Holesky Testnet
holesky/lagrange-sc           Lagrange State Committees AVS on Holesky Testnet
holesky/lagrange-zkpn         Lagrange ZK Prover Network AVS on Holesky Testnet
holesky/opacity               Opacity Network AVS on Holesky Testnet
holesky/predicate             Predicate AVS on Holesky Testnet
holesky/skate                 Skate AVS on Holesky Testnet
holesky/witness-chain         Witness Chain on Holesky Testnet
holesky/xterio-mach           Xterio MACH on Holesky Testnet
mainnet/altlayer-mach         ALT Layer MACH
mainnet/arpa                  ARPA Network
mainnet/automata-mp           Automata Multi-Prover AVS
mainnet/ava-protocol          AVA Protocol AVS
mainnet/brevis-cochain        Brevis coChain AVS
mainnet/cyber-mach            Cyber MACH
mainnet/dodochain-mach        DODOchain MACH
mainnet/eigenda               EigenDA
mainnet/eoracle               eOracle
mainnet/gm-network-mach       GM Network MACH
mainnet/hyperlane             Hyperlane AVS
mainnet/lagrange-sc           Lagrange State Committees AVS
mainnet/lagrange-zkpn         Lagrange ZK Prover Network AVS
mainnet/opacity               Opacity Network AVS
mainnet/predicate             Predicate AVS
mainnet/witness-chain         Witness Chain
mainnet/xterio-mach           Xterio MACH
sample/cli-plugin             Sample specification for custom CLI plugin
```

### Resetting Specifications

The local specification storage can be reset to reflect the embedded set of specifications by running the `avs specs reset` command.

```
$ eigenlayer avs specs reset
```

### Updating Specifications

New specification releases done by adding them to the master branch of the GitHub repository of the EigenLayer CLI project can be used as a remote specification repository. A user may update the local specification storage by downloading from this remote repository using the `avs specs update` command.

```
$ eigenlayer avs specs update
Jan  8 21:48:17.958 INF Downloading: https://raw.githubusercontent.com/Layr-Labs/eigenlayer-cli/refs/heads/main/pkg/avs/specs/manifest
Jan  8 21:48:20.328 INF Downloading: sample/cli-plugin/avs.json
Jan  8 21:48:21.180 INF Downloading: sample/cli-plugin/config.yaml
Jan  8 21:48:22.103 INF Downloading: mainnet/eoracle/registry_coordinator.json
Jan  8 21:48:22.922 INF Downloading: mainnet/eoracle/config.yaml
Jan  8 21:48:23.679 INF Downloading: mainnet/eoracle/avs.json
Jan  8 21:48:24.355 INF Downloading: mainnet/eoracle/avs_directory.json
Jan  8 21:48:26.197 INF Downloading: mainnet/eoracle/service_manager.json
Jan  8 21:48:27.016 INF Downloading: mainnet/xterio-mach/config.yaml
Jan  8 21:48:27.840 INF Downloading: mainnet/xterio-mach/avs.json
Jan  8 21:48:28.658 INF Downloading: mainnet/automata-mp/avs.json
Jan  8 21:48:30.326 INF Downloading: mainnet/automata-mp/config.yaml
Jan  8 21:48:31.112 INF Downloading: mainnet/arpa/avs.json
Jan  8 21:48:31.764 INF Downloading: mainnet/arpa/config.yaml
Jan  8 21:48:32.543 INF Downloading: mainnet/arpa/avs_directory.json
Jan  8 21:48:33.788 INF Downloading: mainnet/arpa/node_registry.json
...
```

### Specification Schema

Each AVS specification includes an `avs.json` file that describes the details about the AVS.

```json
{
  “name”: “eigenda”,
  “description”: “EigenDA”,
  “network”: “mainnet”,
  “contract_address”: “0x870679e138bcdf293b7ff14dd44b70fc97e12fc0”,
  “coordinator”: “middleware”,
  “remote_signing”: false
}
```

The following attributes are available at the top level of the AVS schema.

| Attribute        | Optionality | Value Type | Description                                                                                        |
|------------------|-------------|------------|----------------------------------------------------------------------------------------------------|
| name             | Required    | String     | Name of the AVS.                                                                                   |
| description      | Optional    | String     | Description of the AVS.                                                                            |
| network          | Required    | String     | Name of the deployed network.                                                                      |
| contract_address | Required    | Address    | Contract address of the AVS Service Manager.                                                       |
| coordinator      | Required    | String     | Name of the registration flow coordinator which contains the registration workflow implementation. |
| remote_signing   | Optional    | Boolean    | Specify whether remote signing is supported for this AVS. If not specified, it defaults to false.  |

### Coordinators

Each specification indicates how the registration workflows are to be managed in the form of a coordinator. Each coordinator implements a different way of managing registration workflows.

The following coordinators are supported.

| Coordinator | Description                                                                                                                                          |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| contract    | Supports registration workflows via dynamic invocation of contract functions described in the specification for managing the registration workflows. |
| middleware  | Supports registration workflows for AVS implementations that uses the `eigenlayer-contracts` and `eigenlayer-middleware` libraries.                  |
| plugin      | Supports custom registration workflows via coordinator plugins included in the specifications.                                                       |


### Contract Coordinator

The contract coordinator can be used when registration flows involve direct or delegated contract invocation. Details of the contract functions must be included in the AVS specification.

The following is an example specification for Lagrange State Committees that makes use of the contract coordinator.

```json
{
    "name": "holesky/lagrange-sc",
    "description": "Lagrange State Committees AVS on Holesky Testnet",
    "network": "holesky",
    "contract_address": "0x18A74E66cc90F0B1744Da27E72Df338cEa0A542b",
    "coordinator": "contract",
    "remote_signing": true,
    "abi": "service_manager.json",
    "functions": {
        "register": {
            "name": "register",
            "parameters": [
                "config:operator.address",
                "func:bls_sign(type=const:local_keystore,file=config:bls_key_file,password=passwd:bls_key_password,hash=call:committee.calculateKeyWithProofHash,salt=last:salt,expiry=last:expiry)->struct(BlsG1PublicKeys:g1,AggG2PublicKey:g2,Signature:signature,Salt:salt,Expiry:expiry)",
                "func:ecdsa_sign(hash=call:avsDirectory.calculateOperatorAVSRegistrationDigestHash,salt=last:salt,expiry=last:expiry)"
            ]
        },
        "opt-in": {
            "name": "subscribe",
            "parameters": [
                "config:roll_up_chain_id"
            ]
        },
        "opt-out": {
            "name": "unsubscribe",
            "parameters": [
                "config:roll_up_chain_id"
            ]
        },
        "deregister": {
            "name": "deregister",
            "parameters": []
        },
        "status": {
            "name": "avsDirectory.avsOperatorStatus"
        }
    },
    "delegates": [
        {
            "name": "avsDirectory",
            "abi": "avs_directory.json",
            "functions": [
                {
                    "name": "calculateOperatorAVSRegistrationDigestHash",
                    "parameters": [
                        "config:operator.address",
                        "spec:contract_address",
                        "last:salt",
                        "last:expiry"
                    ]
                },
                {
                    "name": "avsOperatorStatus",
                    "parameters": [
                        "spec:contract_address",
                        "config:operator.address"
                    ]
                }
            ]
        },
        {
            "name": "committee",
            "abi": "committee.json",
            "functions": [
                {
                    "name": "calculateKeyWithProofHash",
                    "parameters": [
                        "config:operator.address",
                        "func:salt(seed=const:lagrange-sc)",
                        "func:expiry(timeout=const:300)"
                    ]
                }
            ]
        }
    ]
}
```

The following describes the additional attributes available at the top level of the AVS schema.

| Attribute | Optionality | Value Type                  | Description                                                                  |
|-----------|-------------|-----------------------------|------------------------------------------------------------------------------|
| abi       | Required    | String                      | Filename of the ABI specification for the AVS Service Manager contract.      |
| functions | Required    | Function specification set  | Set of contract functions that should be invoked for registration workflows. |
| delegates | Optional    | Delegate specification list | Set of delegate contracts that contain functions to be invoked.              |

#### Functions

Functions can be defined under the `functions` attribute for the following workflows. If one is not defined the workflow is assumed to be unavailable.

* `register`
* `opt_in`
* `opt_out`
* `deregister`
* `status`

Each function specification under the `functions` attribute describe details about a contract function for an AVS registration workflow, as described below.

| Attribute  | Optionality            | Value Type   | Description                                                                                                                                                          |
|------------|------------------------|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name       | Conditionally Required | String       | If the function is defined, specify the name of the function (or the names of the function chain) that should be invoked on the contract.                            |
| parameters | Conditionally Required | String Array | If the function is defined and requires parameters, specify how the parameter values should be derived in order to be passed to the contract function on invocation. |
| message    | Conditionally Required | String       | If the function is not defined or is implied, specify the message to show on the CLI when invoked.                                                                   |

Function parameters can be specified in the following expression formats.

**Configuration Lookup**

An expression in the form of `config:key` is evaluated to the configuration value (specified in the operator configuration file, or the AVS configuration file, or on the command line) with `key`.

If the configuration value cannot be found, then the user is prompted to enter the value.

For example, `config:operator.address` evaluates to the operator address specified in the operator yaml file under `operator:` and then `address:`.

**Constants**

An expression in the form of `const:value` is evaluated to `value`.

For example, `const:lagrage-avs` evaluates to the string `lagrange-avs` and `const:300` evaluates to `300`.

**Specification Lookup**

An expression in the form of `spec:key` is evaluated to the specification value with `key`.

For example, `spec:contract_address` evaluates to the `contract_address:` defined in the corresponding specification, i.e. `0x18A74E66cc90F0B1744Da27E72Df338cEa0A542b` in the above example.

**Built-in Functions**

An expression in the form of `func:name(params)` is evaluated to the value returned by executing the corresponding built-in function with `name` and parameters.

Parameters to a built-in function invocation is a list of named parameter values. These values may in turn be an expression (but not a nested built-in function invocation).

For example, `func:salt(seed=const:lagrange-sc)` is evaluated by invoking the built-in `salt` function with the parameter `seed` set to the constant value `lagrange-sc`.

See `pkg/avs/adapters/contract/functions.go` for the full list of available built-in functions.

**Contract Functions**

An expression in the form of `call:name` is evaluated to the value returned by invoking the corresponding contract function (or contract function chain) with `name`.

For example, `call:avsDirectory.calculateOperatorAVSRegistrationDigestHash` is evaluated by calling the `avsDirectory` function on the main contract (with the address defined in the specification using `contract_address`), and then invoking the `calculateOperatorAVSRegistrationDigestHash` function on the contract with the resulting address.

**Cached Last Values**

An expression int he form of `last:name` is evaluated to the result of the previous invocation of the function `name`. This is useful when multiple function calls require the same salt and expiry values.

**Transformations**

The result of a function can be transformed using a transformation expression, which can be specified at the function using the `transform` attribute, or at parameter level by adding the delimiter `->` at the end of the parameter expression followed by the transformation expression.

The following example shows how the value returned from a contract invocation can be transformed in to a byte array using a function level transform.

```json
{
    "name": "pubkeyRegistrationMessageHash",
    "parameters": [
        "config:operator.address"
    ],
    "transform": "[]byte"
}
```

The following example shows how the results of a BLS signing can be transformed in to a custom anonymous structure using a parameter level transform.

```json
{
    "name": "register",
    "parameters": [
        "config:operator.address",
        "func:bls_sign(type=const:local_keystore,file=config:bls_key_file,password=passwd:bls_key_password,hash=call:committee.calculateKeyWithProofHash,salt=last:salt,expiry=last:expiry)->struct(BlsG1PublicKeys:g1,AggG2PublicKey:g2,Signature:signature,Salt:salt,Expiry:expiry)",
        "func:ecdsa_sign(hash=call:avsDirectory.calculateOperatorAVSRegistrationDigestHash,salt=last:salt,expiry=last:expiry)"
    ]
}
```

#### Delegates

AVS implementations that delegate operator registration functionality to other contracts can be defined using the `delegates` attribute.

For example, the following defined two delegates.

```json
"delegates": [
    {
        "name": "avsDirectory",
        "abi": "avs_directory.json",
        "functions": [
            {
                "name": "calculateOperatorAVSRegistrationDigestHash",
                "parameters": [
                    "config:operator.address",
                    "spec:contract_address",
                    "last:salt",
                    "last:expiry"
                ]
            },
            {
                "name": "avsOperatorStatus",
                "parameters": [
                    "spec:contract_address",
                    "config:operator.address"
                ]
            }
        ]
    },
    {
        "name": "committee",
        "abi": "committee.json",
        "functions": [
            {
                "name": "calculateKeyWithProofHash",
                "parameters": [
                    "config:operator.address",
                    "func:salt(seed=const:lagrange-sc)",
                    "func:expiry(timeout=const:300)"
                ]
            }
        ]
    }
]
```

Each delegate specification is described with the following attributes.

| Attribute        | Optionality | Value Type                  | Description                                                                                                                                                |
|------------------|-------------|-----------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name             | Required    | String                      | Name of the delegate, which also corresponds to the function on the main contract that fetches the delegate contract address.                              |
| abi              | Required    | String                      | Filename of the ABI specification for the delegate contract.                                                                                               |
| contract_address | Optional    | Address                     | If specified, the specified address is used as the delegate contract address instead of invoking the corresponding contract function on the main contract. |
| functions        | Required    | Function specification list | List of functions available on the delegate.                                                                                                               |

Functions specified at delegate level have the same format as regular contract functions.

### Middleware Coordinator

The middleware coordinator can be used for managing registration processes for an AVS that implements `IServiceManager` which encapsulates `IRegistryCoordinator` for managing operator registrations (as defined in `eigenlayer-contracts` and `eigenlayer-middleware`). EigenDA and Ava Protocol are examples of AVS implementations that follow this pattern.

This specific coordinator includes support for registration (with and without churning as required), deregistration and status checks (with opt-in and opt-out being implied).

The following is an example specification for EigenDA that uses the middleware coordinator.

```json
{
    "name": "holesky/eigenda",
    "description": "EigenDA on Holesky Testnet",
    "network": "holesky",
    "contract_address": "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
    "coordinator": "middleware",
    "remote_signing": true,
    "churner_url": "churner-holesky.eigenda.xyz:443"
}
```

The following describes the additional attributes available at the top level of the AVS schema.

| Attribute   | Optionality | Value Type | Description                                                                 |
|-------------|-------------|------------|-----------------------------------------------------------------------------|
| churner_url | Optional    | URL        | URL of the churner service to use to eject operators when quorums are full. |


### Plugin Coordinator

AVS registration flows that require custom off-chain code execution can include that behavior in to the EigenLayer CLI in the form of a custom coordinator. This can be specified as follows.

```json
{
  ... more,
  “coordinator”: “plugin”,
  ... more,
  “library_url”: “<plugin-so-download-url>”,
  ... plugin specific properties
}
```

The following describes the additional attributes available at the top level of the AVS schema.

| Attribute                  | Optionality | Value Type      | Description                                         |
|----------------------------|-------------|-----------------|-----------------------------------------------------|
| library_url                | Required    | URL             | URL of the plugin library (shared object) download. |
| plugin specific properties | Optional    | JSON Attributes | Set of custom required by the coordinator plugin.   |

The `library_url` can have the following variables that would be substituted to allow derivation of platform specific shared libraries.

* `${ARCH}` represents the architecture as defined in `runtime.GOARCH`
* `${OS}` represents the operating system as defined in `runtime.GOOS`
* `${EXT}` represents the shared library extension based on the operating system (`dll` on Windows, `so` on other operating systems).

For example, the `library_url` may be defined as `https://github.com/avs-x/cli-plugin/releases/download/cli-plugin-1.5/avs-x-cli-plugin-${ARCH}-${OS}.${EXT}`.

More details on building a plugin is available in a later section.

### ABI Schema

Each specification includes a set of files that describes the interface of the contracts that the CLI should interact with for that AVS. These are standard ABI descriptions in JSON format and correspond to the Service Manager contract and other delegated contracts for that AVS.

### Configuration Schema

Commands related to AVS registration makes use of 2 configuration files.

**Operator Configuration**

This is a YAML file that contains operator related details. This is the same file that would have been used to register the operator with EigenLayer.

The following values among others are picked up from this file during AVS registration workflows.

* `operator.address` which specifies the address of the operator.
* `eth_rpc_url` which specifies the RPC URL for the Ethereum node for interaction with the network. The same RPC URL is used for all contract invocations during AVS registration processes.
* `signer_type` which specifies how signing should work (local or remote).
* `private_key_store_path` which specifies the location of the ECDSA private keystore file.
* `fireblocks.*` which specifies details related to Fireblocks based remote signing.
* `web3.*` which specifies details related to Web3 based remote signing.

**AVS Configuration**

This is a YAML file that contains AVS specific details. Each specification includes a template for this as `config.yaml`. A user may copy out this template via the `avs config create` CLI command, edit the resulting file and fill in the details before using it in subsequent CLI commands.

AVS configuration file may only contain a predefined set of parameters required for AVS registration flows, which would be referred to by the CLI to build parameter values for contract function invocation.

All AVS specific configuration values required for the CLI to function can be specified in this configuration file. In addition, values that have been specified in the operator configuration can be overridden by specifying them again in the AVS configuration file.

Any required but unspecified parameters would result in the CLI prompting the user for entry of the missing parameter values as and when they are required before proceeding.

**Command Line Overrides**

Configuration values can be specified and existing values can be overridden on the command line passing `--arg key=value`. Multiple such arguments can be passed.

## Registration Workflows

### Prerequisite - Register with EigenLayer

An operator would first register with EigenLayer by doing the following.

* Create an ECDSA key for the operator.
* Create a BLS key for the operator.
* Prepare the configuration file for the operator.
* Register as an operator with EigenLayer.

The above can be done by using `keys` and `operator` commands available in the CLI.

**Prepare AVS Configuration**

The operator must then prepare a configuration file for registration with required AVS. A template configuration file can be created as follows.

```
$ eigenlayer avs config create <avs-id> [avs-config-file]
```

The following example creates a configuration file for the `holesky/eigenda` AVS called `holesky-eigenda-config.yaml`.

```
$ eigenlayer avs config create holesky/eigenda
```

The operator must then edit this file and populate it with the required configuration values. The file would already have content to name the configuration keys as well as documentation for configuration parameters in the form of comments.

### Register

The operator can register with the AVS as follows.

```
$ eigenlayer avs register <operator-config-file> <avs-config-file>
```

For example, assuming the operator used `operator-config.yaml` for registering the operator with EigenLayer, registration with `holesky/eigenda` can be initiated as;

```
$ eigenlayer avs register operator-config.yaml holesky-eigenda-config.yaml
```

### Opt-In

The operator can opt-in to the AVS as follows.

```
$ eigenlayer avs opt-in <operator-config-file> <avs-config-file>
```

Note that the opt-in process may be implied on some AVS implementations at registration. If so this would be indicated to the operator.

### Opt-Out

The operator can opt-out from the AVS as follows.

```
$ eigenlayer avs opt-out <operator-config-file> <avs-config-file>
```

Note that the opt-out process may be implied on some AVS implementations at deregistration. If so this would be indicated to the operator.

### Deregister

The operator can deregister with the AVS as follows.

```
$ eigenlayer avs deregister <operator-config-file> <avs-config-file>
```

For example, assuming the operator used `operator-config.yaml` for registering the operator with EigenLayer, deregistration with `holesky/eigenda` can be initiated as;

```
$ eigenlayer avs deregister operator-config.yaml holesky-eigenda-config.yaml
```

### Check Registration Status

The operator can check the registration status with an AVS as follows.

```
$ eigenlayer avs status <operator-config-file> <avs-config-file>
```

### Dry Runs

Registration workflows that issue transactions (register, opt-in, opt-out, deregister) can be invoked in a dry-run mode using the `--dry-run` command line argument.

In dry-run mode, the actual transactions are not sent out. Instead, the raw transactions are displayed on the console in RLP encoding.

## Remote Signing

All operations that require signing may be done using one of the following methods.

* Local signing via the keys in the local keystore.
* Remote signing via Fireblocks (only for transaction signing and signature computation for operator ECDSA key).
* Remote signing via Web3Signer (only for transaction signing and signature computation for operator ECDSA key).

Remote signing follows the configuration in the operator configuration file.

## Building Plugins

A minimal plugin for EigenLayer CLI requires a shared library to be built that includes a symbol named `PluginCoordinator` that implement the following interface.

```go
type Coordinator interface {
	Register() error
	OptIn() error
	OptOut() error
	Deregister() error
	Status() (int, error)
}
```

The `Register()`, `OptIn()`, `OptOut()` and `Deregister()` functions are invoked when the plugin is to execute the logic for the corresponding AVS registration workflows.

The `Status()` function is called to query the registration status of an operator for an AVS.

**Create the Project**

Create a new directory to hold the project (following example uses `eigenlayer-cli-plugin-demo`).

```bash
$ mkdir eigenlayer-plugin-demo
$ cd eigenlayer-plugin-demo
```

Initialize the project as a module (following example uses `github.com/eigenlayer/eigenlayer-cli-demo`).

```bash
$ go mod init github.com/eigenlayer/eigenlayer-cli-demo
```

**Add a Coordinator**

Create the plugin coordinator implementation (following is added as `main/coordinator.go`).

```go
package main

import (
    "fmt"
)

type Coordinator struct {
}

func (coordinator Coordinator) Register() error {
    fmt.Println("eigenlayer-cli-demo:register")
    return nil
}

func (coordinator Coordinator) OptIn() error {
    fmt.Println("eigenlayer-cli-demo:opt-in")
    return nil
}

func (coordinator Coordinator) OptOut() error {
    fmt.Println("eigenlayer-cli-demo:opt-out")
    return nil
}

func (coordinator Coordinator) Deregister() error {
    fmt.Println("eigenlayer-cli-demo:deregister")
    return nil
}

func (coordinator Coordinator) Status() (int, error) {
    fmt.Println("eigenlayer-cli-demo:status")
    return 0, nil
}

var PluginCoordinator Coordinator
```

**Build the Plugin**

Build the project as a plugin (following example build `eigenlayer-cli-demo.so`).

```bash
$ go build -buildmode=plugin -o eigenlayer-cli-demo.so
```

**Host the Plugin**

Host the plugin shared library so that it can be accessible from a public URL.

**Use it in a Specification**

The plugin can be used in a specification by setting the specification's `coordinator` and `library_url` attributes.

The following example shows an `avs.json` of a specification that uses a plugin hosted at `https://download.eigenlayer.xys/cli/eigenlayer-cli-demo.so`.

```json
{
    "name": "eigenlayer-cli-demo",
    "network": "mainnet",
    "contract_address": "0x870679e138bcdf293b7ff14dd44b70fc97e12fc0",
    "coordinator": "plugin",
    "remote_signing": false,
    "library_url": "https://download.eigenlayer.xyz/cli/eigenlayer-cli-demo.so"
}
```

### Accessing the Specification

The plugin can access the specification used for launching the coordinator by including a symbol named `PluginSpecification` that implements the following interface.

```go
type Specification interface {
    Validate() error
}
```

The `Validate()` function is called after loading the specification in order to check its validity.

Note that this plugin specification can also include custom properties included in the corresponding `avs.json` file.

For example, consider the following `avs.json`.

```json
{
    "name": "eigenlayer-cli-demo",
    "description": "Specification for CLI Plugin Demo",
    "network": "mainnet",
    "contract_address": "0x870679e138bcdf293b7ff14dd44b70fc97e12fc0",
    "coordinator": "plugin",
    "remote_signing": false,
    "library_url": "https://download.eigenlayer.xys/cli/eigenlayer-cli-demo.so",
    "foo": "bar"
}
```

The plugin can access the specification as follows.

```go
package main

import (
    "errors"
)

type Specification struct {
    Name            string `json:"name"`
    Description     string `json:"description"`
    Network         string `json:"network"`
    ContractAddress string `json:"contract_address"`
    Coordinator     string `json:"coordinator"`
    RemoteSigning   bool   `json:"remote_signing"`
    LibraryURL      string `json:"library_url"`
    Foo             string `json:"foo"`
}

func (spec Specification) Validate() error {
    if spec.Foo == "" {
        return errors.New("specification: foo is required")
    }

    return nil
}

var PluginSpecification Specification
```

### Accessing Configuration Parameters

The plugin can access the configuration parameters used when workflows are invoked by including a symbol named `PluginConfiguration` that implements the following interface.

```go
type Configuration interface {
    Set(key string, value interface{})
}
```

The following is a full example.

```go
package main

type Configuration struct {
	registry map[string]interface{}
}

func (config Configuration) Set(key string, value interface{}) {
	config.registry[key] = value
}

var PluginConfiguration Configuration
```
