package script

import (
	"fmt"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

func init() {
	RegisterScript(&NetApps{})
}

// NetApps is a script that checks for any applications that are listening on ports.
// The user will confirm that these applications are allowed to listen on these ports, and if they're
// not, the process will be killed. The user will also be prompted if the program should be
// removed from the machine.
type NetApps struct {
}

func (s *NetApps) Name() string {
	return "net-apps"
}

func (s *NetApps) Description() string {
	return "Checks for any applications that are listening on ports."
}

func (s *NetApps) Run() error {
	output, err := GetCommandOutput("netstat -tunlpw")
	if err != nil {
		return fmt.Errorf("unable to get netstat data: %s", err.Error())
	}

	for _, line := range strings.Split(output, "\n")[2:] {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		// Parse each line from the data given by netstat.
		ip := fields[3]
		split := strings.Split(fields[6], "/")
		if len(split) != 2 {
			continue
		}

		pid, proc := split[0], strings.Split(strings.Split(split[1], " ")[0], ":")[0]

		if prompts.Confirm(fmt.Sprintf("Should the process %s (pid=%s) be listening on %s", proc, pid, ip)) {
			continue
		}

		// Ask the user if they want to kill the process.
		kill := prompts.Confirm(fmt.Sprintf("Should we kill the process %s (pid=%s)", proc, pid))
		// Ask the user if they want to uninstall all instances of the program (malware).
		uninstall := prompts.Confirm(fmt.Sprintf("Should we try to uninstall all instances of %s", proc))

		if kill {
			RunCommand(fmt.Sprintf("kill %s", pid))
		}

		if !uninstall {
			continue
		}

		if err := s.uninstall(proc); err != nil {
			return fmt.Errorf("unable to uninstall program: %s", err.Error())
		}
	}

	RunCommand("reset")
	if err := RunCommand("apt autoremove"); err != nil {
		return fmt.Errorf("unable to autoremove packages: %s", err.Error())
	}

	return nil
}

// uninstall will uninstall the program and remove any traces of it.
func (s *NetApps) uninstall(program string) error {
	// Uninstall the program.
	RunCommand("reset")
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
