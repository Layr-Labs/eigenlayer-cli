package container

import (
	"github.com/urfave/cli/v2"
)

var (
	imageIdFlag = cli.StringFlag{
		Name:     "image-id",
		Aliases:  []string{"ii"},
		Usage:    "The image id of the container as the subject of the command.",
		Required: true,
		EnvVars:  []string{"IMAGE_ID"},
	}
	repositoryLocationFlag = cli.StringFlag{
		Name:     "repository-location",
		Aliases:  []string{"rl"},
		Usage:    "The GHCR repository location to tag the signature artifact with.",
		Required: true,
		EnvVars:  []string{"IMAGE_ID"},
	}
	containerTagFlag = cli.StringFlag{
		Name:     "tag",
		Aliases:  []string{"t"},
		Usage:    "The container image tag.",
		Required: true,
		EnvVars:  []string{"TAG"},
	}
)
