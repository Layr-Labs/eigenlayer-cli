package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"gopkg.in/yaml.v2"

	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func RegisterCmd(p utils.Prompter) *cli.Command {
	registerCmd := &cli.Command{
		Name:      "register",
		Usage:     "Register the operator to EigenLayer contracts",
		UsageText: "register <configuration-file>",
		Description: `
		Register command expects a yaml config file as an argument
		to successfully register an operator address to eigenlayer

		This will register operator to DelegationManager
		`,
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			operatorCfg, err := validateAndReturnConfig(configurationFilePath)
			if err != nil {
				return err
			}

			fmt.Printf(
				"\r%s Operator configuration file validated successfully %s\n",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			ctx := context.Background()
			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			// Check if input is available in the pipe and read the password from it
			ecdsaPassword, readFromPipe := utils.GetStdInPassword()
			if !readFromPipe {
				ecdsaPassword, err = p.InputHiddenString("Enter password to decrypt the ecdsa private key:", "",
					func(password string) error {
						return nil
					},
				)
				if err != nil {
					fmt.Println("Error while reading ecdsa key password")
					return err
				}
			}

			signerCfg := signerv2.Config{
				KeystorePath: operatorCfg.PrivateKeyStorePath,
				Password:     ecdsaPassword,
			}
			sgn, sender, err := signerv2.SignerFromConfig(signerCfg, &operatorCfg.ChainId)
			if err != nil {
				return err
			}
			privateKeyWallet, err := wallet.NewPrivateKeyWallet(ethClient, sgn, sender, logger)
			if err != nil {
				return err
			}
			txMgr := txmgr.NewSimpleTxManager(privateKeyWallet, ethClient, logger, sender)

			noopMetrics := metrics.NewNoopMetrics()

			elWriter, err := elContracts.BuildELChainWriter(
				common.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				common.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
				ethClient,
				logger,
				noopMetrics,
				txMgr)

			if err != nil {
				return err
			}

			reader, err := elContracts.BuildELChainReader(
				common.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				common.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			status, err := reader.IsOperatorRegistered(&bind.CallOpts{Context: ctx}, operatorCfg.Operator)
			if err != nil {
				return err
			}

			if !status {
				receipt, err := elWriter.RegisterAsOperator(ctx, operatorCfg.Operator)
				if err != nil {
					fmt.Printf("%s Error while registering operator\n", utils.EmojiCrossMark)
					return err
				}
				fmt.Printf(
					"%s Operator registration transaction at: %s\n",
					utils.EmojiCheckMark,
					getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
				)

			} else {
				fmt.Printf("%s Operator is already registered on EigenLayer\n", utils.EmojiCheckMark)
				return nil
			}

			fmt.Printf("%s Operator is registered successfully to EigenLayer\n", utils.EmojiCheckMark)
			return nil
		},
	}

	return registerCmd
}

func validateAndReturnConfig(configurationFilePath string) (*types.OperatorConfigNew, error) {
	operatorCfg, err := validateAndMigrateConfigFile(configurationFilePath)
	if err != nil {
		return nil, err
	}
	fmt.Printf(
		"%s Operator configuration file read successfully %s\n",
		utils.EmojiCheckMark,
		operatorCfg.Operator.Address,
	)
	fmt.Printf("%s validating operator config: %s", utils.EmojiWait, operatorCfg.Operator.Address)

	err = operatorCfg.Operator.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err.Error())
	}

	err = validateMetadata(operatorCfg)
	if err != nil {
		return nil, err
	}

	if operatorCfg.ELDelegationManagerAddress == "" {
		return nil, fmt.Errorf("\n%w: ELDelegationManagerAddress is not set", ErrInvalidYamlFile)
	}

	ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
	if err != nil {
		return nil, err
	}

	id, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	if id.Cmp(&operatorCfg.ChainId) != 0 {
		return nil, fmt.Errorf(
			"\r%s %w: chain ID in config file %d does not match the chain ID of the network %d",
			utils.EmojiCrossMark,
			ErrInvalidYamlFile,
			&operatorCfg.ChainId,
			id,
		)
	}

	return operatorCfg, nil
}

