package rewards

import (
	"flag"
	"os"
	"testing"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/testutils"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"

	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestReadAndValidateConfig_NoRecipientProvided(t *testing.T) {
	earnerAddress := testutils.GenerateRandomEthereumAddressString()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(flags.ETHRpcUrlFlag.Name, "rpc", "")
	fs.String(EarnerAddressFlag.Name, earnerAddress, "")
	fs.String(RewardsCoordinatorAddressFlag.Name, "0x1234", "")
	fs.String(ClaimTimestampFlag.Name, "latest", "")
	fs.String(ProofStoreBaseURLFlag.Name, "dummy-url", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress(earnerAddress), config.RecipientAddress)
}

func TestReadAndValidateConfig_RecipientProvided(t *testing.T) {
	earnerAddress := testutils.GenerateRandomEthereumAddressString()
	recipientAddress := testutils.GenerateRandomEthereumAddressString()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(flags.ETHRpcUrlFlag.Name, "rpc", "")
	fs.String(EarnerAddressFlag.Name, earnerAddress, "")
	fs.String(RecipientAddressFlag.Name, recipientAddress, "")
	fs.String(RewardsCoordinatorAddressFlag.Name, "0x1234", "")
	fs.String(ClaimTimestampFlag.Name, "latest", "")
	fs.String(ProofStoreBaseURLFlag.Name, "dummy-url", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress(recipientAddress), config.RecipientAddress)
}

func TestGetLatestActivePostedRoot
