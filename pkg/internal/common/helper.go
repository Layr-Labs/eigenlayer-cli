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
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fatih/color"
)

var ChainMetadataMap = map[int64]types.ChainMetadata{
	MainnetChainId: {
		BlockExplorerUrl:            "https://etherscan.io/tx",
		ELDelegationManagerAddress:  "0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A",
		ELAVSDirectoryAddress:       "0x135dda560e946695d6f155dacafc6f1f25c1f5af",
		ELRewardsCoordinatorAddress: "0x7750d328b314EfFa365A0402CcfD489B80B0adda",
		WebAppUrl:                   "https://app.eigenlayer.xyz/operator",
		ProofStoreBaseURL:           "https://eigenlabs-rewards-mainnet-ethereum.s3.amazonaws.com",
	},
	HoleskyChainId: {
		BlockExplorerUrl:            "https://holesky.etherscan.io/tx",
		ELDelegationManagerAddress:  "0xA44151489861Fe9e3055d95adC98FbD462B948e7",
		ELAVSDirectoryAddress:       "0x055733000064333CaDDbC92763c58BF0192fFeBf",
		ELRewardsCoordinatorAddress: "0xAcc1fb458a1317E886dB376Fc8141540537E68fE",
		WebAppUrl:                   "https://holesky.eigenlayer.xyz/operator",
		ProofStoreBaseURL:           "https://eigenlabs-rewards-testnet-holesky.s3.amazonaws.com",
	},
	AnvilChainId: {
		BlockExplorerUrl:            "",
		ELDelegationManagerAddress:  "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
		ELAVSDirectoryAddress:       "0x0165878A594ca255338adfa4d48449f69242Eb8F",
		ELRewardsCoordinatorAddress: "0x610178dA211FEF7D417bC0e6FeD39F05609AD788",
		WebAppUrl:                   "",
		ProofStoreBaseURL:           "",
	},
}

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

func getWallet(
	cfg types.SignerConfig,
	signerAddress string,
	ethClient *ethclient.Client,
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

	ethClient, err := ethclient.Dial(operatorCfg.EthRPCUrl)
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

	elAVSDirectoryAddress, err := GetAVSDirectoryAddress(&operatorCfg.ChainId)
	if err != nil {
		return nil, err
	}
	operatorCfg.ELAVSDirectoryAddress = elAVSDirectoryAddress

	elRewardsCoordinatorAddress, err := GetRewardCoordinatorAddress(&operatorCfg.ChainId)
	if err != nil {
		return nil, err
	}
	operatorCfg.ELRewardsCoordinatorAddress = elRewardsCoordinatorAddress

	return &operatorCfg, nil
}

func GetRewardCoordinatorAddress(chainID *big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELRewardsCoordinatorAddress, nil
	}
}

func GetAVSDirectoryAddress(chainID *big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELAVSDirectoryAddress, nil
	}
}

func GetDelegationManagerAddress(chainID *big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELDelegationManagerAddress, nil
	}
}

func GetTransactionLink(txHash string, chainId *big.Int) string {
	chainIDInt := chainId.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
	if !ok {
		return txHash
	} else {
		return fmt.Sprintf("%s/%s", chainMetadata.BlockExplorerUrl, txHash)
	}
}

func getWebAppLink(operatorAddress common.Address, chainId *big.Int) string {
	chainIDInt := chainId.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
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
		pk, err := crypto.HexToECDSA(Trim0x(ecdsaPrivateKeyString))
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
	loggerOptions := &eigensdkLogger.SLoggerOptions{
		Level: slog.LevelInfo,
	}
	if verbose {
		loggerOptions = &eigensdkLogger.SLoggerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}
	}
	logger := eigensdkLogger.NewTextSLogger(os.Stdout, loggerOptions)
	return logger
}

func noopSigner(addr common.Address, tx *gethtypes.Transaction) (*gethtypes.Transaction, error) {
	return tx, nil
}

func GetNoSendTxOpts(from common.Address) *bind.TransactOpts {
	return &bind.TransactOpts{
		From:   from,
		Signer: noopSigner,
		NoSend: true,
	}
}

func Trim0x(s string) string {
	return strings.TrimPrefix(s, "0x")
}

func GetEnvFromNetwork(network string) string {
	switch network {
	case utils.HoleskyNetworkName:
		return "testnet"
	case utils.MainnetNetworkName:
		return "mainnet"
	default:
		return "local"
	}
}
