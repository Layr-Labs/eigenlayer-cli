package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

func CreateCmd(p utils.Prompter) *cli.Command {
	createCmd := &cli.Command{
		Name:      "create",
		Usage:     "Used to create operator config and metadata json sample file",
		UsageText: "create",
		Description: `
		This command is used to create a sample empty operator config file 
		and also an empty metadata json file which you need to upload for
		operator metadata

		Both of these are needed for operator registration
		`,
		Flags: []cli.Flag{
			&YesFlag,
		},
		After: telemetry.AfterRunAction(),
		Action: func(ctx *cli.Context) error {
			skipPrompt := ctx.Bool(YesFlag.Name)

			op := types.OperatorConfig{}

			if !skipPrompt {
				// Prompt user to generate empty or non-empty files
				populate, err := p.Confirm("Would you like to populate the operator config file?")
				if err != nil {
					return err
				}

				if populate {
					op, err = promptOperatorInfo(&op, p)
					if err != nil {
						return err
					}
				}
			}
			yamlData, err := yaml.Marshal(&op)
			if err != nil {
				return err
			}
			operatorFile := "operator.yaml"
			err = os.WriteFile(operatorFile, yamlData, 0o644)
			if err != nil {
				return err
			}

			metadata := eigensdkTypes.OperatorMetadata{}
			jsonData, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				return err
			}

			metadataFile := "metadata.json"
			err = os.WriteFile(metadataFile, jsonData, 0o644)
			if err != nil {
				return err
			}

			fmt.Printf(
				"|| %s Created operator.yaml and metadata.json files. Please make sure configuration details(el_delegation_manager_address) is correct based on network by checking our docs.\n",
				utils.EmojiCheckMark,
			)
			fmt.Println()
			fmt.Println(
				"|| Please fill in the metadata.json file and upload it to a public url. Then update the operator.yaml file with the url (metadata_url).",
			)
			fmt.Printf(
				"|| %s  Make sure to read and adhere to our webapp content policy here %s before registering your operator. Any violation will be taken seriously and could lead to removal of your operator from our UI %s\n",
				utils.EmojiWarning,
				"https://docs.eigenlayer.xyz/eigenlayer/operator-guides/operator-content-guidelines",
				utils.EmojiWarning,
			)
			fmt.Printf(
				"|| %s  Do Not use any EigenLayer brand and logo for your operator. This will result into violation of our content policy  %s\n",
				utils.EmojiWarning,
				utils.EmojiWarning,
			)
			fmt.Println()
			fmt.Println(
				"|| Once you have filled in the operator.yaml file, you can register your operator using the configuration file.",
			)
			return nil
		},
	}

	return createCmd
}

