package middleware

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/txmgr"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"
	contractBLSApkRegistry "github.com/Layr-Labs/eigensdk-go/contracts/bindings/BLSApkRegistry"
	contractIAVSDirectory "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IAVSDirectory"
	contractIndexRegistry "github.com/Layr-Labs/eigensdk-go/contracts/bindings/IndexRegistry"
	contractRegistryCoordinator "github.com/Layr-Labs/eigensdk-go/contracts/bindings/RegistryCoordinator"
	contractServiceManagerBase "github.com/Layr-Labs/eigensdk-go/contracts/bindings/ServiceManagerBase"
	sdk "github.com/Layr-Labs/eigensdk-go/utils"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware/bn254"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware/churner"
	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware/pubip"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const MaxQuorumID = 254

// ServiceParams contains parameters needed to check the status of an operator in the AVS.
type ServiceParams struct {
	OperatorAddress       string
	RpcUrl                string
	ServiceManagerAddress string
}

// NewServiceParams initializes and returns a new ServiceParams instance with the provided
// parameters.
func NewServiceParams(
	operatorAddress string,
	rpcUrl string,
	serviceManagerAddress string,
) *ServiceParams {
	return &ServiceParams{
		OperatorAddress:       operatorAddress,
		RpcUrl:                rpcUrl,
		ServiceManagerAddress: serviceManagerAddress,
	}
}

// OperatorRegistrationParams defines the parameters required for operator registration in the service manager.
type OperatorRegistrationParams struct {
	*OperatorParams
	BlsKeyFile string
	Socket     string
	ChurnerUrl string
}

// NewOperatorRegistrationParams initializes and returns a new instance of OperatorRegistrationParams.
// It configures the operator's address, RPC URL, service manager address, quorum IDs, chain ID, signer, BLS key file,
// BLS key password, and socket.
func NewOperatorRegistrationParams(
	operatorAddress string,
	rpcUrl string,
	serviceManagerAddress string,
	quorumIDs []uint8,
	chainID int64,
	blsKeyFile string,
	socket string,
	churnerUrl string,
) *OperatorRegistrationParams {
	return &OperatorRegistrationParams{
		OperatorParams: NewOperatorParams(operatorAddress, rpcUrl, serviceManagerAddress, quorumIDs, chainID),
		BlsKeyFile:     blsKeyFile,
		Socket:         socket,
		ChurnerUrl:     churnerUrl,
	}
}

// OperatorParams represents the parameters required to deregister an operator.
// It includes the operator's address, RPC URL, service manager address, associated quorum IDs, chain ID, and signer.
type OperatorParams struct {
	*ServiceParams
	QuorumIDs []uint8
	ChainID   int64
}

// NewOperatorParams initializes and returns a pointer to an OperatorParams instance.
func NewOperatorParams(
	operatorAddress string,
	rpcUrl string,
	serviceManagerAddress string,
	quorumIDs []uint8,
	chainID int64,
) *OperatorParams {
	return &OperatorParams{
		ServiceParams: NewServiceParams(operatorAddress, rpcUrl, serviceManagerAddress),
		QuorumIDs:     quorumIDs,
		ChainID:       chainID,
	}
}

// OperatorRegistrationHandler manages operator registration processes for various blockchain-related entities.
type OperatorRegistrationHandler struct {
	logger logging.Logger
	dryRun bool
}

// NewOperatorRegistrationHandler initializes and returns a new instance of OperatorRegistrationHandler.
func NewOperatorRegistrationHandler(logger logging.Logger, dryRun bool) *OperatorRegistrationHandler {
	return &OperatorRegistrationHandler{logger, dryRun}
}

// quorumID is a type alias for uint8 used to represent a unique identifier for a quorum.
type quorumID = uint8

// quorumIDsToQuorumNumbers converts a slice of quorumID values to a slice of byte values representing the quorum
// numbers.
func quorumIDsToQuorumNumbers(quorumIds []quorumID) []byte {
	quorumNumbers := make([]byte, len(quorumIds))
	copy(quorumNumbers, quorumIds)
	return quorumNumbers
}

