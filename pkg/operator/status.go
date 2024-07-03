package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/urfave/cli/v2"
)

func StatusCmd(p utils.Prompter) *cli.Command {
	statusCmd := &cli.Command{
		Name:      "status",
		Usage:     "Check if the operator is registered and get the operator details",
		UsageText: "status <configuration-file>",
		Description: `
		Check the registration status of operator to EigenLayer.

		It expects the same configuration yaml file as argument to register command	
		`,
		After: telemetry.AfterRunAction(),
		Action: func(cCtx *cli.Context) error {
			args := cCtx.Args()
			if args.Len() != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, args.Len())
			}

			configurationFilePath := args.Get(0)
			operatorCfg, err := common.ReadConfigFile(configurationFilePath)
			if err != nil {
				return err
			}
			cCtx.App.Metadata["network"] = operatorCfg.ChainId.String()

			fmt.Printf(
				"%s Operator configuration file read successfully %s\n",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)
			fmt.Printf("%s validating operator config:  %s", utils.EmojiWait, operatorCfg.Operator.Address)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
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

			fmt.Printf(
				"\r%s Operator configuration file validated successfully %s\n",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)

			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			reader, err := elContracts.BuildELChainReader(
				gethcommon.HexToAddress(operatorCfg.ELDelegationManagerAddress),
				gethcommon.HexToAddress(operatorCfg.ELAVSDirectoryAddress),
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
				fmt.Printf("%s Operator is registered on EigenLayer\n", utils.EmojiCheckMark)
				operatorDetails, err := reader.GetOperatorDetails(callOpts, operatorCfg.Operator)
				if err != nil {
					return err
				}
				printOperatorDetails(operatorDetails)
				common.PrintRegistrationInfo(
					"",
					gethcommon.HexToAddress(operatorCfg.Operator.Address),
					&operatorCfg.ChainId,
				)
			} else {
				fmt.Printf("%s Operator is not registered to EigenLayer\n", utils.EmojiCrossMark)
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
