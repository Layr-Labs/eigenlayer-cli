package operator

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/telemetry"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

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
		Flags: []cli.Flag{
			&flags.VerboseFlag,
		},
		Action: func(cCtx *cli.Context) error {
			logger := common.GetLogger(cCtx)
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

			logger.Infof(
				"%s Operator configuration file read successfully %s",
				utils.EmojiCheckMark,
				operatorCfg.Operator.Address,
			)
			logger.Info("%s validating operator config:  %s", utils.EmojiWait, operatorCfg.Operator.Address)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

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

			callOpts := &bind.CallOpts{Context: context.Background()}

			status, err := reader.IsOperatorRegistered(callOpts, operatorCfg.Operator)
			if err != nil {
				return err
			}

			if status {
				fmt.Println()
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
				fmt.Println()
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
	fmt.Printf("Delegation Approver Address: %s\n", operator.DelegationApproverAddress)
	fmt.Printf("Staker Opt Out Window Blocks: %d\n", operator.StakerOptOutWindowBlocks)
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println()
}