// pubKeyG1ToBN254G1Point converts a bn254.G1Point to a contractRegistryCoordinator.BN254G1Point.
func pubKeyG1ToBN254G1Point(p *bn254.G1Point) contractRegistryCoordinator.BN254G1Point {
	return contractRegistryCoordinator.BN254G1Point{
		X: p.X.BigInt(new(big.Int)),
		Y: p.Y.BigInt(new(big.Int)),
	}
}

// pubKeyG2ToBN254G2Point converts a bn254.G2Point object to a contractRegistryCoordinator.BN254G2Point structure.
func pubKeyG2ToBN254G2Point(p *bn254.G2Point) contractRegistryCoordinator.BN254G2Point {
	return contractRegistryCoordinator.BN254G2Point{
		X: [2]*big.Int{p.X.A1.BigInt(new(big.Int)), p.X.A0.BigInt(new(big.Int))},
		Y: [2]*big.Int{p.Y.A1.BigInt(new(big.Int)), p.Y.A0.BigInt(new(big.Int))},
	}
}

// getRegistrationParams generates the registration parameters required for registering an operator's public key.
func (m *OperatorRegistrationHandler) getRegistrationParams(
	ctx context.Context,
	operatorAddress common.Address,
	keypair *bn254.KeyPair,
	registryCoordinator *contractRegistryCoordinator.ContractRegistryCoordinator,
) (*contractRegistryCoordinator.IBLSApkRegistryPubkeyRegistrationParams, error) {

	msgToSignG1_, err := registryCoordinator.PubkeyRegistrationMessageHash(&bind.CallOpts{
		Context: ctx,
	}, operatorAddress)
	if err != nil {
		return nil, err
	}

	msgToSignG1 := bn254.NewG1Point(msgToSignG1_.X, msgToSignG1_.Y)
	signature := keypair.SignHashedToCurveMessage(msgToSignG1)

	signedMessageHashParam := contractRegistryCoordinator.BN254G1Point{
		X: signature.X.BigInt(big.NewInt(0)),
		Y: signature.Y.BigInt(big.NewInt(0)),
	}

	g1Point_ := pubKeyG1ToBN254G1Point(keypair.GetPubKeyG1())
	g1Point := contractRegistryCoordinator.BN254G1Point{
		X: g1Point_.X,
		Y: g1Point_.Y,
	}
	g2Point_ := pubKeyG2ToBN254G2Point(keypair.GetPubKeyG2())
	g2Point := contractRegistryCoordinator.BN254G2Point{
		X: g2Point_.X,
		Y: g2Point_.Y,
	}

	params := contractRegistryCoordinator.IBLSApkRegistryPubkeyRegistrationParams{
		PubkeyRegistrationSignature: signedMessageHashParam,
		PubkeyG1:                    g1Point,
		PubkeyG2:                    g2Point,
	}

	return &params, nil
}

// getRegistrationSignature generates a signature with associated salt and expiry for operator registration in AVS
// mapping.
func (m *OperatorRegistrationHandler) getRegistrationSignature(
	ctx context.Context,
	operatorAddress common.Address,
	blsKeyPair *bn254.KeyPair,
	txMgr *txmgr.TxManager,
	avsDirectory *contractIAVSDirectory.ContractIAVSDirectory,
	serviceManagerAddr common.Address,
	quorumIDs []quorumID,
) (*contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry, error) {
	privateKeyBytes := []byte(blsKeyPair.PrivKey.String())
	salt := [32]byte{}
	copy(
		salt[:],
		crypto.Keccak256(
			[]byte("churn"),
			[]byte(time.Now().String()),
			quorumIDsToQuorumNumbers(quorumIDs),
			privateKeyBytes,
		),
	)

	expiry := big.NewInt((time.Now().Add(10 * time.Minute)).Unix())

	// params to register operator in delegation manager's operator-avs mapping
	msgToSign, err := avsDirectory.CalculateOperatorAVSRegistrationDigestHash(
		&bind.CallOpts{
			Context: ctx,
		}, operatorAddress, serviceManagerAddr, salt, expiry)
	if err != nil {
		return nil, err
	}

	operatorSignature, err := txMgr.Sign(msgToSign[:])
	if err != nil {
		return nil, err
	}
	operatorSignature[64] += 27
	operatorSignatureWithSaltAndExpiry := contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: operatorSignature,
		Salt:      salt,
		Expiry:    expiry,
	}

	return &operatorSignatureWithSaltAndExpiry, nil

}

