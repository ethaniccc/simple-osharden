package script

import "github.com/ethaniccc/simple-osharden/prompts"

func init() {
	RegisterScript(&RemovePrograms{})
	RegisterScript(&UpdatePrograms{})
}

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
	return "programs-update"
}

func (s *UpdatePrograms) Description() string {
	return "Updates programs on the system."
}

func (s *UpdatePrograms) RunOnLinux() error {
	return RunCommand("apt upgrade")
}
