package script

import (
	"fmt"
	"os"

	"github.com/ethaniccc/simple-osharden/prompts"
)

func init() {
	RegisterScript(&AllowedUsers{})
}

// AllowedUsers is a script that goes through every user on the machine, and prompts the
// user to verify that they should be on the machine. If the user is not allowed, they
// will be removed from the machine.
type AllowedUsers struct {
}

func (s *AllowedUsers) Name() string {
	return "users-allowed"
}

func (s *AllowedUsers) Description() string {
	return "Verify all users are allowed on the machine."
}

func (s *AllowedUsers) Run() error {
	// Scan the home directory.
	entries, err := os.ReadDir("/home")
	if err != nil {
		return fmt.Errorf("unable to scan /home: %s", err.Error())
	}

	// Iterate through each entry in the directory.
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		user := entry.Name()
		if prompts.Confirm(fmt.Sprintf("Is the user %s allowed on this machine?", user)) {
			continue
		}

		// Remove the user from the machine.
		if err := RunCommand(fmt.Sprintf("deluser --remove-home %s", user)); err != nil {
			return fmt.Errorf("unable to remove user %s: %s", user, err.Error())
		}
	}

	return nil
}
