package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os/user"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/Layr-Labs/eigensdk-go/aws/secretmanager"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
		After: telemetry.AfterRunAction(),
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
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()

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

			keyWallet, sender, err := getWallet(operatorCfg, ethClient, p, logger)
			if err != nil {
				return err
			}

			txMgr := txmgr.NewSimpleTxManager(keyWallet, ethClient, logger, sender)

			noopMetrics := eigenMetrics.NewNoopMetrics()

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
				printRegistrationInfo(
					receipt.TxHash.String(),
					common.HexToAddress(operatorCfg.Operator.Address),
					&operatorCfg.ChainId,
				)
			} else {
				fmt.Printf("%s Operator is already registered on EigenLayer\n", utils.EmojiCheckMark)
				return nil
			}

			fmt.Printf(
				"%s Operator is registered successfully to EigenLayer. There is a 30 minute delay between registration and operator details being shown in our webapp.\n",
				utils.EmojiCheckMark,
			)
			return nil
		},
	}

	return registerCmd
}

func printRegistrationInfo(txHash string, operatorAddress common.Address, chainId *big.Int) {
	fmt.Println(strings.Repeat("-", 100))
	fmt.Printf("%s Chain ID: %s\n", utils.EmojiLink, chainId.String())
	if len(txHash) > 0 {
		fmt.Printf("%s Transaction Link: %s\n", utils.EmojiLink, getTransactionLink(txHash, chainId))
	}

	color.Blue("%s Operator Web App Link: %s\n", utils.EmojiInternet, getWebAppLink(operatorAddress, chainId))
	fmt.Println(strings.Repeat("-", 100))
}

func getWallet(
	cfg *types.OperatorConfig,
	ethClient eth.Client,
	p utils.Prompter,
	logger eigensdkLogger.Logger,
) (wallet.Wallet, common.Address, error) {
	var keyWallet wallet.Wallet
	if cfg.SignerType == types.LocalKeystoreSigner {
		// Check if input is available in the pipe and read the password from it
		ecdsaPassword, readFromPipe := utils.GetStdInPassword()
		var err error
		if !readFromPipe {
			ecdsaPassword, err = p.InputHiddenString("Enter password to decrypt the ecdsa private key:", "",
				func(password string) error {
					return nil
				},
			)
			if err != nil {
				fmt.Println("Error while reading ecdsa key password")
				return nil, common.Address{}, err
			}
		}

		// This is to expand the tilde in the path to the home directory
		// This is not supported by Go's standard library
		keyFullPath, err := expandTilde(cfg.PrivateKeyStorePath)
		if err != nil {
			return nil, common.Address{}, err
		}
		cfg.PrivateKeyStorePath = keyFullPath

		signerCfg := signerv2.Config{
			KeystorePath: cfg.PrivateKeyStorePath,
			Password:     ecdsaPassword,
		}
		sgn, sender, err := signerv2.SignerFromConfig(signerCfg, &cfg.ChainId)
		if err != nil {
			return nil, common.Address{}, err
		}
		keyWallet, err = wallet.NewPrivateKeyWallet(ethClient, sgn, sender, logger)
		if err != nil {
			return nil, common.Address{}, err
		}
		return keyWallet, sender, nil
	} else if cfg.SignerType == types.FireBlocksSigner {
		var secretKey string
		var err error
		switch cfg.FireblocksConfig.SecretStorageType {
		case types.PlainText:
			logger.Info("Using plain text secret storage")
			secretKey = cfg.FireblocksConfig.SecretKey
		case types.AWSSecretManager:
			logger.Info("Using AWS secret manager to get fireblocks secret key")
			secretKey, err = secretmanager.ReadStringFromSecretManager(
				context.Background(),
				cfg.FireblocksConfig.SecretKey,
				cfg.FireblocksConfig.AWSRegion,
			)
			if err != nil {
				return nil, common.Address{}, err
			}
			logger.Infof("Secret key with name %s from region %s read from AWS secret manager",
				cfg.FireblocksConfig.SecretKey,
				cfg.FireblocksConfig.AWSRegion,
			)
		default:
			return nil, common.Address{}, fmt.Errorf("secret storage type %s is not supported",
				cfg.FireblocksConfig.SecretStorageType,
			)
		}
		fireblocksClient, err := fireblocks.NewClient(
			cfg.FireblocksConfig.APIKey,
			[]byte(secretKey),
			cfg.FireblocksConfig.BaseUrl,
			time.Duration(cfg.FireblocksConfig.Timeout)*time.Second,
			logger,
		)
		if err != nil {
			return nil, common.Address{}, err
		}
		keyWallet, err = wallet.NewFireblocksWallet(
			fireblocksClient,
			ethClient,
			cfg.FireblocksConfig.VaultAccountName,
			logger,
		)
		if err != nil {
			return nil, common.Address{}, err
		}
		sender, err := keyWallet.SenderAddress(context.Background())
		if err != nil {
			return nil, common.Address{}, err
		}
		return keyWallet, sender, nil
	} else if cfg.SignerType == types.Web3Signer {
		signerCfg := signerv2.Config{
			Endpoint: cfg.Web3SignerConfig.Url,
			Address:  cfg.Operator.Address,
		}
		sgn, sender, err := signerv2.SignerFromConfig(signerCfg, &cfg.ChainId)
		if err != nil {
			return nil, common.Address{}, err
		}
		keyWallet, err = wallet.NewPrivateKeyWallet(ethClient, sgn, sender, logger)
		if err != nil {
			return nil, common.Address{}, err
		}
		return keyWallet, sender, nil
	} else {
		return nil, common.Address{}, fmt.Errorf("%s signer is not supported", cfg.SignerType)
	}
}

// expandTilde replaces the tilde (~) in the path with the home directory.
func expandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir := usr.HomeDir
		// Replace the first instance of ~ with the home directory
		path = strings.Replace(path, "~", homeDir, 1)
	}
	return path, nil
}

func validateAndReturnConfig(configurationFilePath string) (*types.OperatorConfig, error) {
	operatorCfg, err := readConfigFile(configurationFilePath)
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
		return nil, fmt.Errorf("\r%w: with error %s", ErrInvalidYamlFile, err.Error())
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

func readConfigFile(path string) (*types.OperatorConfig, error) {
	var operatorCfg types.OperatorConfig
	err := utils.ReadYamlConfig(path, &operatorCfg)
	if err != nil {
		return nil, err
	}

	elAVSDirectoryAddress, err := getAVSDirectoryAddress(operatorCfg.ChainId)
	if err != nil {
		return nil, err
	}
	operatorCfg.ELAVSDirectoryAddress = elAVSDirectoryAddress

	elRewardsCoordinatorAddress, err := getRewardCoordinatorAddress(operatorCfg.ChainId)
	if err != nil {
		return nil, err
	}
	operatorCfg.ELRewardsCoordinatorAddress = elRewardsCoordinatorAddress
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

func getRewardCoordinatorAddress(chainID big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := utils.ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELRewardsCoordinatorAddress, nil
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

func getWebAppLink(operatorAddress common.Address, chainId *big.Int) string {
	chainIDInt := chainId.Int64()
	chainMetadata, ok := utils.ChainMetadataMap[chainIDInt]
	if !ok {
		return ""
	} else {
		return fmt.Sprintf("%s/%s", chainMetadata.WebAppUrl, operatorAddress.Hex())
	}
}

func validateMetadata(operatorCfg *types.OperatorConfig) error {
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
