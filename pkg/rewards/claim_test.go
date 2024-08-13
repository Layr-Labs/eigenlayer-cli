package rewards

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/proofDataFetcher"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/testutils"

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

func TestGetLatestActivePostedRoot(t *testing.T) {
	now := time.Now().UTC()
	var tests = []struct {
		name                   string
		postedRoots            []*proofDataFetcher.SubmittedRewardRoot
		expectedRootIndex      uint32
		rewardsTimestampString string
	}{
		{
			name: "found an activated root before current time",
			postedRoots: []*proofDataFetcher.SubmittedRewardRoot{
				{
					RootIndex:        1,
					ActivatedAt:      now,
					CalcEndTimestamp: now.Add(-24 * time.Hour),
				},
				{
					RootIndex:        2,
					ActivatedAt:      now.Add(-2 * time.Hour),
					CalcEndTimestamp: now.Add(-48 * time.Hour),
				},
				{
					RootIndex:        3,
					ActivatedAt:      now.Add(1 * time.Hour),
					CalcEndTimestamp: now,
				},
			},
			expectedRootIndex:      1,
			rewardsTimestampString: now.Add(-24 * time.Hour).Format(time.DateOnly),
		},
		{
			name: "found no activated root before current time",
			postedRoots: []*proofDataFetcher.SubmittedRewardRoot{
				{
					RootIndex:        3,
					ActivatedAt:      now.Add(1 * time.Hour),
					CalcEndTimestamp: now,
				},
				{
					RootIndex:        4,
					ActivatedAt:      now.Add(2 * time.Hour),
					CalcEndTimestamp: now.Add(3 * time.Hour),
				},
			},
			expectedRootIndex: 0,
		},
		{
			name: "found an activated root before current time 2",
			postedRoots: []*proofDataFetcher.SubmittedRewardRoot{
				{
					RootIndex:        2,
					ActivatedAt:      now.Add(-2 * time.Hour),
					CalcEndTimestamp: now.Add(-24 * time.Hour),
				},
				{
					RootIndex:        3,
					ActivatedAt:      now.Add(1 * time.Hour),
					CalcEndTimestamp: now.Add(-48 * time.Hour),
				},
			},
			expectedRootIndex:      2,
			rewardsTimestampString: now.Add(-24 * time.Hour).Format(time.DateOnly),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualRewardTimeString, actualRootIndex, err := getLatestActivePostedRoot(tt.postedRoots)
			if tt.expectedRootIndex == 0 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.rewardsTimestampString, actualRewardTimeString)
			}
			assert.Equal(t, tt.expectedRootIndex, actualRootIndex)

		})
	}
}
