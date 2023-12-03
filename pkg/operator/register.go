package operator

import (
	"context"
	"fmt"
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/Layr-Labs/eigensdk-go/signer"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func RegisterCmd(p utils.Prompter) *cli.Command {
	registerCmd := &cli.Command{
		Name:      "register",
		Usage:     "Register the operator and the BLS public key in the EigenLayer contracts",
		UsageText: "register <configuration-file>",
		Description: `
		Register command expects a yaml config file as an argument
		to successfully register an operator address to eigenlayer

		This will register operator to DelegationManager and will register
		the BLS public key on eigenlayer
		`,
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			var operatorCfg types.OperatorConfig
			err := eigensdkUtils.ReadYamlConfig(configurationFilePath, &operatorCfg)
			if err != nil {
				return err
			}
			fmt.Printf(
				"Operator configuration file read successfully %s %s\n",
				operatorCfg.Operator.Address,
				utils.EmojiCheckMark,
			)
			fmt.Printf("validating operator config: %s %s\n", operatorCfg.Operator.Address, utils.EmojiInfo)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

			signerType, err := validateSignerType(operatorCfg)
			if err != nil {
				return err
			}

			fmt.Printf(
				"Operator configuration file validated successfully %s %s\n",
				operatorCfg.Operator.Address,
				utils.EmojiCheckMark,
			)

			ctx := context.Background()
			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			localSigner, err := getSigner(p, signerType, operatorCfg)
			if err != nil {
				return err
			}

			blsKeyPassword, err := p.InputHiddenString("Enter password to decrypt the bls private key:", "",
				func(password string) error {
					return nil
				},
			)
			if err != nil {
				fmt.Println("Error while reading bls key password")
				return err
			}

			keyPair, err := bls.ReadPrivateKeyFromFile(operatorCfg.BlsPrivateKeyStorePath, blsKeyPassword)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			elContractsClient, err := eigenChainio.NewELContractsChainClient(
				common.HexToAddress(operatorCfg.ELSlasherAddress),
				common.HexToAddress(operatorCfg.BlsPublicKeyCompendiumAddress),
				ethClient,
				ethClient,
				logger)
			if err != nil {
				return err
			}

			noopMetrics := metrics.NewNoopMetrics()
			elWriter := elContracts.NewELChainWriter(
				elContractsClient,
				ethClient,
				localSigner,
				logger,
				noopMetrics,
			)
			if err != nil {
				return err
			}

			reader, err := elContracts.NewELChainReader(
				elContractsClient,
				logger,
				ethClient,
			)
			if err != nil {
				return err
			}

			status, err := reader.IsOperatorRegistered(ctx, operatorCfg.Operator)
			if err != nil {
				return err
			}

			if !status {
				receipt, err := elWriter.RegisterAsOperator(ctx, operatorCfg.Operator)
				if err != nil {
					logger.Infof("Error while registering operator %s", utils.EmojiCrossMark)
					return err
				}
				logger.Infof(
					"Operator registration transaction at: %s %s",
					getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
					utils.EmojiCheckMark,
				)

			} else {
				logger.Infof("Operator is already registered on EigenLayer %s\n", utils.EmojiCheckMark)
			}

			receipt, err := elWriter.RegisterBLSPublicKey(ctx, keyPair, operatorCfg.Operator)
			if err != nil {
				logger.Infof("Error while registering BLS public key %s", utils.EmojiCrossMark)
				return err
			}
			logger.Infof(
				"Operator bls key added transaction at: %s %s",
				getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
				utils.EmojiCheckMark,
			)

			logger.Infof("Operator is registered and bls key added successfully %s\n", utils.EmojiCheckMark)
			return nil
		},
	}

	return registerCmd
}

func validateSignerType(operatorCfg types.OperatorConfig) (types.SignerType, error) {
	signerType := string(operatorCfg.SignerType)

	switch signerType {
	case string(types.PrivateKeySigner):
		return types.PrivateKeySigner, nil
	case string(types.LocalKeystoreSigner):
		return types.LocalKeystoreSigner, nil
	default:
		return "", fmt.Errorf("invalid signer type %s", signerType)
	}
}

func getSigner(p utils.Prompter, signerType types.SignerType, operatorCfg types.OperatorConfig) (signer.Signer, error) {
	switch signerType {
	case types.LocalKeystoreSigner:
		fmt.Println("Using local keystore signer")
		ecdsaPassword, err := p.InputHiddenString("Enter password to decrypt the ecdsa private key:", "",
			func(password string) error {
				return nil
			},
		)
		if err != nil {
			fmt.Println("Error while reading ecdsa key password")
			return nil, err
		}
		// TODO: Get chain ID from config
		localSigner, err := signer.NewPrivateKeyFromKeystoreSigner(
			operatorCfg.PrivateKeyStorePath,
			ecdsaPassword,
			&operatorCfg.ChainId,
		)
		if err != nil {
			return nil, err
		}
		return localSigner, nil

	default:
		return nil, fmt.Errorf("invalid signer type %s", signerType)
	}
}

func getTransactionLink(txHash string, chainId *big.Int) string {
	// Create chainId for eth and goerli
	ethChainId := big.NewInt(1)
	goerliChainId := big.NewInt(5)
	holeskyChainId := big.NewInt(17000)

	// Return link of chainId is a live network
	if chainId.Cmp(ethChainId) == 0 {
		return fmt.Sprintf("https://etherscan.io/tx/%s", txHash)
	} else if chainId.Cmp(goerliChainId) == 0 {
		return fmt.Sprintf("https://goerli.etherscan.io/tx/%s", txHash)
	} else if chainId.Cmp(holeskyChainId) == 0 {
		return fmt.Sprintf("https://holesky.etherscan.io/tx/%s", txHash)
	} else {
		return txHash
	}
}
