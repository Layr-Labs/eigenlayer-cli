package rewards

import (
	"context"
	"errors"
	"flag"
	"math/big"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/testutils"

	rewardscoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IRewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type fakeELReader struct {
	roots []rewardscoordinator.IRewardsCoordinatorDistributionRoot
}

func newFakeELReader(now time.Time) *fakeELReader {
	roots := make([]rewardscoordinator.IRewardsCoordinatorDistributionRoot, 0)
	rootOne := rewardscoordinator.IRewardsCoordinatorDistributionRoot{
		Root:                           [32]byte{0x01},
		RewardsCalculationEndTimestamp: uint32(now.Add(-time.Hour).Unix()),
		ActivatedAt:                    uint32(now.Add(time.Hour).Unix()),
		Disabled:                       false,
	}

	// This is the current claimable distribution root
	rootTwo := rewardscoordinator.IRewardsCoordinatorDistributionRoot{
		Root:                           [32]byte{0x02},
		RewardsCalculationEndTimestamp: uint32(now.Add(48 * -time.Hour).Unix()),
		ActivatedAt:                    uint32(now.Add(-24 * time.Hour).Unix()),
		Disabled:                       false,
	}

	rootThree := rewardscoordinator.IRewardsCoordinatorDistributionRoot{
		Root:                           [32]byte{0x03},
		RewardsCalculationEndTimestamp: uint32(now.Add(32 * -time.Hour).Unix()),
		ActivatedAt:                    uint32(now.Add(-12 * time.Minute).Unix()),
		Disabled:                       true,
	}

	roots = append(roots, rootOne, rootTwo, rootThree)
	// ascending sort order
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ActivatedAt < roots[j].ActivatedAt
	})
	return &fakeELReader{
		roots: roots,
	}
}

func (f *fakeELReader) GetDistributionRootsLength(opts *bind.CallOpts) (*big.Int, error) {
	return big.NewInt(int64(len(f.roots))), nil
}

func (f *fakeELReader) GetRootIndexFromHash(opts *bind.CallOpts, hash [32]byte) (uint32, error) {
	for i, root := range f.roots {
		if root.Root == hash {
			return uint32(i), nil
		}
	}
	return 0, nil
}

func (f *fakeELReader) GetCurrentClaimableDistributionRoot(
	opts *bind.CallOpts,
) (rewardscoordinator.IRewardsCoordinatorDistributionRoot, error) {
	// iterate from end to start since we want the latest active root
	// and the roots are sorted in ascending order of activation time
	for i := len(f.roots) - 1; i >= 0; i-- {
		if !f.roots[i].Disabled && f.roots[i].ActivatedAt < uint32(time.Now().Unix()) {
			return f.roots[i], nil
		}
	}

	return rewardscoordinator.IRewardsCoordinatorDistributionRoot{}, errors.New("no active distribution root found")
}

func (f *fakeELReader) CurrRewardsCalculationEndTimestamp(opts *bind.CallOpts) (uint32, error) {
	rootLen, err := f.GetDistributionRootsLength(opts)
	if err != nil {
		return 0, err
	}
	return f.roots[rootLen.Uint64()-1].RewardsCalculationEndTimestamp, nil
}

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

func TestReadAndValidateConfig_NoClaimerProvided(t *testing.T) {
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
	assert.Equal(t, common.HexToAddress(earnerAddress), config.ClaimerAddress)
}

func TestReadAndValidateConfig_ClaimerProvided(t *testing.T) {
	claimerAddress := testutils.GenerateRandomEthereumAddressString()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(flags.ETHRpcUrlFlag.Name, "rpc", "")
	fs.String(ClaimerAddressFlag.Name, claimerAddress, "")
	fs.String(RewardsCoordinatorAddressFlag.Name, "0x1234", "")
	fs.String(ClaimTimestampFlag.Name, "latest", "")
	fs.String(ProofStoreBaseURLFlag.Name, "dummy-url", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress(claimerAddress), config.ClaimerAddress)
}

func TestGetClaimDistributionRoot(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		claimTimestamp    string
		expectErr         bool
		expectedClaimDate string
		expectedRootIndex uint32
	}{
		{
			name:              "latest root",
			claimTimestamp:    "latest",
			expectErr:         false,
			expectedClaimDate: now.Add(-time.Hour).UTC().Format(time.DateOnly),
			expectedRootIndex: 2,
		},
		{
			name:              "latest active root",
			claimTimestamp:    "latest_active",
			expectErr:         false,
			expectedClaimDate: now.Add(-48 * time.Hour).UTC().Format(time.DateOnly),
			expectedRootIndex: 0,
		},
		{
			name:           "none of them",
			claimTimestamp: "none",
			expectErr:      true,
		},
	}

	reader := newFakeELReader(now)
	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claimDate, rootIndex, err := getClaimDistributionRoot(
				context.Background(),
				tt.claimTimestamp,
				reader,
				logger,
			)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedClaimDate, claimDate)
			assert.Equal(t, tt.expectedRootIndex, rootIndex)
		})
	}
}
