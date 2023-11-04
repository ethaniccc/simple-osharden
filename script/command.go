package script

import (
	"os"
	"os/exec"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

// LoggedCommand is a command wrapper that includes a log message.
type LoggedCommand struct {
	LogMessage string
	Command    string
	IgnoreErr  bool
}

// ExecuteLoggedCommands executes a list of LoggedCommand instances.
func ExecuteLoggedCommands(cmds []LoggedCommand) error {
	for _, cmd := range cmds {
		logger.Info(cmd.LogMessage)
		err := RunCommand(cmd.Command)
		if err == nil {
			continue
		}

		if cmd.IgnoreErr {
			logger.Warnf("Error running command (%s): %s", cmd.Command, err)
			continue
		}

		return err
	}

	return nil
}

// CreateCommand creates a new exec.Cmd instance.
func CreateCommand(c string, args ...string) *exec.Cmd {
	cmd := exec.Command(c, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// RunCommand runs a command.
func RunCommand(c string) error {
	split := strings.Split(c, " ")
	return CreateCommand(split[0], split[1:]...).Run()
}

// GetCommandOutput runs a command and returns the output.
func GetCommandOutput(c string) (string, error) {
	split := strings.Split(c, " ")
	cmd := exec.Command(split[0], split[1:]...)
	out, err := cmd.Output()
	return string(out), err
}

// ConfirmCommand runs a command if the user confirms it should be run.
func ConfirmCommand(msg, c string) error {
	if !prompts.Confirm(msg) {
		return nil
	}

	return RunCommand(c)
}