func promptOperatorInfo(config *types.OperatorConfig, p utils.Prompter) (types.OperatorConfig, error) {
	// Prompt and set operator address
	operatorAddress, err := p.InputString("Enter your operator address:", "", "",
		func(s string) error {
			return validateAddressIsNonZeroAndValid(s)
		},
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.Operator.Address = operatorAddress

	// TODO(madhur): Disabling this for now as the feature doesn't work but
	// we need to keep the code around for future
	// Prompt to gate stakers approval
	//gateApproval, err := p.Confirm("Do you want to gate stakers approval?")
	//if err != nil {
	//	return types.OperatorConfig{}, err
	//}
	// Prompt for address if operator wants to gate approvals
	//if gateApproval {
	//	delegationApprover, err := p.InputString("Enter your staker approver address:", "", "",
	//		func(s string) error {
	//			isValidAddress := eigenSdkUtils.IsValidEthereumAddress(s)
	//
	//			if !isValidAddress {
	//				return errors.New("address is invalid")
	//			}
	//
	//			return nil
	//		},
	//	)
	//	if err != nil {
	//		return types.OperatorConfig{}, err
	//	}
	//	config.Operator.DelegationApproverAddress = delegationApprover
	//} else {
	//	config.Operator.DelegationApproverAddress = eigensdkTypes.ZeroAddress
	//}

	// TODO(madhur): Remove this once we have the feature working and want to prompt users for this address
	config.Operator.DelegationApproverAddress = eigensdkTypes.ZeroAddress

	// Prompt for eth node
	rpcUrl, err := p.InputString("Enter your ETH rpc url:", "http://localhost:8545", "",
		func(s string) error { return nil },
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.EthRPCUrl = rpcUrl

	// Prompt for network & set chainId
	chainId, err := p.Select("Select your network:", []string{"mainnet", "holesky", "local"})
	if err != nil {
		return types.OperatorConfig{}, err
	}

	switch chainId {
	case utils.MainnetNetworkName:
		config.ChainId = *big.NewInt(utils.MainnetChainId)
		config.ELDelegationManagerAddress = common.ChainMetadataMap[utils.MainnetChainId].ELDelegationManagerAddress
	case utils.HoleskyNetworkName:
		config.ChainId = *big.NewInt(utils.HoleskyChainId)
		config.ELDelegationManagerAddress = common.ChainMetadataMap[utils.HoleskyChainId].ELDelegationManagerAddress
	case utils.AnvilNetworkName:
		config.ChainId = *big.NewInt(utils.AnvilChainId)
		config.ELDelegationManagerAddress = common.ChainMetadataMap[utils.AnvilChainId].ELDelegationManagerAddress
	}

	// Prompt for signer type
	signerType, err := p.Select("Select your signer type:", []string{"local_keystore", "fireblocks", "web3"})
	if err != nil {
		return types.OperatorConfig{}, err
	}

	switch signerType {
	case "local_keystore":
		config.SignerConfig.SignerType = types.LocalKeystoreSigner
		// Prompt for ecdsa key path
		ecdsaKeyPath, err := p.InputString("Enter your ecdsa key path:", "", "",
			func(s string) error {
				_, err := os.Stat(s)
				if os.IsNotExist(err) {
					return err
				}
				return nil
			},
		)

		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.SignerConfig.PrivateKeyStorePath = ecdsaKeyPath
	case "fireblocks":
		config.SignerConfig.SignerType = types.FireBlocksSigner
		// Prompt for fireblocks API key
		apiKey, err := p.InputString("Enter your fireblocks api key:", "", "",
			func(s string) error {
				if len(s) == 0 {
					return errors.New("fireblocks API key should not be empty")
				}
				return nil
			},
		)
		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.SignerConfig.FireblocksConfig.APIKey = apiKey

		// Prompt for fireblocks base url
		baseUrl, err := p.InputString("Enter your fireblocks base url:", "https://api.fireblocks.io/", "",
			func(s string) error {
				if len(s) == 0 {
					return errors.New("base url should not be empty")
				}
				return nil
			},
		)
		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.SignerConfig.FireblocksConfig.BaseUrl = baseUrl

		// Prompt for fireblocks vault account name
		vaultAccountName, err := p.InputString("Enter the name of fireblocks vault:", "", "",
			func(s string) error {
				if len(s) == 0 {
					return errors.New("vault account name should not be empty")
				}
				return nil
			},
		)
		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.SignerConfig.FireblocksConfig.VaultAccountName = vaultAccountName

		// Prompt for fireblocks API timeout
		timeout, err := p.InputInteger("Enter the timeout for fireblocks API (in seconds):", "3", "",
			func(i int64) error {
				if i <= 0 {
					return errors.New("timeout should be greater than 0")
				}
				return nil
			},
		)
		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.SignerConfig.FireblocksConfig.Timeout = timeout

		// Prompt for fireblocks vault account name
		secretStorageType, err := p.Select(
			"Select your fireblocks secret storage type:",
			[]string{"Plain Text", "AWS Secret Manager"},
		)
		switch secretStorageType {
		case "Plain Text":
			config.SignerConfig.FireblocksConfig.SecretStorageType = types.PlainText
			config.SignerConfig.FireblocksConfig.SecretKey = "<FILL-ME>"
			fmt.Println()
			fmt.Println("Please fill in the secret key in the operator.yaml file")
			fmt.Println()
		case "AWS Secret Manager":
			config.SignerConfig.FireblocksConfig.SecretStorageType = types.AWSSecretManager
			keyName, err := p.InputString("Enter the name of the secret in AWS Secret Manager:", "", "",
				func(s string) error {
					if len(s) == 0 {
						return errors.New("key name should not be empty")
					}
					return nil
				},
			)
			if err != nil {
				return types.OperatorConfig{}, err
			}
			config.SignerConfig.FireblocksConfig.SecretKey = keyName
			awsRegion, err := p.InputString("Enter the AWS region where the secret is stored:", "us-east-1", "",
				func(s string) error {
					if len(s) == 0 {
						return errors.New("AWS region should not be empty")
					}
					return nil
				},
			)
			if err != nil {
				return types.OperatorConfig{}, err
			}
			config.SignerConfig.FireblocksConfig.AWSRegion = awsRegion

		}
		if err != nil {
			return types.OperatorConfig{}, err
		}
	case "web3":
		config.SignerConfig.SignerType = types.Web3Signer
		// Prompt for fireblocks API key
		web3SignerUrl, err := p.InputString("Enter your web3 signer url:", "", "",
			func(s string) error {
				if len(s) == 0 {
					return errors.New("web3 signer should not be empty")
				}
				return nil
			},
		)
		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.SignerConfig.Web3SignerConfig.Url = web3SignerUrl
	default:
		return types.OperatorConfig{}, fmt.Errorf("unknown signer type %s", signerType)
	}

	return *config, nil
}

func validateAddressIsNonZeroAndValid(address string) error {
	if address == eigensdkTypes.ZeroAddress {
		return errors.New("address is 0")
	}

	addressIsValid := eigenSdkUtils.IsValidEthereumAddress(address)

	if !addressIsValid {
		return errors.New("address is invalid")
	}

	return nil
}