// isLocalhost checks if the provided socket string refers to a localhost address (localhost, 127.0.0.1, or 0.0.0.0).
func isLocalhost(socket string) bool {
	return strings.Contains(socket, "localhost") || strings.Contains(socket, "127.0.0.1") ||
		strings.Contains(socket, "0.0.0.0")
}

// parseOperatorSocket splits a socket string into host, dispersal port, and retrieval port, returning an error for
// invalid formats.
func parseOperatorSocket(socket string) (host string, dispersalPort string, retrievalPort string, err error) {
	s := strings.Split(socket, ";")
	if len(s) != 2 {
		err = fmt.Errorf("invalid socket address format, missing retrieval port: %s", socket)
		return
	}
	retrievalPort = s[1]

	s = strings.Split(s[0], ":")
	if len(s) != 2 {
		err = fmt.Errorf("invalid socket address format: %s", socket)
		return
	}
	host = s[0]
	dispersalPort = s[1]

	return
}

// socketAddress generates a formatted socket string by combining a public IP address with dispersal and retrieval
// ports.
// It retrieves the public IP using the provided IP provider, and returns an error if the IP retrieval fails.
func socketAddress(
	ctx context.Context,
	provider pubip.Provider,
	dispersalPort string,
	retrievalPort string,
) (string, error) {
	ip, err := provider.PublicIPAddress(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get public ip address from IP provider: %w", err)
	}
	socket := makeOperatorSocket(ip, dispersalPort, retrievalPort)
	return socket.String(), nil
}

// operatorSocket represents a custom string type used for encoding network socket information for an operator.
type operatorSocket string

// String converts the operatorSocket value to its underlying string representation.
func (s operatorSocket) String() string {
	return string(s)
}

// makeOperatorSocket creates an operatorSocket by formatting the node IP, dispersal port, and retrieval port into a
// string.
func makeOperatorSocket(nodeIP, dispersalPort, retrievalPort string) operatorSocket {
	return operatorSocket(fmt.Sprintf("%s:%s;%s", nodeIP, dispersalPort, retrievalPort))
}

// getSocket resolves the operator's socket address, replacing localhost with a publicly accessible IP if necessary.
func (m *OperatorRegistrationHandler) getSocket(ctx context.Context, socket string) (string, error) {
	_, dispersalPort, retrievalPort, err := parseOperatorSocket(socket)
	if err != nil {
		return "", sdk.WrapError("failed to parse operator socket", err)
	}

	sock := socket
	if isLocalhost(sock) {
		pubIPProvider := pubip.SeeIP
		sock, err = socketAddress(ctx, pubIPProvider, dispersalPort, retrievalPort)
		if err != nil {
			return "", sdk.WrapError("failed to get socket address from ip provider", err)
		}
	}
	return sock, nil
}

// getBlsKeyPair reads and decrypts a BLS private key from a file and returns a KeyPair containing the private and
// public keys.
func (m *OperatorRegistrationHandler) getBlsKeyPair(
	blsKeyFile string,
	blsKeyPassword string,
) (*bn254.KeyPair, error) {
	kp, err := bls.ReadPrivateKeyFromFile(blsKeyFile, blsKeyPassword)
	if err != nil {
		return nil, sdk.WrapError("failed to read or decrypt the BLS private key", err)
	}

	keyPair := &bn254.KeyPair{
		PrivKey: kp.PrivKey,
		PubKey: &bn254.G1Point{
			G1Affine: kp.PubKey.G1Affine,
		},
	}
	return keyPair, nil
}

