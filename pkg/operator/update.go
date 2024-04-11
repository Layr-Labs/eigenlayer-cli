package operator

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/signerv2"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"
)

func UpdateCmd(p utils.Prompter) *cli.Command {
	updateCmd := &cli.Command{
		Name:      "update",
		Usage:     "Update the operator metadata onchain",
		UsageText: "update  <configuration-file> <yubihsm http endpoint> <yubihsm password file> <auth key id> <operator key id> ",
		Description: `
		Updates the operator metadata onchain which includes 
			- metadata url
			- delegation approver address
			- earnings receiver address
			- staker opt out window blocks

		Requires the same file used for registration as argument
		`,
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args()
			if args.Len() != 5 {
				return fmt.Errorf("%w: accepts 5 args, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			operatorCfg, err := validateAndMigrateConfigFile(configurationFilePath)
			if err != nil {
				return err
			}
			fmt.Printf(
				"%s Operator configuration file read successfully %s\n",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

			err = validateMetadata(operatorCfg)
			if err != nil {
				return err
			}

			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			// Arguments for the YubiHSM
			yubihsmEndpoint := args.Get(1)
			yubihsmPasswordFile := args.Get(2)
			yubihsmAuthKeyId, err := strconv.ParseUint(args.Get(3), 10, 16)
			if err != nil {
				return fmt.Errorf("unable to parse auth key id: %s", err.Error())
			}
			yubihsmOperatorKeyId, err := strconv.ParseUint(args.Get(4), 10, 16)
			if err != nil {
				return fmt.Errorf("unable to parse auth key id: %s", err.Error())
			}

			fmt.Printf("%s Connecting to YubiHSM2 at: %s\n", utils.EmojiWait, yubihsmEndpoint)

			yubihsmPasswordBytes, err := os.ReadFile(yubihsmPasswordFile)
			if err != nil || len(yubihsmPasswordBytes) == 0 {
				panic(fmt.Errorf("unable to read password: %s", err.Error()))
			}
			yubiHsmPassword := strings.TrimSpace(string(yubihsmPasswordBytes))

			yubiWallet, err := NewYubihsmWallet(
				yubihsmEndpoint,
				uint16(yubihsmAuthKeyId),
				yubiHsmPassword,
				uint16(yubihsmOperatorKeyId),
				logger,
				ethClient,
			)
			if err != nil {
				return fmt.Errorf("error connecting to yubihsm: %s", err.Error())
			}
			fmt.Printf(
				"\r%s Connected to YubiHSM2 at %s\n",
				utils.EmojiCheckMark,
				yubihsmEndpoint,
			)

			sender, err := yubiWallet.SenderAddress(context.Background())
			if err != nil {
				return fmt.Errorf("error fetching address: %s", err.Error())
			}
			logger.Debugf("eigenlayer operator address will be: %s", sender.String())

			chainID := &operatorCfg.ChainId
			var boundSignerFunc bind.SignerFn = func(address common.Address, tx *gethtypes.Transaction) (*gethtypes.Transaction, error) {
				signer := gethtypes.LatestSignerForChainID(chainID)

				fmt.Println(tx)

				digest := signer.Hash(tx).Bytes()
				signature, err := yubiWallet.Sign(digest, sender)
				if err != nil {
					return nil, fmt.Errorf("error creating signature: %s", err.Error())
				}

				return tx.WithSignature(signer, signature)
			}

			var sgn signerv2.SignerFn = func(ctx context.Context, address common.Address) (bind.SignerFn, error) {
				return boundSignerFunc, nil
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

			receipt, err := elWriter.UpdateOperatorDetails(context.Background(), operatorCfg.Operator)
			if err != nil {
				fmt.Printf("%s Error while updating operator details\n", utils.EmojiCrossMark)
				return err
			}
			fmt.Printf(
				"%s Operator details updated at: %s\n",
				utils.EmojiCheckMark,
				getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId),
			)

			fmt.Printf("%s Operator updated successfully\n", utils.EmojiCheckMark)
			return nil
		},
	}

	return updateCmd
}
