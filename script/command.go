package script

import (
	"os"
	"os/exec"
)

// CreateCommand creates a new exec.Cmd instance.
func CreateCommand(c string, args ...string) *exec.Cmd {
	cmd := exec.Command(c, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