func (m *OperatorRegistrationHandler) createServices(
	ctx context.Context,
	txMgr *txmgr.TxManager,
	params *ServiceParams,
) (*contractServiceManagerBase.ContractServiceManagerBase, *contractRegistryCoordinator.ContractRegistryCoordinator, error) {
	m.logger.Debug("Creating service manager", "ServiceManagerAddress", params.ServiceManagerAddress)
	blsApkReg, err := contractBLSApkRegistry.NewContractBLSApkRegistry(
		common.HexToAddress(params.ServiceManagerAddress),
		txMgr.Client,
	)
	if err != nil {
		return nil, nil, sdk.WrapError("failed to create service manager", err)
	}

	serviceManager, err := contractServiceManagerBase.NewContractServiceManagerBase(
		common.HexToAddress(params.ServiceManagerAddress),
		txMgr.Client,
	)
	if err != nil {
		return nil, nil, sdk.WrapError("failed to create service manager", err)
	}

	m.logger.Debug("Creating registry coordinator")
	coordinatorAddr, err := blsApkReg.RegistryCoordinator(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, sdk.WrapError("failed to create coordinator address", err)
	}

	m.logger.Debug("Creating contract registry coordinator", "CoordinatorAddress", coordinatorAddr.String())
	registryCoordinator, err := contractRegistryCoordinator.NewContractRegistryCoordinator(
		coordinatorAddr,
		txMgr.Client,
	)
	if err != nil {
		return nil, nil, sdk.WrapError("failed to create service manager", err)
	}

	return serviceManager, registryCoordinator, nil
}

func (m *OperatorRegistrationHandler) churn(
	ctx context.Context,
	params *OperatorRegistrationParams,
	blsKeyPair *bn254.KeyPair,
) ([]contractRegistryCoordinator.IRegistryCoordinatorOperatorKickParam, contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry, error) {
	m.logger.Info("Calling churner to churn operators")
	churnerClient := churner.NewChurnerClient(params.ChurnerUrl, true, 10*time.Second, m.logger)
	churnReply, err := churnerClient.Churn(ctx, params.OperatorAddress, blsKeyPair, params.QuorumIDs)
	if err != nil {
		return nil, contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry{}, fmt.Errorf(
			"failed to request churn approval: %w",
			err,
		)
	}

	operatorsToChurn := make(
		[]contractRegistryCoordinator.IRegistryCoordinatorOperatorKickParam,
		len(churnReply.OperatorsToChurn),
	)
	for i := range churnReply.OperatorsToChurn {
		if churnReply.OperatorsToChurn[i].QuorumId >= MaxQuorumID {
			return nil, contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry{}, errors.New(
				"quorum id is out of range",
			)
		}

		operatorsToChurn[i] = contractRegistryCoordinator.IRegistryCoordinatorOperatorKickParam{
			QuorumNumber: uint8(churnReply.OperatorsToChurn[i].QuorumId),
			Operator:     common.BytesToAddress(churnReply.OperatorsToChurn[i].Operator),
		}
	}

	var salt [32]byte
	copy(salt[:], churnReply.SignatureWithSaltAndExpiry.Salt[:])
	churnApproverSignature := contractRegistryCoordinator.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: churnReply.SignatureWithSaltAndExpiry.Signature,
		Salt:      salt,
		Expiry:    new(big.Int).SetInt64(churnReply.SignatureWithSaltAndExpiry.Expiry),
	}
	return operatorsToChurn, churnApproverSignature, nil
}

func (m *OperatorRegistrationHandler) shouldCallChurner(
	ctx context.Context,
	params *OperatorRegistrationParams,
	registryCoordinator *contractRegistryCoordinator.ContractRegistryCoordinator,
	client *ethclient.Client,
) (bool, error) {
	m.logger.Debug("Creating index registry")
	indexRegistryAddr, err := registryCoordinator.IndexRegistry(&bind.CallOpts{Context: ctx})
	if err != nil {
		return false, sdk.WrapError("failed to fetch index registry address", err)
	}

	m.logger.Debug("Creating Index Registry")
	indexregistry, err := contractIndexRegistry.NewContractIndexRegistry(indexRegistryAddr, client)
	if err != nil {
		return false, sdk.WrapError("failed to create index registry", err)
	}

	m.logger.Debug("Checking if quorums are full")
	shouldCallChurner := false
	for _, quorumID := range params.QuorumIDs {
		m.logger.Debug("Getting operator set parameters", "QuorumID", quorumID)
		operatorSetParams, err := registryCoordinator.GetOperatorSetParams(&bind.CallOpts{Context: ctx}, quorumID)
		if err != nil {
			return false, err
		}

		m.logger.Debug("Getting total operators for quorum", "QuorumID", quorumID)
		numberOfRegisteredOperators, err := indexregistry.TotalOperatorsForQuorum(
			&bind.CallOpts{Context: ctx},
			quorumID,
		)
		if err != nil {
			return false, err
		}

		// if the quorum is full, we need to call the churner
		if operatorSetParams.MaxOperatorCount == numberOfRegisteredOperators {
			m.logger.Info(
				"Quorum is full and need to call the churner",
				"QuorumID",
				quorumID,
				"MaxOperatorCount",
				operatorSetParams.MaxOperatorCount,
				"NumberOfRegisteredOperators",
				numberOfRegisteredOperators,
			)
			shouldCallChurner = true
			break
		}
		m.logger.Debug(
			"Quorum is not full and no need to call the churner",
			"QuorumID",
			quorumID,
			"MaxOperatorCount",
			operatorSetParams.MaxOperatorCount,
			"NumberOfRegisteredOperators",
			numberOfRegisteredOperators,
		)
	}
	return shouldCallChurner, nil
}

