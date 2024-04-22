package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

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
		Action: func(ctx *cli.Context) error {
			skipPrompt := ctx.Bool(YesFlag.Name)

			op := types.OperatorConfigNew{}

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

func promptOperatorInfo(config *types.OperatorConfigNew, p utils.Prompter) (types.OperatorConfigNew, error) {
	// Prompt and set operator address
	operatorAddress, err := p.InputString("Enter your operator address:", "", "",
		func(s string) error {
			return validateAddressIsNonZeroAndValid(s)
		},
	)
	if err != nil {
		return types.OperatorConfigNew{}, err
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

	// Prompt and set earnings address
	earningsAddress, err := p.InputString(
		"Enter your earnings address (default to your operator address):",
		config.Operator.Address,
		"",
		func(s string) error {
			return validateAddressIsNonZeroAndValid(s)
		},
	)
	if err != nil {
		return types.OperatorConfigNew{}, err
	}
	config.Operator.EarningsReceiverAddress = earningsAddress

	// Prompt for eth node
	rpcUrl, err := p.InputString("Enter your ETH rpc url:", "http://localhost:8545", "",
		func(s string) error { return nil },
	)
	if err != nil {
		return types.OperatorConfigNew{}, err
	}
	config.EthRPCUrl = rpcUrl

	// Prompt for ecdsa key path
	ecdsaKeyPath, err := p.InputString("Enter your ecdsa key path:", "", "",
		func(s string) error { return nil },
	)
	if err != nil {
		return types.OperatorConfigNew{}, err
	}
	config.PrivateKeyStorePath = ecdsaKeyPath

	// Prompt for network & set chainId
	chainId, err := p.Select("Select your network:", []string{"mainnet", "holesky", "local"})
	if err != nil {
		return types.OperatorConfigNew{}, err
	}

	switch chainId {
	case "mainnet":
		config.ChainId = *big.NewInt(utils.MainnetChainId)
		config.ELDelegationManagerAddress = utils.ChainMetadataMap[utils.MainnetChainId].ELDelegationManagerAddress
	case "holesky":
		config.ChainId = *big.NewInt(utils.HoleskyChainId)
		config.ELDelegationManagerAddress = utils.ChainMetadataMap[utils.HoleskyChainId].ELDelegationManagerAddress
	case "local":
		config.ChainId = *big.NewInt(utils.LocalChainId)
		config.ELDelegationManagerAddress = utils.ChainMetadataMap[utils.LocalChainId].ELDelegationManagerAddress
	}

	config.SignerType = types.LocalKeystoreSigner

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
