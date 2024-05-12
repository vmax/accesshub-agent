package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetServerUsers(directory string) ([]ServerUser, error) {
	var users []ServerUser = []ServerUser{}

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			user := file.Name()
			sshDir := filepath.Join(directory, user, ".ssh")
			authorizedKeysFile := filepath.Join(sshDir, "authorized_keys")

			// Check if .ssh directory exists
			if _, err := os.Stat(sshDir); err == nil {
				// Read authorized_keys file
				keys, err := os.ReadFile(authorizedKeysFile)
				if err == nil {
					// Split keys by newline
					keyList := strings.Split(string(keys), "\n")

					// Remove empty lines
					var cleanedKeys []string
					for _, key := range keyList {
						if key != "" && !strings.HasPrefix(key, "#") {
							cleanedKeys = append(cleanedKeys, key)
						}
					}

					// Append user and keys to slice
					users = append(users, ServerUser{Name: user, SSHKeys: cleanedKeys})
				}
			}
		}
	}

	return users, nil
}

func RegisterServer(apiDomain string, organizationId string) (string, error) {
	hostname, _ := os.Hostname()
	users, errUsers := GetServerUsers("/home")

	if errUsers != nil {
		Error("Error getting server users", "error", errUsers)
	} else {
		Info("Obtained server users", "len(users)", len(users))
	}

	requestBody, err := json.Marshal(ServerRegistrationRequest{
		IP:             GetOutboundIP(),
		Name:           hostname,
		Users:          users,
		OrganizationId: organizationId,
		MachineId:      GetMachineID(),
	})
	if err != nil {
		return "", err
	}
	Info("Registering server", "organization_id", organizationId, "requestBody", string(requestBody))

	// Send POST request
	apiURL := fmt.Sprintf("https://%s/api/agent/v1/register", apiDomain)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var token RegistrationResponse
	json.Unmarshal(body, &token)

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		Error("registration failed", "status_code", resp.StatusCode, "response", body)
		return "", fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	Info("Server registered successfully")
	return token.Token, nil
}

func FetchUsers(apiURL string, serverToken string) ([]User, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("X-Server-Token", serverToken)
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(body, &users)
	if err != nil {
		Error("Could not deserialize response", "body", string(body))
		return nil, err
	}

	return users, nil
}
