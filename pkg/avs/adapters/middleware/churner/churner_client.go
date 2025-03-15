package churner

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	grpc2 "github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware/churner/grpc"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/middleware/bn254"

	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/utils"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type QuorumID = uint8
type ChurnRequest struct {
	OperatorAddress            gethcommon.Address
	OperatorToRegisterPubkeyG1 *bn254.G1Point
	OperatorToRegisterPubkeyG2 *bn254.G2Point
	OperatorRequestSignature   *bn254.Signature
	Salt                       [32]byte
	QuorumIDs                  []QuorumID
}

type ChurnerClient interface {
	// Churn sends a churn request to the churner service
	// The quorumIDs cannot be empty, but may contain quorums that the operator is already registered in.
	// If the operator is already registered in a quorum, the churner will ignore it and continue with the other
	// quorums.
	Churn(
		ctx context.Context,
		operatorAddress string,
		keyPair *bn254.KeyPair,
		quorumIDs []QuorumID,
	) (*grpc2.ChurnReply, error)
}

type churnerClient struct {
	churnerURL    string
	useSecureGrpc bool
	timeout       time.Duration
	logger        logging.Logger
}

func NewChurnerClient(
	churnerURL string,
	useSecureGrpc bool,
	timeout time.Duration,
	logger logging.Logger,
) ChurnerClient {
	return &churnerClient{
		churnerURL:    churnerURL,
		useSecureGrpc: useSecureGrpc,
		timeout:       timeout,
		logger:        logger.With("component", "ChurnerClient"),
	}
}

func CalculateRequestHash(churnRequest *ChurnRequest) [32]byte {
	var requestHash [32]byte
	requestHashBytes := crypto.Keccak256(
		[]byte("ChurnRequest"),
		[]byte(churnRequest.OperatorAddress.Hex()),
		churnRequest.OperatorToRegisterPubkeyG1.Serialize(),
		churnRequest.OperatorToRegisterPubkeyG2.Serialize(),
		churnRequest.Salt[:],
	)
	copy(requestHash[:], requestHashBytes)
	return requestHash
}

func (c *churnerClient) Churn(
	ctx context.Context,
	operatorAddress string,
	keyPair *bn254.KeyPair,
	quorumIDs []QuorumID,
) (*grpc2.ChurnReply, error) {
	if len(quorumIDs) == 0 {
		return nil, errors.New("quorumIDs cannot be empty")
	}
	// generate salt
	privateKeyBytes := []byte(keyPair.PrivKey.String())
	salt := crypto.Keccak256([]byte("churn"), []byte(time.Now().String()), quorumIDs[:], privateKeyBytes)

	churnRequest := &ChurnRequest{
		OperatorAddress:            gethcommon.HexToAddress(operatorAddress),
		OperatorToRegisterPubkeyG1: keyPair.PubKey,
		OperatorToRegisterPubkeyG2: keyPair.GetPubKeyG2(),
		OperatorRequestSignature:   &bn254.Signature{},
		QuorumIDs:                  quorumIDs,
	}

	copy(churnRequest.Salt[:], salt)

	// sign the request
	churnRequest.OperatorRequestSignature = keyPair.SignMessage(CalculateRequestHash(churnRequest))

	// convert to protobuf
	churnRequestPb := &grpc2.ChurnRequest{
		OperatorToRegisterPubkeyG1: churnRequest.OperatorToRegisterPubkeyG1.Serialize(),
		OperatorToRegisterPubkeyG2: churnRequest.OperatorToRegisterPubkeyG2.Serialize(),
		OperatorRequestSignature:   churnRequest.OperatorRequestSignature.Serialize(),
		Salt:                       salt[:],
		OperatorAddress:            operatorAddress,
	}

	churnRequestPb.QuorumIds = make([]uint32, len(quorumIDs))
	for i, quorumID := range quorumIDs {
		churnRequestPb.QuorumIds[i] = uint32(quorumID)
	}
	credential := insecure.NewCredentials()
	if c.useSecureGrpc {
		config := &tls.Config{}
		credential = credentials.NewTLS(config)
	}

	conn, err := grpc.NewClient(
		c.churnerURL,
		grpc.WithTransportCredentials(credential),
	)
	if err != nil {
		return nil, utils.WrapError("cannot connect to churner", err)
	}
	defer conn.Close()

	gc := grpc2.NewChurnerClient(conn)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	opt := grpc.MaxCallSendMsgSize(1024 * 1024 * 300)

	return gc.Churn(ctx, churnRequestPb, opt)
}
