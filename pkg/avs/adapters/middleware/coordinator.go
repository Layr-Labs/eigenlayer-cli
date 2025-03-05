package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/txmgr"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils/codec/types"
	"github.com/Layr-Labs/eigensdk-go/logging"
)

// CoordinatorConfigOperator defines the operator configurations, including its address, which is required and
// validated.
type CoordinatorConfigOperator struct {
	Address string `json:"address" validate:"required"`
}

// CoordinatorBasicConfig provides basic configuration for the coordinator, including the operator and Ethereum RPC URL.
type CoordinatorBasicConfig struct {
	Operator  CoordinatorConfigOperator `json:"operator"    validate:"required"`
	EthRpcUrl string                    `json:"eth_rpc_url" validate:"required"`
}

// CoordinatorTargetConfig defines the parameters required for a coordinator target, including quorum IDs and chain
// details.
type CoordinatorTargetConfig struct {
	QuorumIDList types.JSONString `json:"quorum_id_list" validate:"required"`
	ChainID      int64            `json:"chain_id"       validate:"required"`
	SignerType   string           `json:"signer_type"`
}

// CoordinatorRegistrationConfig contains the configuration required for coordinator registration.
type CoordinatorRegistrationConfig struct {
	CoordinatorBasicConfig
	CoordinatorTargetConfig
	BlsKeyFile string `json:"bls_key_file" validate:"required"`
	Socket     string `json:"socket"`
}

// CoordinatorDeregistrationConfig combines basic and target configurations required for deregistering a coordinator.
type CoordinatorDeregistrationConfig struct {
	CoordinatorBasicConfig
	CoordinatorTargetConfig
}

// CoordinatorStatusCheckConfig represents the configuration required for checking coordinator status, extending basic
// settings.
type CoordinatorStatusCheckConfig struct {
	CoordinatorBasicConfig
}

// PrivateKeyStoreConfig holds the configuration for accessing a private key store, including its path and password.
type PrivateKeyStoreConfig struct {
	PrivateKeyStorePath     string `json:"private_key_store_path"     validate:"required"`
	PrivateKeyStorePassword string `json:"private_key_store_password"`
}

type Coordinator struct {
	repository    adapters.Repository
	logger        logging.Logger
	specification Specification
	configuration adapters.Configuration
	dryRun        bool
}

func NewCoordinator(
	repository adapters.Repository,
	logger logging.Logger,
	specification Specification,
	configuration adapters.Configuration,
	dryRun bool,
) *Coordinator {
	return &Coordinator{repository, logger, specification, configuration, dryRun}
}

func (coordinator *Coordinator) Type() string {
	return "middleware"
}

func stringToUnit8Array(s string) ([]uint8, error) {
	quorumIDStrings := strings.Split(s, ",")
	quorumIDList := make([]uint8, len(quorumIDStrings))

	for i, idStr := range quorumIDStrings {
		id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 8)
		if err != nil {
			return quorumIDList, fmt.Errorf("invalid quorum ID in list: %s, error: %w", idStr, err)
		}
		quorumIDList[i] = uint8(id)
	}
	return quorumIDList, nil
}

func (coordinator *Coordinator) Register() error {
	coordinator.logger.Debug("Reading coordinator configuration")
	cfg := &CoordinatorRegistrationConfig{}
	if err := coordinator.configuration.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	coordinator.logger.Debug(fmt.Sprintf("Configuring coordinator: %+v", cfg))

	coordinator.logger.Debug(fmt.Sprintf("Computing quorum id list: %s", string(cfg.QuorumIDList)))
	quorumIDList, err := stringToUnit8Array(string(cfg.QuorumIDList))
	if err != nil {
		return err
	}

	coordinator.logger.Debug("Creating handler")
	h := NewOperatorRegistrationHandler(coordinator.logger, coordinator.dryRun)

	txMgr, err := txmgr.NewTxManager(
		coordinator.specification.BaseSpecification,
		coordinator.configuration,
		coordinator.logger,
		coordinator.dryRun,
	)
	if err != nil {
		return err
	}

	coordinator.logger.Debug("Invoking operator registration")
	return h.Register(context.Background(), txMgr, NewOperatorRegistrationParams(
		cfg.Operator.Address,
		cfg.EthRpcUrl,
		coordinator.specification.ContractAddress,
		quorumIDList,
		cfg.ChainID,
		cfg.BlsKeyFile,
		cfg.Socket,
		coordinator.specification.ChurnerURL,
	))
}

func (coordinator *Coordinator) OptIn() error {
	fmt.Println("opt-in operation not required, use register instead")
	return nil
}

func (coordinator *Coordinator) OptOut() error {
	fmt.Println("opt-out operation not required, use deregister instead")
	return nil
}

func (coordinator *Coordinator) Deregister() error {
	coordinator.logger.Debug("Reading coordinator configuration")
	cfg := &CoordinatorDeregistrationConfig{}
	if err := coordinator.configuration.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	coordinator.logger.Debug(fmt.Sprintf("Configuring coordinator: %+v", cfg))

	coordinator.logger.Debug(fmt.Sprintf("Computing quorum id list: %s", string(cfg.QuorumIDList)))
	quorumIDList, err := stringToUnit8Array(string(cfg.QuorumIDList))
	if err != nil {
		return err
	}

	coordinator.logger.Debug("Creating handler")
	h := NewOperatorRegistrationHandler(coordinator.logger, coordinator.dryRun)

	txMgr, err := txmgr.NewTxManager(
		coordinator.specification.BaseSpecification,
		coordinator.configuration,
		coordinator.logger,
		coordinator.dryRun,
	)
	if err != nil {
		return err
	}

	coordinator.logger.Debug("Invoking operator deregistration")
	return h.Deregister(context.Background(), txMgr, NewOperatorParams(
		cfg.Operator.Address,
		cfg.EthRpcUrl,
		coordinator.specification.ContractAddress,
		quorumIDList,
		cfg.ChainID,
	))
}

func (coordinator *Coordinator) Status() (int, error) {
	coordinator.logger.Debug("Reading coordinator configuration")
	cfg := &CoordinatorStatusCheckConfig{}
	if err := coordinator.configuration.Unmarshal(cfg); err != nil {
		return -1, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	coordinator.logger.Debug(fmt.Sprintf("Configuring coordinator: %+v", cfg))

	coordinator.logger.Debug("Creating handler")
	h := NewOperatorRegistrationHandler(coordinator.logger, coordinator.dryRun)

	txMgr, err := txmgr.NewTxManager(
		coordinator.specification.BaseSpecification,
		coordinator.configuration,
		coordinator.logger,
		coordinator.dryRun,
	)
	if err != nil {
		return -1, err
	}

	coordinator.logger.Debug("Checking operator registration status")
	return h.CheckStatus(context.Background(), txMgr, NewServiceParams(
		cfg.Operator.Address,
		cfg.EthRpcUrl,
		coordinator.specification.ContractAddress,
	))
}