// Register registers an operator on the blockchain using provided parameters and validates the necessary components.
func (m *OperatorRegistrationHandler) Register(
	ctx context.Context,
	txMgr *txmgr.TxManager,
	params *OperatorRegistrationParams,
) error {
	m.logger.Debug("Registration parameters",
		"OperatorAddress", params.OperatorAddress,
		"RpcUrl", params.RpcUrl,
		"ServiceManagerAddress", params.ServiceManagerAddress,
		"BlsKeyFile", params.BlsKeyFile,
		"QuorumIDs", params.QuorumIDs,
		"Socket", params.Socket,
		"ChurnerUrl", params.ChurnerUrl,
		"ChainID", params.ChainID)
	serviceManager, registryCoordinator, err := m.createServices(ctx, txMgr, params.ServiceParams)
	if err != nil {
		return err
	}

	m.logger.Debug("Creating AVS directory")
	avsDirectoryAddr, err := serviceManager.AvsDirectory(&bind.CallOpts{Context: ctx})
	if err != nil {
		return sdk.WrapError("failed to create avs directory address", err)
	}

	m.logger.Debug("Creating contract AVS directory", "AVSDirectoryAddress", avsDirectoryAddr.String())
	avsDirectory, err := contractIAVSDirectory.NewContractIAVSDirectory(avsDirectoryAddr, txMgr.Client)
	if err != nil {
		return sdk.WrapError("failed to create avs directory", err)
	}

	m.logger.Debug("Reading BLS key pair")
	p := utils.NewPrompter()
	blsPassword, err := p.InputHiddenString("Enter password to decrypt the bls private key:", "",
		func(password string) error {
			return nil
		},
	)
	if err != nil {
		return sdk.WrapError("failed to read bls key password", err)
	}
	blsKeyPair, err := m.getBlsKeyPair(params.BlsKeyFile, blsPassword)
	if err != nil {
		return sdk.WrapError("failed to get bls key pair", err)
	}

	m.logger.Debug("Creating registration parameters", "OperatorAddress", params.OperatorAddress)
	regParam, err := m.getRegistrationParams(
		ctx,
		common.HexToAddress(params.OperatorAddress),
		blsKeyPair,
		registryCoordinator,
	)
	if err != nil {
		return sdk.WrapError("failed to get registration parameters", err)
	}

	m.logger.Debug("Creating registration signature")
	signature, err := m.getRegistrationSignature(
		ctx,
		common.HexToAddress(params.OperatorAddress),
		blsKeyPair,
		txMgr,
		avsDirectory,
		common.HexToAddress(params.ServiceManagerAddress),
		params.QuorumIDs,
	)
	if err != nil {
		return sdk.WrapError("failed to get registration signature", err)
	}

	m.logger.Debug("Creating socket", "Socket", params.Socket)
	socket, err := m.getSocket(ctx, params.Socket)
	if err != nil {
		return sdk.WrapError("failed to get socket", err)
	}
	m.logger.Debug("Socket created", "Socket", socket)

	var shouldCallChurner bool
	if params.ChurnerUrl == "" {
		m.logger.Debug("Skipping churning as no churner URL is specified")
		shouldCallChurner = false
	} else {
		shouldCallChurner, err = m.shouldCallChurner(ctx, params, registryCoordinator, txMgr.Client)
		if err != nil {
			return err
		}
	}

	//opts.NoSend = true
	if shouldCallChurner {
		operatorsToChurn, churnApproverSignature, err := m.churn(ctx, params, blsKeyPair)
		if err != nil {
			return err
		}

		m.logger.Debug("Registering operator with churning",
			"quorumNumbers", quorumIDsToQuorumNumbers(params.QuorumIDs),
			"socket", socket,
			"params", *regParam,
			"operatorsToChurn", operatorsToChurn,
			"churnApproverSignature", churnApproverSignature,
			"operatorSignature", *signature)

		tx, receipt, err := txMgr.CallAndWaitForReceipt(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			return registryCoordinator.RegisterOperatorWithChurn(
				opts,
				quorumIDsToQuorumNumbers(params.QuorumIDs),
				socket,
				*regParam,
				operatorsToChurn,
				churnApproverSignature,
				*signature,
			)
		})

		if err != nil {
			return sdk.WrapError("failed to register operator with churn", err)
		}

		if m.dryRun {
			m.logger.Debug("Operator registered with churn", "txHash", tx.Hash().Hex())
		} else {
			m.logger.Debug("Operator registered with churn", "txHash", tx.Hash().Hex(), "receipt", receipt.Status)
		}
	} else {
		m.logger.Debug("Registering operator",
			"quorumNumbers", quorumIDsToQuorumNumbers(params.QuorumIDs),
			"socket", socket,
			"params", *regParam,
			"operatorSignature", *signature)

		tx, receipt, err := txMgr.CallAndWaitForReceipt(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
			op, err2 := registryCoordinator.RegisterOperator(
				opts,
				quorumIDsToQuorumNumbers(params.QuorumIDs),
				socket,
				*regParam,
				*signature,
			)
			if err2 != nil {
				return nil, sdk.WrapError("failed to register operator", err2)
			}
			return op, err2
		})

		if err != nil {
			return sdk.WrapError("Failed to register operator", err)
		}

		if m.dryRun {
			m.logger.Debug("Operator registered with out churn", "txHash", tx.Hash().Hex())
		} else {
			m.logger.Debug("Operator registered with out churn", "txHash", tx.Hash().Hex(), "receipt", receipt.Status)
		}
	}

	return nil
}

