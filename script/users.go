package script

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

func init() {
	RegisterScript(&VerifyUsers{})
}

// VerifyUsers is a script that goes through every user on the machine, and prompts the
// user to verify that they should be on the machine. If the user is not allowed, they
// will be removed from the machine.
type VerifyUsers struct {
}

func (s *VerifyUsers) Name() string {
	return "vfusers"
}

func (s *VerifyUsers) Description() string {
	return "Verify all users are allowed on the machine."
}

func (s *VerifyUsers) RunOnLinux() error {
	// Scan the home directory.
	entries, err := os.ReadDir("/home")
	if err != nil {
		return fmt.Errorf("unable to scan /home: %s", err.Error())
	}

	// Iterate through each entry in the /home/ directory.
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		user := entry.Name()
		if prompts.Confirm(fmt.Sprintf("Is the user %s allowed on this machine?", user)) {
			groups, err := GetCommandOutput(fmt.Sprintf("groups %s", user))
			if err != nil {
				return fmt.Errorf("unable to get groups for user %s: %s", user, err.Error())
			}

			hasAdmin := strings.Contains(groups, "sudo")

			// Ask if this user is an administrator.
			if prompts.Confirm(fmt.Sprintf("Is the user %s an admin?", user)) {
				if !hasAdmin {
					logger.Warnf("Adding %s to sudoers", user)
					RunCommand(fmt.Sprintf("adduser %s sudo", user))
				}

				continue
			}

			// If the user shouldn't be an admin, remove their sudo access if they have it.
			if hasAdmin {
				logger.Warnf("Removing %s from sudoers", user)
				RunCommand(fmt.Sprintf("deluser %s sudo", user))
				continue
			}

			continue
		}

		// Remove the user from the machine.
		if err := RunCommand(fmt.Sprintf("deluser --remove-home %s", user)); err != nil {
			return fmt.Errorf("unable to remove user %s: %s", user, err.Error())
		}
	}

	return nil
}

func (s *VerifyUsers) RunOnWindows() error {
	enteries, err := os.ReadDir("C:\\Users")
	if err != nil {
		return fmt.Errorf("unable to scan C:\\Users: %s", err.Error())
	}

	for _, entry := range enteries {
		if !entry.IsDir() {
			continue
		}

		user := entry.Name()
		// Ignore the default users.
		if user == "Public" || user == "Default" || user == "Default User" {
			continue
		}

		if prompts.Confirm(fmt.Sprintf("Is the user %s allowed on this machine?", user)) {
			groups, err := GetCommandOutput(fmt.Sprintf("net user %s", user))
			if err != nil {
				return fmt.Errorf("unable to get groups for user %s: %s", user, err.Error())
			}

			hasAdmin := strings.Contains(groups, "Administrators")

			// Ask if this user is an administrator.
			if prompts.Confirm(fmt.Sprintf("Is the user %s an admin?", user)) {
				if !hasAdmin {
					logger.Warnf("Adding %s to administrators", user)
					RunCommand(fmt.Sprintf("net localgroup administrators %s /add", user))
				}

				continue
			}

			// If the user shouldn't be an admin, remove their sudo access if they have it.
			if hasAdmin {
				logger.Warnf("Removing %s from administrators", user)
				RunCommand(fmt.Sprintf("net localgroup administrators %s /delete", user))
				continue
			}

			continue
		}
	}

	return nil
}
