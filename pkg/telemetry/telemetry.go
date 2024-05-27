package telemetry

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/user"
	"runtime"

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
		// In v3, c.Command.FullName() can be used to get the full command name
		// TODO(madhur): to update once v3 is released
		if IsTelemetryEnabled() {
			HandleTacking(c.Command.HelpName)
		}
		return nil
	}
}

func IsTelemetryEnabled() bool {
	telemetryEnabled := os.Getenv(cliTelemetryEnabledKey)
	return len(telemetryEnabled) == 0 || telemetryEnabled == "true"
}

func HandleTacking(commandPath string) {
	if telemetryToken == "" {
		return
	}
	client, _ := posthog.NewWithConfig(telemetryToken, posthog.Config{Endpoint: telemetryInstance})
	defer client.Close()

	usr, _ := user.Current() // use empty string if err
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s", usr.Username, usr.Uid)))
	userID := base64.StdEncoding.EncodeToString(hash[:])
	telemetryProperties := make(map[string]interface{})
	telemetryProperties["command"] = commandPath
	telemetryProperties["version"] = version
	telemetryProperties["os"] = runtime.GOOS
	_ = client.Enqueue(posthog.Capture{
		DistinctId: userID,
		Event:      "eigenlayer-cli",
		Properties: telemetryProperties,
	})
}
