package script

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/ethaniccc/simple-osharden/prompts"
)

func init() {
	RegisterScript(&RunAntivirus{})
}

// RunAntivirus is a script that installs an antivirus and then runs it on the machine.
type RunAntivirus struct {
}

func (s *RunAntivirus) Name() string {
	return "runav"
}

func (s *RunAntivirus) Description() string {
	return "Run the antivirus."
}

func (s *RunAntivirus) RunOnLinux() error {
	ResetTerminal()

	// Install the antivirus.
	if err := RunCommand("apt install clamav"); err != nil {
		return fmt.Errorf("unable to install antivirus: %s", err.Error())
	}

	// Update the antivirus.
	if err := RunCommand("freshclam"); err != nil {
		logger.Errorf("unable to update antivirus definitions")
		if !prompts.Confirm("Would you still like to continue?") {
			return nil
		}
	}

	ResetTerminal()

	// Scan the machine.
	if err := RunCommand("clamscan -r /"); err != nil {
		return fmt.Errorf("unable to scan machine: %s", err.Error())
	}

	return nil
}

func (s *RunAntivirus) RunOnWindows() error {
	scanType := prompt.Input("Select scan type: ", func(d prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix([]prompt.Suggest{
			{Text: "Quick", Description: "Runs a quick antivirus scan."},
			{Text: "Full", Description: "Runs a full antivirus scan."},
			{Text: "Custom", Description: "Runs a custom antivirus scan."},
		}, d.GetWordBeforeCursor(), true)
	}, prompts.DummyPromptOption)
	if scanType != "Quick" && scanType != "Full" && scanType != "Custom" {
		logger.Error("Invalid scan type - please select a valid scan type.")
		s.RunOnWindows()
		return nil
	}

	if err := RunCommand(fmt.Sprintf("powershell.exe -Command \"Start-MpScan -ScanType %s\"", scanType)); err != nil {
		return fmt.Errorf("unable to scan machine: %s", err.Error())
	}

	return nil
}
