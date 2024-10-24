package operator

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/operator/keys"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

func GetApprovalCmd(p utils.Prompter) *cli.Command {
	getApprovalCmd := &cli.Command{
		Name:      "get-approval",
		Usage:     "Generate the smart contract approval details for the delegateTo method",
		UsageText: "get-approval <configuration-file> <staker-address>",
		Description: `
		Generate the smart contract approval details for the delegateTo method.

		It expects the same configuration yaml file as an argument to the register command, along with the staker address.

		Use the --expiry flag to override the default expiration of 3600 seconds.
		`,
		After: telemetry.AfterRunAction(),
		Flags: []cli.Flag{
			&flags.VerboseFlag,
			&flags.ExpiryFlag,
		},
		Action: func(cCtx *cli.Context) error {
			logger := common.GetLogger(cCtx)
			args := cCtx.Args()
			if args.Len() != 2 {
				return fmt.Errorf("%w: accepts 2 arg, received %d", keys.ErrInvalidNumberOfArgs, args.Len())
			}

			expirySeconds := cCtx.Int64(flags.ExpiryFlag.Name)

			configurationFilePath := args.Get(0)
			stakerAddress := args.Get(1)

			if !eigenSdkUtils.IsValidEthereumAddress(stakerAddress) {
				return fmt.Errorf("staker address %s is not valid address", stakerAddress)
			}

			operatorCfg, err := common.ReadConfigFile(configurationFilePath)
			if err != nil {
				return err
			}
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()

			logger.Infof(
				"%s Operator configuration file read successfully %s",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)
			logger.Info("%s validating operator config:  %s", utils.EmojiWait, operatorCfg.Operator.Address)

			ethClient, err := ethclient.Dial(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}
			id, err := ethClient.ChainID(context.Background())
			if err != nil {
				return err
			}

			if id.Cmp(&operatorCfg.ChainId) != 0 {
				return fmt.Errorf(
					"%w: chain ID in config file %d does not match the chain ID of the network %d",
					ErrInvalidYamlFile,
					&operatorCfg.ChainId,
					id,
				)
			}

			logger.Infof(
				"%s Operator configuration file validated successfully %s",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			contractCfg := elcontracts.Config{
				DelegationManagerAddress: gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				AvsDirectoryAddress:      gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
			}
			reader, err := elContracts.NewReaderFromConfig(
				contractCfg,
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			var staker = gethcommon.HexToAddress(stakerAddress)
			var operator = gethcommon.HexToAddress(operatorCfg.Operator.Address)
			var delegationApprover = gethcommon.HexToAddress(operatorCfg.Operator.DelegationApproverAddress)
			salt := make([]byte, 32)

			if _, err := rand.Read(salt); err != nil {
				return err
			}
			var expiry = new(big.Int).SetInt64(time.Now().Unix() + expirySeconds)

			callOpts := &bind.CallOpts{Context: context.Background()}

			hash, err := reader.CalculateDelegationApprovalDigestHash(
				callOpts,
				staker,
				operator,
				delegationApprover,
				[32]byte(salt),
				expiry,
			)
			if err != nil {
				return err
			}

			signed, err := common.Sign(hash[:], operatorCfg.SignerConfig, p)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Println("--------------------------- delegateTo for the staker ---------------------------")
			fmt.Println()
			fmt.Printf("operator: %s\n", operator)
			fmt.Printf("approverSignatureAndExpiry.signature: %s\n", eigenSdkUtils.Add0x(hex.EncodeToString(signed)))
			fmt.Printf("approverSignatureAndExpiry.expiry: %d\n", expiry)
			fmt.Printf("approverSalt: %s\n", eigenSdkUtils.Add0x(hex.EncodeToString(salt)))
			fmt.Println()
			fmt.Println(
				"---------------------------  CalculateDelegationApprovalDigestHash details ---------------------------",
			)
			fmt.Println()
			fmt.Printf("staker: %s\n", staker)
			fmt.Printf("operator: %s\n", operator)
			fmt.Printf("_delegationApprover: %s\n", delegationApprover)
			fmt.Printf("approverSalt: %s\n", eigenSdkUtils.Add0x(hex.EncodeToString(salt)))
			fmt.Printf("expiry: %d\n", expiry)
			fmt.Println()
			fmt.Printf("result: %s\n", eigenSdkUtils.Add0x(hex.EncodeToString(hash[:])))
			fmt.Println()
			fmt.Println("------------------------------------------------------------------------")
			fmt.Println()

			return nil
		},
	}
	return getApprovalCmd
}
