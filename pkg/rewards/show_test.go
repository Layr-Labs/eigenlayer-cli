package rewards

import (
	"context"
	"errors"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

// FakeELReader is a mock implementation of the ELReader interface
type FakeELReader struct {
	claimedRewards map[gethcommon.Address]*big.Int
	shouldError    bool
}

func (f *FakeELReader) GetCumulativeClaimed(
	ctx context.Context,
	earnerAddress, tokenAddress gethcommon.Address,
) (*big.Int, error) {
	if f.shouldError {
		return nil, errors.New("mock error")
	}
	return f.claimedRewards[tokenAddress], nil
}
