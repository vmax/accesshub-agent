package main

import (
	"fmt"
	"os"
	"time"
	"vmax/accesshub-agent/internal"
)

const errorThreshold = 5

func main() {
	var errors = 0
	if argv := os.Args; len(argv) > 1 {
		if argv[1] == "register" {
			domain := argv[2]
			organizationId := argv[3]
			internal.Info("Registering server", "organization_id", organizationId, "domain", domain)
			result, err := internal.RegisterServer(domain, organizationId)
			if err != nil {
				internal.Error("Error registering server", "error", err)
				os.Exit(1)
			}
			systemdUnitWrite := internal.WriteSystemdUnit(domain, result)
			if systemdUnitWrite != nil {
				internal.Error("Error writing systemd unit", "error", err)
				os.Exit(1)
			}
			return
		}
	}

	domain := os.Getenv("API_DOMAIN")
	serverToken := os.Getenv("SERVER_TOKEN")
	apiURL := fmt.Sprintf("https://%s/api/agent/v1/users", domain)
	updateInterval := 10 * time.Second

	for {
		if errors >= errorThreshold {
			internal.Error("Hit error threshold, exiting")
			internal.DisableSystemdUnit()
			os.Exit(1)
		}

		// Fetch user information from the API
		users, err := internal.FetchUsers(apiURL, serverToken)
		if err != nil {
			errors++
			internal.Error("Error fetching user information", "error", err, "num_errors", errors)
			time.Sleep(updateInterval)
			continue
		} else {
			internal.Info("Fetched users", "len(users)", len(users))
			errors = 0
			err = internal.UpdateUsers(users)
			if err != nil {
				internal.Error("Error updating users", "err", err)
			}
		}

		time.Sleep(updateInterval)
	}
}
