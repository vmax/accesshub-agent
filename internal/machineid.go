package internal

import (
	"os"
	"strings"
)

func GetMachineID() string {
	id, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		Error("Error reading machine ID", "error", err)
		return ""
	}
	return strings.TrimSpace(string(id))
}
