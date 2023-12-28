package operator

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func StatusCmd(p utils.Prompter) *cli.Command {
	statusCmd := &cli.Command{
		Name:      "status",
		Usage:     "Check if the operator is registered and get the operator details",
		UsageText: "status <configuration-file>",
		Description: `
		Check the registration status of operator to Eigenlayer.

		It expects the same configuration yaml file as argument to register command	
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
			fmt.Printf("validating operator config: %s %s\n", operatorCfg.Operator.Address, utils.EmojiWait)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

			fmt.Printf(
				"Operator configuration file validated successfully %s %s\n",
				operatorCfg.Operator.Address,
				utils.EmojiCheckMark,
			)

			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			reader, err := elContracts.BuildELChainReader(
				common.HexToAddress(operatorCfg.ELSlasherAddress),
				common.HexToAddress(operatorCfg.BlsPublicKeyCompendiumAddress),
				ethClient,
				logger,
			)
			if err != nil {
				return err
			}

			callOpts := &bind.CallOpts{Context: context.Background()}

			status, err := reader.IsOperatorRegistered(callOpts, operatorCfg.Operator)
			if err != nil {
				return err
			}

			if status {
				fmt.Printf("Operator is registered on EigenLayer %s\n", utils.EmojiCheckMark)
				operatorDetails, err := reader.GetOperatorDetails(callOpts, operatorCfg.Operator)
				if err != nil {
					return err
				}
				printOperatorDetails(operatorDetails)
				hash, err := reader.GetOperatorPubkeyHash(callOpts, operatorCfg.Operator)
				if err != nil {
					return err
				}
				if hash == [32]byte{} {
					fmt.Printf(
						"Operator BLS pubkey is empty, please run the register command again %s\n",
						utils.EmojiCrossMark,
					)
					return nil
				}
				fmt.Printf("Operator BLS pubkey hash registered on EigenLayer %s\n", utils.EmojiCheckMark)
			} else {
				fmt.Printf("Operator is not registered %s\n", utils.EmojiCrossMark)
			}
			return nil
		},
	}
	return statusCmd
}

func printOperatorDetails(operator eigensdkTypes.Operator) {
	fmt.Println()
	fmt.Println("--------------------------- Operator Details ---------------------------")
	fmt.Printf("Address: %s\n", operator.Address)
	fmt.Printf("Earnings Receiver Address: %s\n", operator.EarningsReceiverAddress)
	fmt.Printf("Delegation Approver Address: %s\n", operator.DelegationApproverAddress)
	fmt.Printf("Staker Opt Out Window Blocks: %d\n", operator.StakerOptOutWindowBlocks)
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println()
}
