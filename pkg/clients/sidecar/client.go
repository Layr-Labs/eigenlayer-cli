package sidecar

import (
	"context"

	rewardsV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/sidecar/v1/rewards"
	"github.com/akuity/grpc-gateway-client/pkg/grpc/gateway"
)

//go:generate mockgen -destination=mocks/client.go -package=mocks github.com/Layr-Labs/eigenlayer-cli/pkg/clients/sidecar ISidecarClient

func NewSidecarRewardsClient(url string, opts ...gateway.ClientOption) (rewardsV1.RewardsGatewayClient, error) {
	client := rewardsV1.NewRewardsGatewayClient(gateway.NewClient(url, opts...))

	return client, nil
}

type ISidecarClient interface {
	GenerateClaimProof(
		context.Context,
		*rewardsV1.GenerateClaimProofRequest,
	) (*rewardsV1.GenerateClaimProofResponse, error)
	GetSummarizedRewardsForEarner(
		context.Context,
		*rewardsV1.GetSummarizedRewardsForEarnerRequest,
	) (*rewardsV1.GetSummarizedRewardsForEarnerResponse, error)
	ListDistributionRoots(
		context.Context,
		*rewardsV1.ListDistributionRootsRequest,
	) (*rewardsV1.ListDistributionRootsResponse, error)
}
