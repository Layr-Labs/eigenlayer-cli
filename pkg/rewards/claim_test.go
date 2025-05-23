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

	mocks2 "github.com/Layr-Labs/eigenlayer-cli/pkg/clients/sidecar/mocks"
	"github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/sidecar/v1/rewards"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common/flags"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/internal/testutils"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigenlayer-rewards-proofs/pkg/distribution"

	rewardscoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RewardsCoordinator"
	"github.com/Layr-Labs/eigensdk-go/logging"

	"github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type fakeELReader struct {
	roots                 []rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot
	earnerTokenClaimedMap map[common.Address]map[common.Address]*big.Int
}

func newFakeELReader(
	now time.Time,
	earnerTokenClaimedMap map[common.Address]map[common.Address]*big.Int,
) *fakeELReader {
	roots := make([]rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot, 0)
	rootOne := rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot{
		Root:                           [32]byte{0x01},
		RewardsCalculationEndTimestamp: uint32(now.Add(-time.Hour).Unix()),
		ActivatedAt:                    uint32(now.Add(time.Hour).Unix()),
		Disabled:                       false,
	}

	// This is the current claimable distribution root
	rootTwo := rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot{
		Root:                           [32]byte{0x02},
		RewardsCalculationEndTimestamp: uint32(now.Add(48 * -time.Hour).Unix()),
		ActivatedAt:                    uint32(now.Add(-24 * time.Hour).Unix()),
		Disabled:                       false,
	}

	rootThree := rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot{
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
		roots:                 roots,
		earnerTokenClaimedMap: earnerTokenClaimedMap,
	}
}

func (f *fakeELReader) GetDistributionRootsLength(ctx context.Context) (*big.Int, error) {
	return big.NewInt(int64(len(f.roots))), nil
}

func (f *fakeELReader) GetRootIndexFromHash(ctx context.Context, hash [32]byte) (uint32, error) {
	for i, root := range f.roots {
		if root.Root == hash {
			return uint32(i), nil
		}
	}
	return 0, nil
}

func (f *fakeELReader) GetCurrentClaimableDistributionRoot(
	ctx context.Context,
) (rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot, error) {
	// iterate from end to start since we want the latest active root
	// and the roots are sorted in ascending order of activation time
	for i := len(f.roots) - 1; i >= 0; i-- {
		if !f.roots[i].Disabled && f.roots[i].ActivatedAt < uint32(time.Now().Unix()) {
			return f.roots[i], nil
		}
	}

	return rewardscoordinator.IRewardsCoordinatorTypesDistributionRoot{}, errors.New(
		"no active distribution root found",
	)
}

func (f *fakeELReader) GetCumulativeClaimed(
	ctx context.Context,
	earnerAddress,
	tokenAddress common.Address,
) (*big.Int, error) {
	if f.earnerTokenClaimedMap == nil {
		return big.NewInt(0), nil
	}
	claimed, ok := f.earnerTokenClaimedMap[earnerAddress][tokenAddress]
	if !ok {
		return big.NewInt(0), nil
	}
	return claimed, nil
}

func (f *fakeELReader) CurrRewardsCalculationEndTimestamp(ctx context.Context) (uint32, error) {
	rootLen, err := f.GetDistributionRootsLength(ctx)
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
	fs.String(SidecarUrlFlag.Name, "sidecar", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress(earnerAddress), config.RecipientAddress)
}

