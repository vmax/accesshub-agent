package internal

import (
	"fmt"
	"os"
)

func WriteSystemdUnit(apiDomain string, token string) error {
	unitPath := "/etc/systemd/system/accesshub-agent.service"
	unitContent := fmt.Sprintf(`[Unit]
Description=AccessHub Agent service
After=network.target
[Service]
Type=simple
ExecStart=/usr/local/bin/accesshub-agent
Environment=API_DOMAIN=%s
Environment=SERVER_TOKEN=%s
Restart=always

[Install]
WantedBy=multi-user.target
	`, apiDomain, token)
	return os.WriteFile(unitPath, []byte(unitContent), 0644)
}
