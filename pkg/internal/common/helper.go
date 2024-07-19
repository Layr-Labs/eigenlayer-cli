package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
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
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/fatih/color"
)

func PrintRegistrationInfo(txHash string, operatorAddress common.Address, chainId *big.Int) {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 100))
	PrintTransactionInfo(txHash, chainId)

	color.Blue("%s Operator Web App Link: %s\n", utils.EmojiInternet, getWebAppLink(operatorAddress, chainId))
	fmt.Println(strings.Repeat("-", 100))
}

func PrintTransactionInfo(txHash string, chainId *big.Int) {
	fmt.Printf("%s Chain ID: %s\n", utils.EmojiLink, chainId.String())
	if len(txHash) > 0 {
		fmt.Printf("%s Transaction Link: %s\n", utils.EmojiLink, GetTransactionLink(txHash, chainId))
	}
}

func GetWallet(
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
	} else if cfg.SignerType == types.PrivateKeySigner {
		signerCfg := signerv2.Config{
			PrivateKey: cfg.PrivateKey,
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

func ValidateAndReturnConfig(
	configurationFilePath string,
	logger eigensdkLogger.Logger,
) (*types.OperatorConfig, error) {
	operatorCfg, err := ReadConfigFile(configurationFilePath)
	if err != nil {
		return nil, err
	}
	logger.Infof(
		"%s Operator configuration file read successfully %s",
		utils.EmojiCheckMark,
		operatorCfg.Operator.Address,
	)
	logger.Infof("%s Validating operator config: %s", utils.EmojiWait, operatorCfg.Operator.Address)

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
	logger.Debugf("ELDelegationManagerAddress: %s", operatorCfg.ELDelegationManagerAddress)
	logger.Debugf("ELAVSDirectoryAddress: %s", operatorCfg.ELAVSDirectoryAddress)
	logger.Debugf("ELRewardsCoordinatorAddress: %s", operatorCfg.ELRewardsCoordinatorAddress)

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

func ReadConfigFile(path string) (*types.OperatorConfig, error) {
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

// TODO(shrimalmadhur): remove this and use the utils one in a separate PR
func getRewardCoordinatorAddress(chainID big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := utils.ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELRewardsCoordinatorAddress, nil
	}
}

func GetTransactionLink(txHash string, chainId *big.Int) string {
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

func GetSignerConfig(cCtx *cli.Context, logger eigensdkLogger.Logger) (*types.SignerConfig, error) {
	ecdsaPrivateKeyString := cCtx.String(flags.EcdsaPrivateKeyFlag.Name)
	if !IsEmptyString(ecdsaPrivateKeyString) {
		logger.Debug("Using private key signer")
		pk, err := crypto.HexToECDSA(ecdsaPrivateKeyString)
		if err != nil {
			return nil, err
		}
		return &types.SignerConfig{
			SignerType: types.PrivateKeySigner,
			PrivateKey: pk,
		}, nil
	}

	pathToKeyStore := cCtx.String(flags.PathToKeyStoreFlag.Name)
	if !IsEmptyString(pathToKeyStore) {
		logger.Debug("Using local keystore signer")
		return &types.SignerConfig{
			SignerType:          types.LocalKeystoreSigner,
			PrivateKeyStorePath: pathToKeyStore,
		}, nil
	}

	fireblocksAPIKey := cCtx.String(flags.FireblocksAPIKeyFlag.Name)
	if !IsEmptyString(fireblocksAPIKey) {
		logger.Debug("Using fireblocks signer")
		fireblocksSecretKey := cCtx.String(flags.FireblocksSecretKeyFlag.Name)
		if IsEmptyString(fireblocksSecretKey) {
			return nil, errors.New("fireblocks secret key is required")
		}
		fireblocksVaultAccountName := cCtx.String(flags.FireblocksVaultAccountNameFlag.Name)
		if IsEmptyString(fireblocksVaultAccountName) {
			return nil, errors.New("fireblocks vault account name is required")
		}
		fireblocksBaseUrl := cCtx.String(flags.FireblocksBaseUrlFlag.Name)
		if IsEmptyString(fireblocksBaseUrl) {
			return nil, errors.New("fireblocks base url is required")
		}
		fireblocksTimeout := int64(cCtx.Int(flags.FireblocksTimeoutFlag.Name))
		if fireblocksTimeout <= 0 {
			return nil, errors.New("fireblocks timeout should be greater than 0")
		}
		fireblocksSecretAWSRegion := cCtx.String(flags.FireblocksAWSRegionFlag.Name)
		secretStorageType := cCtx.String(flags.FireblocksSecretStorageTypeFlag.Name)
		if IsEmptyString(secretStorageType) {
			return nil, errors.New("fireblocks secret storage type is required")
		}
		return &types.SignerConfig{
			SignerType: types.FireBlocksSigner,
			FireblocksConfig: types.FireblocksConfig{
				APIKey:            fireblocksAPIKey,
				SecretKey:         fireblocksSecretKey,
				VaultAccountName:  fireblocksVaultAccountName,
				BaseUrl:           fireblocksBaseUrl,
				Timeout:           fireblocksTimeout,
				AWSRegion:         fireblocksSecretAWSRegion,
				SecretStorageType: types.SecretStorageType(secretStorageType),
			},
		}, nil
	}

	we3SignerUrl := cCtx.String(flags.Web3SignerUrlFlag.Name)
	if !IsEmptyString(we3SignerUrl) {
		logger.Debug("Using web3 signer")
		return &types.SignerConfig{
			SignerType: types.Web3Signer,
			Web3SignerConfig: types.Web3SignerConfig{
				Url: we3SignerUrl,
			},
		}, nil
	}

	return nil, fmt.Errorf("supported signer not found, please provide details for signers to use")
}

func IsEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func GetLogger(cCtx *cli.Context) eigensdkLogger.Logger {
	verbose := cCtx.Bool(flags.VerboseFlag.Name)
	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	logger := eigensdkLogger.NewTextSLogger(os.Stdout, &eigensdkLogger.SLoggerOptions{Level: logLevel})
	return logger
}
