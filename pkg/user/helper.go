package user

import (
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigensdk-go/logging"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func PopulateCallerAddress(
	cliContext *cli.Context,
	logger logging.Logger,
	accountAddress gethcommon.Address,
) gethcommon.Address {
	// TODO: these are copied across both callers of this method. Will clean this up in the CLI refactor of flags.
	callerAddress := cliContext.String(CallerAddressFlag.Name)
	if common.IsEmptyString(callerAddress) {
		logger.Infof(
			"Caller address not provided. Using account address (%s) as caller address",
			accountAddress,
		)

		return accountAddress
	}
	return gethcommon.HexToAddress(callerAddress)
}
