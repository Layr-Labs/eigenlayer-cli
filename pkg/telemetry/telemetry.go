package telemetry

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/utils"

	"github.com/posthog/posthog-go"

	"github.com/urfave/cli/v2"
)

const (
	cliTelemetryEnabledKey = "EIGENLAYER_CLI_TELEMETRY_ENABLED"
)

// telemetryToken value is set at build and install scripts using ldflags
// version value is set at build scripts using ldflags
var (
	telemetryToken    = ""
	telemetryInstance = "https://us.i.posthog.com"
	version           = "development"
)

func AfterRunAction() cli.AfterFunc {
	return func(c *cli.Context) error {
		if IsTelemetryEnabled() {
			HandleTacking(c)
		}
		return nil
	}
}

func IsTelemetryEnabled() bool {
	telemetryEnabled := os.Getenv(cliTelemetryEnabledKey)
	return len(telemetryEnabled) == 0 || telemetryEnabled == "true"
}

func HandleTacking(cCtx *cli.Context) {
	if telemetryToken == "" {
		return
	}
	client, _ := posthog.NewWithConfig(telemetryToken, posthog.Config{Endpoint: telemetryInstance})
	defer client.Close()

	// In v3, c.Command.FullName() can be used to get the full command name
	// TODO(madhur): to update once v3 is released
	commandPath := cCtx.Command.HelpName
	usr, _ := user.Current() // use empty string if err
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s", usr.Username, usr.Uid)))
	userID := base64.StdEncoding.EncodeToString(hash[:])
	network := "unknown"
	if chainIdVal, ok := cCtx.App.Metadata["network"]; ok {
		chainIdInt, err := strconv.Atoi(chainIdVal.(string))
		// If this is an error just ignore it and continue
		if err == nil {
			network = utils.ChainIdToNetworkName(int64(chainIdInt))
		}

	}
	telemetryProperties := make(map[string]interface{})
	telemetryProperties["command"] = commandPath
	telemetryProperties["version"] = version
	telemetryProperties["os"] = runtime.GOOS
	telemetryProperties["network"] = network
	_ = client.Enqueue(posthog.Capture{
		DistinctId: userID,
		Event:      "eigenlayer-cli",
		Properties: telemetryProperties,
	})
}
