package operator

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	contractIAVSDirectory "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAVSDirectory"
	"github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli/v2"
)

type deregisterConfig struct {
	operatorSetIds  []uint32
	avsAddress      gethcommon.Address
	operatorAddress gethcommon.Address
	network         string
	signerConfig    *types.SignerConfig
	chainId         *big.Int
	broadcast       bool
	ethClient       *ethclient.Client
	outputType      string
	outputFile      string
}

func ForceDeregister(p utils.Prompter) *cli.Command {
	return &cli.Command{
		Name:      "force-deregister",
		Usage:     "Force deregisters operator from operator sets",
		UsageText: "deregister [flags] <avs-address> <operator-set-ids>",
		Description: `
Force deregisters operator sets. This can be use to deregister from any operator set in the AVS if in case AVS does not provide a way to deregister. This does not require any signature from the AVS.

<avs-address> is the address of the AVS contract
<operator-set-ids> is a comma-separated list of operator set IDs to deregister from
`,
		Flags: getDeregisterFlags(),
		Action: func(c *cli.Context) error {
			return deregisterOperatorSet(c, p)
		},
		// TODO(shrimalmadhur): Add this flag when we test it
		Hidden: true,
	}
}

func getDeregisterFlags() []cli.Flag {
	baseFlags := []cli.Flag{
		&flags.NetworkFlag,
		&flags.OperatorAddressFlag,
		&flags.VerboseFlag,
		&flags.BroadcastFlag,
		&flags.ETHRpcUrlFlag,
	}

	allFlags := append(baseFlags, flags.GetSignerFlags()...)
	sort.Sort(cli.FlagsByName(allFlags))
	return allFlags
}

func deregisterOperatorSet(c *cli.Context, p utils.Prompter) error {
	ctx := c.Context
	logger := common.GetLogger(c)

	config, err := readAndValidateDeregisterConfig(c, logger)
	if err != nil {
		return eigenSdkUtils.WrapError("failed to read and validate force deregister config. use --help", err)
	}
	c.App.Metadata["network"] = config.chainId.String()
	avsDirectoryAddress, err := common.GetAVSDirectoryAddress(*config.chainId)
	if err != nil {
		return err
	}
	if config.broadcast {
		if config.signerConfig == nil {
			return errors.New("signerConfig is required for broadcasting")
		}
		logger.Info("Broadcasting claim...")
		keyWallet, sender, err := common.GetWallet(
			*config.signerConfig,
			config.operatorAddress.String(),
			config.ethClient,
			p,
			*config.chainId,
			logger,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to get wallet", err)
		}

		txMgr := txmgr.NewSimpleTxManager(keyWallet, config.ethClient, logger, sender)
		noopMetrics := eigenMetrics.NewNoopMetrics()
		eLWriter, err := elcontracts.NewWriterFromConfig(
			elcontracts.Config{
				AvsDirectoryAddress: gethcommon.HexToAddress(avsDirectoryAddress),
			},
			config.ethClient,
			logger,
			noopMetrics,
			txMgr,
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to create new writer from config", err)
		}

		// TODO(shrimalmadhur): contractIAVSDirectory.ISignatureUtilsSignatureWithSaltAndExpiry{} is a placeholder for
		// the signature right now we are not using force deregister via signature but in future we will support it so
		// that someone can force deregister on behalf of an operator
		receipt, err := eLWriter.ForceDeregisterFromOperatorSets(
			ctx,
			config.operatorAddress,
			config.avsAddress,
			config.operatorSetIds,
			contractIAVSDirectory.ISignatureUtilsSignatureWithSaltAndExpiry{},
		)
		if err != nil {
			return eigenSdkUtils.WrapError("failed to submit force deregister transaction", err)
		}

		logger.Infof("Force deregiister transaction submitted successfully")
		common.PrintTransactionInfo(receipt.TxHash.String(), config.chainId)
	} else {
		if config.outputType == string(common.OutputType_Calldata) {
			noSendTxOpts := common.GetNoSendTxOpts(config.operatorAddress)
			_, _, contractBindings, err := elcontracts.BuildClients(elcontracts.Config{
				AvsDirectoryAddress: gethcommon.HexToAddress(avsDirectoryAddress),
			}, config.ethClient, nil, logger, nil)
			if err != nil {
				return err
			}

			unsignedTx, err := contractBindings.AvsDirectory.ForceDeregisterFromOperatorSets(
				noSendTxOpts,
				config.operatorAddress,
				config.avsAddress,
				config.operatorSetIds,
				contractIAVSDirectory.ISignatureUtilsSignatureWithSaltAndExpiry{},
			)
			if err != nil {
				return err
			}

			calldataHex := gethcommon.Bytes2Hex(unsignedTx.Data())
			if !common.IsEmptyString(config.outputFile) {
				err = common.WriteToFile([]byte(calldataHex), config.outputFile)
				if err != nil {
					return err
				}
				logger.Infof("Call data written to file: %s", config.outputFile)
			} else {
				fmt.Println(calldataHex)
			}
		} else {
			fmt.Println("Force Deregister Operator Set")

			fmt.Println("Operator Address (required if operator is not the sender):", config.operatorAddress.String())
			fmt.Println("AVS Address:", config.avsAddress.String())
			// Convert uint32 to strings
			stringSlice := make([]string, len(config.operatorSetIds))
			for i, num := range config.operatorSetIds {
				stringSlice[i] = fmt.Sprint(num)
			}
			fmt.Println("Operator Set IDs:", strings.Join(stringSlice, ","))
			fmt.Println("To broadcast the force deregister, use the --broadcast flag")
		}
	}
	return nil
}

