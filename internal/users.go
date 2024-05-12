package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/willdonnelly/passwd"
)

func getSudoGroups() string {
	var groups []string
	for _, group := range []string{"wheel", "sudo", "admin"} {
		_, err := user.LookupGroup(group)
		if err == nil {
			groups = append(groups, group)
		}
	}
	return strings.Join(groups, ",")
}

func createUser(username string) error {
	cmd := exec.Command("useradd", username, "--password", "", "--create-home", "--shell", "/bin/bash", "--groups", getSudoGroups())
	err := cmd.Run()
	return err
}

func deleteUser(username string) error {
	cmd := exec.Command("userdel", "--force", "--remove", username)
	err := cmd.Run()
	return err
}

func getUIDMin() (int, error) {
	file, err := os.Open("/etc/login.defs")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "UID_MIN") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uidMinStr := fields[1]
				uidMin, err := strconv.Atoi(uidMinStr)
				if err != nil {
					return 0, err
				}
				return uidMin, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, nil
}

func UpdateUsers(users []User) error {
	existingUsers, existingUsersErr := passwd.Parse()

	if existingUsersErr != nil {
		Error("Error getting existing users", "error", existingUsersErr)
		return existingUsersErr
	}

	minUid, minUidErr := getUIDMin()
	Info("minUid", "_", minUid)
	if minUidErr != nil {
		Error("Error getting UID_MIN", "error", minUidErr)
		return minUidErr
	}

	for existingUsername, existingUser := range existingUsers {
		uid, _ := strconv.Atoi(existingUser.Uid)
		if uid < minUid {
			continue
		}
		found := false
		for _, newUser := range users {
			if existingUsername == newUser.Name {
				found = true
				break
			}
		}
		if !found {
			Info("Deleting user", "user", existingUsername)
			err := deleteUser(existingUsername)
			if err != nil {
				return err
			}
		}
	}

	for _, u := range users {
		// Check if the user exists on the system
		_, ok := existingUsers[u.Name]
		if !ok {
			// User does not exist, create the user
			err := createUser(u.Name)
			if err != nil {
				Error("Error creating user", "user", u.Name, "error", err)
				return err
			}
		}
		existing, err := user.Lookup(u.Name)
		if err != nil {
			Error("Error looking up user", "user", u.Name, "error", err)
			return err
		}

		// Update SSH keys for the user
		uid, _ := strconv.Atoi(existing.Uid)
		gid, _ := strconv.Atoi(existing.Gid)
		Info("Updating SSH keys", "user", u.Name, "keys", u.Keys, "uid", uid, "gid", gid)
		err = updateSSHKeys(u.Name, u.Keys, uid, gid)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateSSHKeys updates the SSH keys for a given user on the Linux system
func updateSSHKeys(username string, keys []string, uid int, gid int) error {
	homeDir, err := getHomeDir(username)
	Info("Updating SSH keys", "user", username, "homeDir", homeDir)
	if err != nil {
		return err
	}

	// Write SSH keys to the authorized_keys file
	authorizedKeysPath := fmt.Sprintf("%s/.ssh/authorized_keys", homeDir)
	sshDir := fmt.Sprintf("%s/.ssh", homeDir)
	createErr := os.MkdirAll(sshDir, 0700)
	chownErr := os.Chown(sshDir, uid, gid)
	if createErr != nil {
		Error("Error creating .ssh directory", "path", sshDir, "error", createErr)
	}
	if chownErr != nil {
		Error("Error changing ownership of .ssh directory", "path", sshDir, "error", chownErr)
	}

	err = os.WriteFile(authorizedKeysPath, []byte(strings.Join(keys, "\n")), 0644)
	chownErr = os.Chown(authorizedKeysPath, uid, gid)
	if chownErr != nil {
		Error("Error changing ownership of authorized_keys file", "path", authorizedKeysPath, "error", chownErr)
	}
	if err != nil {
		Error("Error writing SSH keys", "path", authorizedKeysPath, "error", err)
		return err
	}

	return nil
}

// getHomeDir retrieves the home directory for a given user
func getHomeDir(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}
