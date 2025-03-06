package txmgr

import (
	"context"
	"fmt"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	common2 "github.com/Layr-Labs/eigenlayer-cli/pkg/internal/common"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/types"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	txmgr2 "github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxManager struct {
	txmgr2.TxManager
	OperatorConfig types.OperatorConfig
	Client         *ethclient.Client
	signer         common2.Signer
	logger         logging.Logger
	dryRun         bool
}

func NewTxManager(
	spec adapters.BaseSpecification,
	cfg adapters.Configuration,
	logger logging.Logger,
	dryRun bool,
) (*TxManager, error) {
	var operatorConfig types.OperatorConfig
	err := cfg.Unmarshal(&operatorConfig)
	if err != nil {
		return nil, err
	}

	var signerConfig types.SignerConfig
	err = cfg.Unmarshal(&signerConfig)
	if err != nil {
		return nil, err
	}

	if spec.RemoteSigning {
		if signerConfig.SignerType != "local_keystore" && signerConfig.SignerType != "fireblocks" {
			return nil, fmt.Errorf("only local_keystore and fireblocks signing supported")
		}
	} else {
		if signerConfig.SignerType != "local_keystore" {
			return nil, fmt.Errorf("only local_keystore signing supported")
		}
	}

	operatorConfig.SignerConfig = signerConfig

	ethClient, err := ethclient.Dial(operatorConfig.EthRPCUrl)
	if err != nil {
		return nil, err
	}

	return &TxManager{
		OperatorConfig: operatorConfig,
		Client:         ethClient,
		logger:         logger,
		dryRun:         dryRun,
	}, nil
}

func (m *TxManager) init() error {
	if m.TxManager == nil {
		keyWallet, sender, signer, err := common2.GetWalletWithSigner(m.OperatorConfig.SignerConfig,
			m.OperatorConfig.Operator.Address,
			m.Client,
			utils.NewPrompter(),
			m.OperatorConfig.ChainId,
			m.logger)

		if err != nil {
			return err
		}

		m.TxManager = NewSimpleTxManager(
			keyWallet,
			m.Client,
			m.logger,
			sender,
			m.OperatorConfig.SignerConfig.SignerType != types.FireBlocksSigner,
		)
		m.signer = signer
	}
	return nil
}

func (m *TxManager) Call(
	ctx context.Context,
	f func(opts *bind.TransactOpts) (*types2.Transaction, error),
) (*types2.Transaction, error) {
	err := m.init()
	if err != nil {
		return nil, err
	}
	opts, err := m.TxManager.GetNoSendTxOpts()
	if err != nil {
		return nil, err
	}
	transaction, err := f(opts)
	if err != nil {
		return nil, err
	}
	_, err = m.Send(ctx, transaction, true)
	return transaction, err
}

func (m *TxManager) CallAndWaitForReceipt(
	ctx context.Context,
	f func(opts *bind.TransactOpts) (*types2.Transaction, error),
) (*types2.Transaction, *types2.Receipt, error) {
	err := m.init()
	if err != nil {
		return nil, nil, err
	}
	opts, err := m.TxManager.GetNoSendTxOpts()
	if err != nil {
		return nil, nil, err
	}
	transaction, err := f(opts)
	if err != nil {
		return nil, nil, err
	}

	if m.dryRun {
		m.logger.Warn("Dry run mode: transaction not sent")
		if err := transaction.EncodeRLP(m); err != nil {
			return nil, nil, err
		}
		return transaction, nil, nil
	}

	receipt, err := m.Send(ctx, transaction, true)
	return transaction, receipt, err
}

// Sign computes a cryptographic signature for the given digest hash and returns the signature and any potential
// error.
func (m *TxManager) Sign(digestHash []byte) ([]byte, error) {
	err := m.init()
	if err != nil {
		return nil, err
	}
	return m.signer.Sign(digestHash)
}

func (m *TxManager) Write(p []byte) (n int, err error) {
	fmt.Println("Raw Transaction (RLP) ===")
	fmt.Printf("%x\n", p)
	fmt.Println("=== Raw Transaction (RLP)")
	return len(p), nil
}
