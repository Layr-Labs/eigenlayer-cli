package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os/user"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/aws/secretmanager"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/common"

	"github.com/fatih/color"
)

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
	cfg types.SignerConfig,
	signerAddress string,
	ethClient eth.Client,
	p utils.Prompter,
	chainID big.Int,
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
		sgn, sender, err := signerv2.SignerFromConfig(signerCfg, &chainID)
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
			Address:  signerAddress,
		}
		sgn, sender, err := signerv2.SignerFromConfig(signerCfg, &chainID)
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