func readAndValidateDeregisterConfig(c *cli.Context, logger logging.Logger) (*deregisterConfig, error) {
	// Read and validate the deregister config
	network := c.String(flags.NetworkFlag.Name)
	chainId := utils.NetworkNameToChainId(network)
	broadcast := c.Bool(flags.BroadcastFlag.Name)

	output := c.String(flags.OutputFileFlag.Name)
	outputType := c.String(flags.OutputTypeFlag.Name)

	ethRpcClient, err := ethclient.Dial(c.String(flags.ETHRpcUrlFlag.Name))
	if err != nil {
		return nil, err
	}

	args := c.Args().Slice()
	if len(args) != 2 {
		return nil, errors.New("invalid number of arguments")
	}
	avsAddress := gethcommon.HexToAddress(c.Args().Get(0))
	operatorSetIds, err := stringToUnit32Array(strings.Split(c.Args().Get(1), ","))
	if err != nil {
		return nil, err
	}

	operatorAddress := gethcommon.HexToAddress(c.String(flags.OperatorAddressFlag.Name))
	signer, err := common.GetSignerConfig(c, logger)
	if err != nil {
		// We don't want to throw error since people can still use it to generate the claim
		// without broadcasting it
		logger.Debugf("Failed to get signerConfig config: %s", err)
	}

	return &deregisterConfig{
		operatorSetIds:  operatorSetIds,
		avsAddress:      avsAddress,
		operatorAddress: operatorAddress,
		network:         network,
		signerConfig:    signer,
		chainId:         chainId,
		broadcast:       broadcast,
		ethClient:       ethRpcClient,
		outputFile:      output,
		outputType:      outputType,
	}, nil
}

func stringToUnit32Array(arr []string) ([]uint32, error) {
	// Convert a string array to an uint32 array
	res := make([]uint32, len(arr))
	for i, v := range arr {
		vUint, err := parseStringToUint32Array(v)
		if err != nil {
			return nil, err
		}
		res[i] = vUint
	}
	return res, nil
}

func parseStringToUint32Array(s string) (uint32, error) {
	// Convert a string to uint32
	parseUint, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(parseUint), nil
}
