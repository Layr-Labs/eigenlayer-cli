package release

import "github.com/Layr-Labs/release-management-service-client/pkg/client"

type listOperatorReleasesConfig struct {
	Network     string
	OperatorId  string
	Environment string
	OutputType  string
	OutputFile  string
	RmsClient   client.ReleaseManagementServiceClient
}

type listAvsReleaseKeysConfig struct {
	Network     string
	AvsId       string
	Environment string
	OutputType  string
	OutputFile  string
	RmsClient   client.ReleaseManagementServiceClient
}
