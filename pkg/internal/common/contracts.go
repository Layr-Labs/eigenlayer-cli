package common

import (
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
