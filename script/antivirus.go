package script

import (
	"fmt"
	"strings"

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
	// Install the antivirus.
	if err := RunCommand("apt install clamav"); err != nil {
		return fmt.Errorf("unable to install antivirus: %s", err.Error())
	}

	// Update the antivirus.
	if err := RunCommand("freshclam"); err != nil {
		return fmt.Errorf("unable to update antivirus: %s", err.Error())
	}

	// Scan the machine.
	if err := RunCommand("clamscan -r /"); err != nil {
		return fmt.Errorf("unable to scan machine: %s", err.Error())
	}

	return nil
}

func (s *RunAntivirus) RunOnWindows() error {
	var scanType string
	for scanType != "" {
		switch strings.ToLower(prompts.RawResponsePrompt("What type of scan would you like to run? [Quick, Full, Custom]")) {
		case "quick":
			scanType = "QuickScan"
		case "full":
			scanType = "FullScan"
		case "custom":
			scanType = "CustomScan"
		default:
			logger.Errorf("%s is not a valid scan type.", scanType)
		}
	}

	if err := RunCommand(fmt.Sprintf("powershell.exe -Command \"Start-MpScan -ScanType %s\"", scanType)); err != nil {
		return fmt.Errorf("unable to scan machine: %s", err.Error())
	}

	return nil
}
