package internal

// User struct to represent user information from the API
type User struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Keys []string `json:"ssh_keys"`
}

type ServerUser struct {
	Name    string   `json:"name"`
	SSHKeys []string `json:"ssh_keys"`
}

type ServerRegistrationRequest struct {
	OrganizationId string       `json:"organization_id"`
	Name           string       `json:"name"`
	IP             string       `json:"ip"`
	Users          []ServerUser `json:"users"`
	MachineId      string       `json:"machine_id"`
}

type RegistrationResponse struct {
	Token string `json:"token"`
}
