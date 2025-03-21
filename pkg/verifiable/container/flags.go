package container

import (
	"github.com/urfave/cli/v2"
)

var (
	containerDigestFlag = cli.StringFlag{
		Name:     "container-digest",
		Aliases:  []string{"cd"},
		Usage:    "The digest of the container.",
		Required: true,
		EnvVars:  []string{"CONTAINER_DIGEST"},
	}
	repositoryLocationFlag = cli.StringFlag{
		Name:     "repository-location",
		Aliases:  []string{"rl"},
		Usage:    "The GHCR repository location to tag the signature artifact with.",
		Required: true,
		EnvVars:  []string{"REPOSITORY_LOCATION"},
	}
	ecdsaPublicKeyFlag = cli.StringFlag{
		Name:    "ecdsa-public-key",
		Aliases: []string{"e"},
		Usage:   "ECDSA public key to annotate releases with for verification by users.",
		EnvVars: []string{"ECDSA_PUBLIC_KEY"},
	}
)