func validateAndMigrateConfigFile(path string) (*types.OperatorConfigNew, error) {
	operatorCfg := types.OperatorConfigNew{}
	var operatorCfgOld types.OperatorConfig
	err := utils.ReadYamlConfig(path, &operatorCfgOld)
	if err != nil {
		return nil, err
	}
	if operatorCfgOld.ELSlasherAddress != "" || operatorCfgOld.BlsPublicKeyCompendiumAddress != "" {
		fmt.Printf("%s Old config detected, migrating to new config\n", utils.EmojiCheckMark)
		chainIDInt := operatorCfgOld.ChainId.Int64()
		chainMetadata, ok := utils.ChainMetadataMap[chainIDInt]
		if !ok {
			return nil, fmt.Errorf("chain ID %d not supported", chainIDInt)
		}
		operatorCfg = types.OperatorConfigNew{
			Operator:                   operatorCfgOld.Operator,
			ELDelegationManagerAddress: chainMetadata.ELDelegationManagerAddress,
			EthRPCUrl:                  operatorCfgOld.EthRPCUrl,
			PrivateKeyStorePath:        operatorCfgOld.PrivateKeyStorePath,
			SignerType:                 operatorCfgOld.SignerType,
			ChainId:                    operatorCfgOld.ChainId,
		}

		fmt.Printf("%s Backing up old config file to %s", utils.EmojiWait, path+".old")
		err := os.Rename(path, path+".old")
		if err != nil {
			return nil, err
		}
		fmt.Printf("\r%s Old Config file backed up at %s\n", utils.EmojiCheckMark, path+".old")
		fmt.Printf("Writing new config to %s", path)
		yamlData, err := yaml.Marshal(&operatorCfg)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(path, yamlData, 0o644)
		if err != nil {
			return nil, err
		}
		fmt.Printf("\r%s New config file written to %s\n", utils.EmojiCheckMark, path)
	} else {
		err = utils.ReadYamlConfig(path, &operatorCfg)
		if err != nil {
			return nil, err
		}
	}
	elAVSDirectoryAddress, err := getAVSDirectoryAddress(operatorCfg.ChainId)
	if err != nil {
		return nil, err
	}
	operatorCfg.ELAVSDirectoryAddress = elAVSDirectoryAddress
	return &operatorCfg, nil
}

func getAVSDirectoryAddress(chainID big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := utils.ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELAVSDirectoryAddress, nil
	}
}

func getTransactionLink(txHash string, chainId *big.Int) string {
	chainIDInt := chainId.Int64()
	chainMetadata, ok := utils.ChainMetadataMap[chainIDInt]
	if !ok {
		return txHash
	} else {
		return fmt.Sprintf("%s/%s", chainMetadata.BlockExplorerUrl, txHash)
	}
}

func validateMetadata(operatorCfg *types.OperatorConfigNew) error {
	// Raw GitHub URL validation is only for mainnet
	if operatorCfg.ChainId.Cmp(big.NewInt(utils.MainnetChainId)) == 0 {
		err := eigenSdkUtils.ValidateRawGithubUrl(operatorCfg.Operator.MetadataUrl)
		if err != nil {
			return fmt.Errorf("%w: with error %s", ErrInvalidMetadata, err.Error())
		}

		metadataBytes, err := eigenSdkUtils.ReadPublicURL(operatorCfg.Operator.MetadataUrl)
		if err != nil {
			return err
		}

		var metadata *eigensdkTypes.OperatorMetadata
		err = json.Unmarshal(metadataBytes, &metadata)
		if err != nil {
			return fmt.Errorf("%w: unable to parse metadata with error %s", ErrInvalidMetadata, err.Error())
		}

		err = eigenSdkUtils.ValidateRawGithubUrl(metadata.Logo)
		if err != nil {
			return fmt.Errorf("%w: logo url should be valid github raw url, error %s", ErrInvalidMetadata, err.Error())
		}
	}
	return nil
}
