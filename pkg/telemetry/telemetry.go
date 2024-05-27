package telemetry

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os/user"
	"runtime"

	"github.com/posthog/posthog-go"

	"github.com/urfave/cli/v2"
)

// telemetryToken value is set at build and install scripts using ldflags
var (
	telemetryToken    = ""
	telemetryInstance = "https://us.i.posthog.com"
)

func AfterRunAction() cli.AfterFunc {
	return func(c *cli.Context) error {
		// In v3, c.Command.FullName() can be used to get the full command name
		// TODO(madhur): to update once v3 is released
		HandleTacking(c.Command.HelpName)
		return nil
	}
}

func GetCLIVersion() string {
	return "development"
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
	telemetryProperties["version"] = GetCLIVersion()
	telemetryProperties["os"] = runtime.GOOS
	_ = client.Enqueue(posthog.Capture{
		DistinctId: userID,
		Event:      "cli-command",
		Properties: telemetryProperties,
	})
}