// Deregister removes an operator's registration from the registry coordinator using the provided deregistration
// parameters.
func (m *OperatorRegistrationHandler) Deregister(
	ctx context.Context,
	txMgr *txmgr.TxManager,
	params *OperatorParams,
) error {
	m.logger.Debug("Deregistration parameters",
		"OperatorAddress", params.OperatorAddress,
		"RpcUrl", params.RpcUrl,
		"ServiceManagerAddress", params.ServiceManagerAddress,
		"QuorumIDs", params.QuorumIDs,
		"ChainID", params.ChainID)
	_, registryCoordinator, err := m.createServices(ctx, txMgr, params.ServiceParams)
	if err != nil {
		return err
	}

	tx, receipt, err := txMgr.CallAndWaitForReceipt(ctx, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		op, err2 := registryCoordinator.DeregisterOperator(opts, quorumIDsToQuorumNumbers(params.QuorumIDs))
		if err2 != nil {
			return nil, sdk.WrapError("failed to deregister operator", err2)
		}
		return op, err2
	})

	if err != nil {
		return err
	}

	if m.dryRun {
		m.logger.Debug("Operator deregistered", "tx", tx.Hash().Hex())
	} else {
		m.logger.Debug("Operator deregistered", "tx", tx.Hash().Hex(), "receipt", receipt.Status)
	}

	return nil
}

// CheckStatus obtains the status of an operator from the RegistryCoordinator using the provided parameters.
// It connects to the Ethereum client and interacts with contracts to fetch the operator's status.
func (m *OperatorRegistrationHandler) CheckStatus(
	ctx context.Context,
	txMgr *txmgr.TxManager,
	params *ServiceParams,
) (int, error) {
	m.logger.Debug("Check Status parameters",
		"OperatorAddress", params.OperatorAddress,
		"RpcUrl", params.RpcUrl,
		"ServiceManagerAddress", params.ServiceManagerAddress)
	_, registryCoordinator, err := m.createServices(ctx, txMgr, params)
	if err != nil {
		return -1, err
	}
	status, err := registryCoordinator.GetOperatorStatus(
		&bind.CallOpts{Context: ctx},
		common.HexToAddress(params.OperatorAddress),
	)
	if err != nil {
		return -1, sdk.WrapError("Failed to get registration status of operator", err)
	}

	return int(status), nil
}
