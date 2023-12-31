package script

import (
	"fmt"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

func init() {
	RegisterScript(&RemovePrograms{})
	RegisterScript(&UpdatePrograms{})
}

// RemovePrograms is a script that removes programs that may increase the attack surface of the system.
type RemovePrograms struct {
}

func (s *RemovePrograms) Name() string {
	return "rmprograms"
}

func (s *RemovePrograms) Description() string {
	return "Removes programs that may increase the attack surface of the system."
}

func (s *RemovePrograms) RunOnLinux() error {
	commands := []LoggedCommand{}

	if prompts.Confirm("Would you like to uninstall wireshark?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling Wireshark",
			Command:    "apt remove wireshark",
			IgnoreErr:  true,
		})
	}

	if prompts.Confirm("Would you like to uninstall ophcrack?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling ophcrack",
			Command:    "apt remove ophcrack",
			IgnoreErr:  true,
		})
	}

	if prompts.Confirm("Would you like to uninstall john?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling john",
			Command:    "apt remove john",
			IgnoreErr:  true,
		})
	}

	if prompts.Confirm("Would you like to uninstall hydra?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling hydra",
			Command:    "apt remove hydra",
			IgnoreErr:  true,
		})
	}

	if prompts.Confirm("Would you like to uninstall nmap?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling nmap",
			Command:    "apt remove nmap",
			IgnoreErr:  true,
		})
	}

	if prompts.Confirm("Would you like to uninstall snort?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling snort",
			Command:    "apt remove snort",
			IgnoreErr:  true,
		})
	}

	if prompts.Confirm("Would you like to uninstall netcat?") {
		commands = append(commands, LoggedCommand{
			LogMessage: "Uninstalling netcat",
			Command:    "apt remove netcat",
			IgnoreErr:  true,
		})
	}

	commands = append(commands, LoggedCommand{
		LogMessage: "Removing unused packages",
		Command:    "apt autoremove",
		IgnoreErr:  true,
	})

	ResetTerminal()
	return ExecuteLoggedCommands(commands)
}

// UpdatePrograms is a script that updates programs on the system. This is a very simple script
// that essentially only runs `apt upgrade`.
type UpdatePrograms struct {
}

func (s *UpdatePrograms) Name() string {
	return "updateprograms"
}

func (s *UpdatePrograms) Description() string {
	return "Updates programs on the system."
}

func (s *UpdatePrograms) RunOnLinux() error {
	ResetTerminal()
	return RunCommand("apt upgrade")
}

// appUninstallLinux will uninstall the program on linux and remove any traces of it.
func AppUninstallLinux(program string) error {
	// Uninstall the program.
	ResetTerminal()
	RunCommand(fmt.Sprintf("apt purge %s", program))

	// Find any traces of the program and remove them.
	dat, err := GetCommandOutput(fmt.Sprintf("find / -name \"%s\"", program))
	if err != nil {
		return fmt.Errorf("unable to find traces of program: %s", err.Error())
	}

	for _, line := range strings.Split(dat, "\n") {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "find:") {
			continue
		}

		if err := RunCommand(fmt.Sprintf("rm -rf %s", line)); err != nil {
			return fmt.Errorf("unable to remove %s at [%s]: %s", program, line, err.Error())
		}
	}

	return nil
}