func TestReadAndValidateConfig_NoTokenAddressesProvided(t *testing.T) {
	earnerAddress := testutils.GenerateRandomEthereumAddressString()
	recipientAddress := testutils.GenerateRandomEthereumAddressString()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(flags.ETHRpcUrlFlag.Name, "rpc", "")
	fs.String(EarnerAddressFlag.Name, earnerAddress, "")
	fs.String(RecipientAddressFlag.Name, recipientAddress, "")
	fs.String(RewardsCoordinatorAddressFlag.Name, "0x1234", "")
	fs.String(TokenAddressesFlag.Name, "", "")
	fs.String(ClaimTimestampFlag.Name, "latest", "")
	fs.String(SidecarUrlFlag.Name, "sidecar", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.ElementsMatch(t, config.TokenAddresses, []common.Address{})
}

func TestReadAndValidateConfig_ZeroTokenAddressesProvided(t *testing.T) {
	earnerAddress := testutils.GenerateRandomEthereumAddressString()
	recipientAddress := testutils.GenerateRandomEthereumAddressString()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String(flags.ETHRpcUrlFlag.Name, "rpc", "")
	fs.String(EarnerAddressFlag.Name, earnerAddress, "")
	fs.String(RecipientAddressFlag.Name, recipientAddress, "")
	fs.String(RewardsCoordinatorAddressFlag.Name, "0x1234", "")
	fs.String(TokenAddressesFlag.Name, utils.ZeroAddress.String(), "")
	fs.String(ClaimTimestampFlag.Name, "latest", "")
	fs.String(SidecarUrlFlag.Name, "sidecar", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.ElementsMatch(t, config.TokenAddresses, []common.Address{})
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
	fs.String(SidecarUrlFlag.Name, "sidecar", "")
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
	fs.String(SidecarUrlFlag.Name, "sidecar", "")
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
	fs.String(SidecarUrlFlag.Name, "sidecar", "")
	cliCtx := cli.NewContext(nil, fs, nil)

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	config, err := readAndValidateClaimConfig(cliCtx, logger)

	assert.NoError(t, err)
	assert.Equal(t, common.HexToAddress(claimerAddress), config.ClaimerAddress)
}

func parseTimestampToTime(ts string) time.Time {
	s, _ := time.Parse(time.RFC3339, ts)
	return s
}

func TestGetClaimDistributionRoot(t *testing.T) {
	ctrl := gomock.NewController(t)

	client := mocks2.NewMockISidecarClient(ctrl)

	expectedRoots := []*rewards.DistributionRoot{
		{
			Root:      "0x3283c1d2ce88be155961e397782d3318f368b2408fdf43cd58efb5f5ae3015d2",
			RootIndex: 210,
			// fake the end date to be 24 hours in the future
			RewardsCalculationEnd: timestamppb.New(parseTimestampToTime("2025-02-15T00:00:00Z")),
			ActivatedAt:           timestamppb.New(time.Now().Add(24 * time.Hour)),
			BlockHeight:           3371820,
			LogIndex:              43,
		}, {
			Root:                  "0xa3f31747245d8fc3fc8cd589cc6c56d751924e934fda73125774beed6fd6c135",
			RootIndex:             209,
			RewardsCalculationEnd: timestamppb.New(parseTimestampToTime("2025-02-14T00:00:00Z")),
			ActivatedAt:           timestamppb.New(parseTimestampToTime("2025-02-16T19:00:24Z")),
			BlockHeight:           3365501,
			LogIndex:              272,
		}, {
			Root:                  "0x0c7809acafad83b51bc2efa2d8ca71d92d11a2b419761a110d1b9217b1d1bbda",
			RootIndex:             208,
			RewardsCalculationEnd: timestamppb.New(parseTimestampToTime("2025-02-13T00:00:00Z")),
			ActivatedAt:           timestamppb.New(parseTimestampToTime("2025-02-15T19:00:36Z")),
			BlockHeight:           3358860,
			LogIndex:              54,
		},
	}

	tests := []struct {
		name              string
		claimTimestamp    string
		expectErr         bool
		expectedClaimDate string
		expectedRootIndex uint32
		expect            func()
	}{
		{
			name:              "latest root",
			claimTimestamp:    "latest",
			expectErr:         false,
			expectedClaimDate: "2025-02-15",
			expectedRootIndex: 210,
			expect: func() {
				client.EXPECT().
					ListDistributionRoots(gomock.Any(), gomock.Any()).
					Return(&rewards.ListDistributionRootsResponse{
						DistributionRoots: expectedRoots,
					}, nil)
			},
		},
		{
			name:              "latest active root",
			claimTimestamp:    "latest_active",
			expectErr:         false,
			expectedClaimDate: expectedRoots[1].RewardsCalculationEnd.AsTime().UTC().Format(time.DateOnly),
			expectedRootIndex: 209,
			expect: func() {
				client.EXPECT().
					ListDistributionRoots(gomock.Any(), gomock.Any()).
					Return(&rewards.ListDistributionRootsResponse{
						DistributionRoots: expectedRoots,
					}, nil)
			},
		},
		{
			name:           "none of them",
			claimTimestamp: "none",
			expectErr:      true,
			expect: func() {
				client.EXPECT().
					ListDistributionRoots(gomock.Any(), gomock.Any()).
					Return(&rewards.ListDistributionRootsResponse{
						DistributionRoots: []*rewards.DistributionRoot{},
					}, nil)
			},
		},
	}

	logger := logging.NewJsonSLogger(os.Stdout, &logging.SLoggerOptions{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.expect()

			claimDate, rootIndex, _, err := getClaimDistributionRoot(
				context.Background(),
				tt.claimTimestamp,
				logger,
				client,
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

func TestGetTokensToClaim(t *testing.T) {
	// Set up a mock claimableTokens map
	claimableTokens := orderedmap.New[common.Address, *distribution.BigInt]()
	addr1 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	addr2 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	addr3 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())

	claimableTokens.Set(addr1, newBigInt(100))
	claimableTokens.Set(addr2, newBigInt(200))

	// Case 1: No token addresses provided, should return all addresses in claimableTokens
	result := getTokensToClaim(claimableTokens, []common.Address{})
	expected := map[common.Address]*big.Int{
		addr1: big.NewInt(100),
		addr2: big.NewInt(200),
	}
	assert.Equal(t, result, expected)

	// Case 2: Provided token addresses, should return only those present in claimableTokens
	result = getTokensToClaim(claimableTokens, []common.Address{addr2, addr3})
	expected = map[common.Address]*big.Int{
		addr2: big.NewInt(200),
	}
	assert.Equal(t, result, expected)
}

func TestGetTokenAddresses(t *testing.T) {
	// Set up a mock addresses map
	addressesMap := orderedmap.New[common.Address, *distribution.BigInt]()
	addr1 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	addr2 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())

	addressesMap.Set(addr1, newBigInt(100))
	addressesMap.Set(addr2, newBigInt(200))

	// Test that the function returns all addresses in the map
	result := getAllClaimableTokenAddresses(addressesMap)
	expected := map[common.Address]*big.Int{
		addr1: big.NewInt(100),
		addr2: big.NewInt(200),
	}
	assert.Equal(t, result, expected)
}

func TestFilterClaimableTokenAddresses(t *testing.T) {
	// Set up a mock addresses map
	addressesMap := orderedmap.New[common.Address, *distribution.BigInt]()
	addr1 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	addr2 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())

	addressesMap.Set(addr1, newBigInt(100))
	addressesMap.Set(addr2, newBigInt(200))

	// Test filtering with provided addresses
	newMissingAddress := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	providedAddresses := []common.Address{
		addr1,
		newMissingAddress,
	}

	result := filterClaimableTokenAddresses(addressesMap, providedAddresses)
	expected := map[common.Address]*big.Int{
		addr1: big.NewInt(100),
	}
	assert.Equal(t, result, expected)
}

func TestFilterClaimableTokens(t *testing.T) {
	// Set up a mock claimableTokens map
	earnerAddress := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	tokenAddress1 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	tokenAddress2 := common.HexToAddress(testutils.GenerateRandomEthereumAddressString())
	amountClaimed1 := big.NewInt(100)
	amountClaimed2 := big.NewInt(200)
	elReaderClaimedMap := map[common.Address]map[common.Address]*big.Int{
		earnerAddress: {
			tokenAddress1: amountClaimed1,
			tokenAddress2: amountClaimed2,
		},
	}
	now := time.Now()
	reader := newFakeELReader(now, elReaderClaimedMap)
	tests := []struct {
		name                    string
		earnerAddress           common.Address
		claimableTokensMap      map[common.Address]*big.Int
		expectedClaimableTokens []common.Address
	}{
		{
			name:          "all tokens are claimable and non zero",
			earnerAddress: earnerAddress,
			claimableTokensMap: map[common.Address]*big.Int{
				tokenAddress1: big.NewInt(2345),
				tokenAddress2: big.NewInt(3345),
			},
			expectedClaimableTokens: []common.Address{
				tokenAddress1,
				tokenAddress2,
			},
		},
		{
			name:          "one token is already claimed",
			earnerAddress: earnerAddress,
			claimableTokensMap: map[common.Address]*big.Int{
				tokenAddress1: amountClaimed1,
				tokenAddress2: big.NewInt(1234),
			},
			expectedClaimableTokens: []common.Address{
				tokenAddress2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterClaimableTokens(
				context.Background(),
				reader,
				tt.earnerAddress,
				tt.claimableTokensMap,
			)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedClaimableTokens, result)
		})
	}
}

func newBigInt(value int64) *distribution.BigInt {
	return &distribution.BigInt{Int: big.NewInt(value)}
}
