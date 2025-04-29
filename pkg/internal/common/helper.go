package common

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/aws/secretmanager"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	mainnet             = "mainnet"
	testnet             = "testnet"
	sepolia             = "sepolia"
	hoodi               = "hoodi"
	local               = "local"
	selectorHexIdLength = 10
	addressPrefix       = "0x"
)

var ChainMetadataMap = map[int64]types.ChainMetadata{
	utils.MainnetChainId: {
		BlockExplorerUrl:              utils.MainnetBlockExplorerUrl,
		ELDelegationManagerAddress:    "0x39053D51B77DC0d36036Fc1fCc8Cb819df8Ef37A",
		ELAVSDirectoryAddress:         "0x135dda560e946695d6f155dacafc6f1f25c1f5af",
		ELRewardsCoordinatorAddress:   "0x7750d328b314EfFa365A0402CcfD489B80B0adda",
		ELPermissionControllerAddress: "0x25E5F8B1E7aDf44518d35D5B2271f114e081f0E5",
		WebAppUrl:                     "https://app.eigenlayer.xyz/operator",
		SidecarHttpRpcURL:             "https://sidecar-rpc.eigenlayer.xyz/mainnet",
	},
	utils.HoleskyChainId: {
		BlockExplorerUrl:              utils.HoleskyBlockExplorerUrl,
		ELDelegationManagerAddress:    "0xA44151489861Fe9e3055d95adC98FbD462B948e7",
		ELAVSDirectoryAddress:         "0x055733000064333CaDDbC92763c58BF0192fFeBf",
		ELRewardsCoordinatorAddress:   "0xAcc1fb458a1317E886dB376Fc8141540537E68fE",
		ELPermissionControllerAddress: "0x598cb226B591155F767dA17AfE7A2241a68C5C10",
		WebAppUrl:                     "https://holesky.eigenlayer.xyz/operator",
		SidecarHttpRpcURL:             "https://sidecar-rpc.eigenlayer.xyz/holesky",
	},
	utils.SepoliaChainId: {
		BlockExplorerUrl:              utils.SepoliaBlockExplorerUrl,
		ELDelegationManagerAddress:    "0xD4A7E1Bd8015057293f0D0A557088c286942e84b",
		ELAVSDirectoryAddress:         "0xa789c91ECDdae96865913130B786140Ee17aF545",
		ELRewardsCoordinatorAddress:   "0x5ae8152fb88c26ff9ca5C014c94fca3c68029349",
		ELPermissionControllerAddress: "0x44632dfBdCb6D3E21EF613B0ca8A6A0c618F5a37",
		WebAppUrl:                     "",
		SidecarHttpRpcURL:             "",
	},
	utils.HoodiChainId: {
		BlockExplorerUrl:              utils.HoodiBlockExplorerUrl,
		ELDelegationManagerAddress:    "0x867837a9722C512e0862d8c2E15b8bE220E8b87d",
		ELAVSDirectoryAddress:         "0xD58f6844f79eB1fbd9f7091d05f7cb30d3363926",
		ELRewardsCoordinatorAddress:   "0x29e8572678e0c272350aa0b4B8f304E47EBcd5e7",
		ELPermissionControllerAddress: "0xdcCF401fD121d8C542E96BC1d0078884422aFAD2",
		WebAppUrl:                     "",
		SidecarHttpRpcURL:             "",
	},
	utils.AnvilChainId: {
		BlockExplorerUrl:              "",
		ELDelegationManagerAddress:    "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9",
		ELAVSDirectoryAddress:         "0x0165878A594ca255338adfa4d48449f69242Eb8F",
		ELRewardsCoordinatorAddress:   "0x2279B7A0a67DB372996a5FaB50D91eAA73d2eBe6",
		ELPermissionControllerAddress: "0x3Aa5ebB10DC797CAC828524e59A333d0A371443c",
		WebAppUrl:                     "",
		SidecarHttpRpcURL:             "",
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

func GetPermissionControllerAddress(chainID *big.Int) (string, error) {
	chainIDInt := chainID.Int64()
	chainMetadata, ok := ChainMetadataMap[chainIDInt]
	if !ok {
		return "", fmt.Errorf("chain ID %d not supported", chainIDInt)
	} else {
		return chainMetadata.ELPermissionControllerAddress, nil
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

	if cCtx.Bool(flags.SilentFlag.Name) {
		return eigensdkLogger.NewTextSLogger(io.Discard, nil)
	}

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
	return strings.TrimPrefix(s, addressPrefix)
}

func Sign(digest []byte, cfg types.SignerConfig, p utils.Prompter) ([]byte, error) {
	var privateKey *ecdsa.PrivateKey

	if cfg.SignerType == types.LocalKeystoreSigner {
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
				return nil, err
			}
		}

		jsonContent, err := os.ReadFile(cfg.PrivateKeyStorePath)
		if err != nil {
			return nil, err
		}
		key, err := keystore.DecryptKey(jsonContent, ecdsaPassword)
		if err != nil {
			return nil, err
		}

		privateKey = key.PrivateKey
	} else if cfg.SignerType == types.FireBlocksSigner {
		return nil, errors.New("FireBlocksSigner is not implemented")
	} else if cfg.SignerType == types.Web3Signer {
		return nil, errors.New("Web3Signer is not implemented")
	} else if cfg.SignerType == types.PrivateKeySigner {
		privateKey = cfg.PrivateKey
	} else {
		return nil, errors.New("signer is not implemented")
	}

	signed, err := crypto.Sign(digest, privateKey)
	if err != nil {
		return nil, err
	}

	// account for EIP-155 by incrementing V if necessary
	if signed[crypto.RecoveryIDOffset] < 27 {
		signed[crypto.RecoveryIDOffset] += 27
	}

	return signed, nil
}

func ValidateAndConvertSelectorString(selector string) ([4]byte, error) {
	if len(selector) != selectorHexIdLength || selector[:2] != addressPrefix {
		return [4]byte{}, errors.New("selector must be a 4-byte hex string prefixed with '0x'")
	}

	decoded, err := hex.DecodeString(selector[2:])
	if err != nil {
		return [4]byte{}, eigenSdkUtils.WrapError("invalid hex encoding: %v", err)
	}

	if len(decoded) != 4 {
		return [4]byte{}, fmt.Errorf("decoded selector must be 4 bytes, got %d bytes", len(decoded))
	}

	var selectorBytes [4]byte
	copy(selectorBytes[:], decoded)

	return selectorBytes, nil
}

func PopulateCallerAddress(
	cliContext *cli.Context,
	logger logging.Logger,
	defaultAddress common.Address,
	defaultName string,
) common.Address {
	// TODO: these are copied across both callers of this method. Will clean this up in the CLI refactor of flags.
	callerAddress := cliContext.String(flags.CallerAddressFlag.Name)
	if IsEmptyString(callerAddress) {
		logger.Infof(
			"Caller address not provided. Using %s as default address (%s)",
			defaultName,
			defaultAddress,
		)

		return defaultAddress
	}
	return common.HexToAddress(callerAddress)
}

func GetEnvFromNetwork(network string) string {
	switch network {
	case utils.HoleskyNetworkName:
		return testnet
	case utils.MainnetNetworkName:
		return mainnet
	case utils.SepoliaNetworkName:
		return sepolia
	case utils.HoodiNetworkName:
		return hoodi
	default:
		return local
	}
}
