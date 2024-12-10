package common

import (
	"context"
	"errors"
	"math/big"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigenMetrics "github.com/Layr-Labs/eigensdk-go/metrics"
	eigenSdkUtils "github.com/Layr-Labs/eigensdk-go/utils"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetELWriter(
	signerAddress gethcommon.Address,
	signerConfig *types.SignerConfig,
	ethClient *ethclient.Client,
	contractConfig elcontracts.Config,
	prompter utils.Prompter,
	chainId *big.Int,
	logger eigensdkLogger.Logger,
) (*elcontracts.ChainWriter, error) {
	if signerConfig == nil {
		return nil, errors.New("signer is required for broadcasting")
	}
	logger.Debug("Getting Writer from config")
	keyWallet, sender, err := getWallet(
		*signerConfig,
		signerAddress.String(),
		ethClient,
		prompter,
		*chainId,
		logger,
	)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to get wallet", err)
	}

	txMgr := txmgr.NewSimpleTxManager(keyWallet, ethClient, logger, sender)
	noopMetrics := eigenMetrics.NewNoopMetrics()
	eLWriter, err := elcontracts.NewWriterFromConfig(
		contractConfig,
		ethClient,
		logger,
		noopMetrics,
		txMgr,
	)
	if err != nil {
		return nil, eigenSdkUtils.WrapError("failed to create new writer from config", err)
	}

	return eLWriter, nil
}

func IsSmartContractAddress(address gethcommon.Address, ethClient *ethclient.Client) bool {
	code, err := ethClient.CodeAt(context.Background(), address, nil)
	if err != nil {
		// We return true here because we want to treat the address as a smart contract
		// This is only used to gas estimation and creating unsigned transactions
		// So it's fine if eth client return an error
		return true
	}
	return len(code) > 0
}
