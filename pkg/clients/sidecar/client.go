package sidecar

import (
	"context"
	"net/http"
	"time"

	rewardsV1 "github.com/Layr-Labs/protocol-apis/gen/protos/eigenlayer/sidecar/v1/rewards"
	"github.com/akuity/grpc-gateway-client/pkg/grpc/gateway"
)

//go:generate mockgen -destination=mocks/client.go -package=mocks github.com/Layr-Labs/eigenlayer-cli/pkg/clients/sidecar ISidecarClient

type customTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Add default headers
	for key, value := range t.headers {
		// Only set the header if it's not already set
		if reqClone.Header.Get(key) == "" {
			reqClone.Header.Set(key, value)
		}
	}

	// Pass the modified request to the base transport
	return t.base.RoundTrip(reqClone)
}

func NewSidecarRewardsClient(url string, opts ...gateway.ClientOption) (rewardsV1.RewardsGatewayClient, error) {
	httpClient := http.Client{
		Transport: &customTransport{
			base: http.DefaultTransport,
			headers: map[string]string{
				"x-sidecar-source": "eigenlayer-cli",
			},
		},
		Timeout: 300 * time.Second,
	}

	if len(opts) == 0 {
		opts = make([]gateway.ClientOption, 0)
	}
	opts = append(opts, gateway.WithHTTPClient(&httpClient))

	gwClient := gateway.NewClient(url, opts...)

	client := rewardsV1.NewRewardsGatewayClient(gwClient)

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
