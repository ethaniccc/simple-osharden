package script

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethaniccc/simple-osharden/prompts"
	"github.com/ethaniccc/simple-osharden/utils"
)

func init() {
	RegisterScript(&NetworkSetup{})
	RegisterScript(&NetworkApps{})
}

// NetworkSetup is a script that installs and enables UFW, and sets other network settings.
// By default, it allows SSH connections, and denies all other incoming connections.
// All outgoing connections are allowed by default.
type NetworkSetup struct {
}

func (s *NetworkSetup) Name() string {
	return "netsetup"
}

func (s *NetworkSetup) Description() string {
	return "Installs and configures firewall, and sets other network settings."
}

func (s *NetworkSetup) RunOnLinux() error {
	if err := ExecuteLoggedCommands([]LoggedCommand{
		{"Installing UFW", "apt install ufw", true},
		{"Enabling UFW Firewall", "ufw enable", false},
		{"Allowing SSH through firewall", "ufw allow openssh", false},
		{"Setting option to deny incoming connections by default", "ufw default deny incoming", false},
		{"Setting option to allow outgoing connections by default", "ufw default allow outgoing", false},
	}); err != nil {
		return err
	}

	networkOpts := map[string]string{}

	if prompts.Confirm("Would you like to enable TCP SYN cookies?") {
		networkOpts["net.ipv4.tcp_syncookies"] = "1"
	}

	if prompts.Confirm("Would you like to enable IPv4 TIME-WAIT ASSASSINATION protection?") {
		networkOpts["net.ipv4.tcp_rfc1337"] = "1"
	}

	if prompts.Confirm("Would you like to disable IPv4 forwarding?") {
		networkOpts["net.ipv4.ip_forward"] = "0"
	}

	if prompts.Confirm("Would you like to disable source packet routing?") {
		networkOpts["net.ipv4.conf.all.accept_source_route"] = "0"
		networkOpts["net.ipv4.conf.default.accept_source_route"] = "0"
	}

	if prompts.Confirm("Would you like to disable send redirects?") {
		networkOpts["net.ipv4.conf.all.send_redirects"] = "0"
		networkOpts["net.ipv4.conf.default.send_redirects"] = "0"
	}

	if prompts.Confirm("Would you like to disable Martian packet logging?") {
		networkOpts["net.ipv4.conf.all.log_martians"] = "1"
	}

	if prompts.Confirm("Would you like to enable source address verification?") {
		networkOpts["net.ipv4.conf.all.rp_filter"] = "1"
		networkOpts["net.ipv4.conf.default.rp_filter"] = "1"
	}

	if prompts.Confirm("Would you like to ignore ICMP redirects?") {
		networkOpts["net.ipv4.conf.all.accept_redirects"] = "0"
		networkOpts["net.ipv4.conf.default.accept_redirects"] = "0"
	}

	if prompts.Confirm("Would you like to disable IPv6?") {
		networkOpts["net.ipv6.conf.all.disable_ipv6"] = "1"
		networkOpts["net.ipv6.conf.default.disable_ipv6"] = "1"
	}

	if err := utils.WriteOptsToFile(networkOpts, " = ", "/etc/sysctl.conf"); err != nil {
		return err
	}

	logger.Warnf("--------------- IMPORTANT ---------------")
	logger.Warnf("For changes to be applied, please restart the machine.")
	logger.Warnf("--------------- IMPORTANT ---------------")
	<-time.After(time.Second * 3)

	return nil
}

func (s *NetworkSetup) RunOnWindows() error {
	commands := []LoggedCommand{
		{"Enabling Windows Firewall", "netsh advfirewall set allprofiles state on", false},
	}

	if prompts.Confirm("Disable inbound connections by default?") {
		commands = append(commands, LoggedCommand{"Disabling inbound connections by default", "netsh advfirewall set allprofiles firewallpolicy blockinbound,allowoutbound", false})
	}

	if err := ExecuteLoggedCommands(commands); err != nil {
		return fmt.Errorf("unable to set up network: %s", err.Error())
	}

	return nil
}

// NetworkApps is a script that checks for any applications that are listening on ports.
// The user will confirm that these applications are allowed to listen on these ports, and if they're
// not, the process will be killed. The user will also be prompted if the program should be
// removed from the machine.
type NetworkApps struct {
}

func (s *NetworkApps) Name() string {
	return "netapps"
}

func (s *NetworkApps) Description() string {
	return "Checks for any applications that are listening on ports."
}

func (s *NetworkApps) RunOnLinux() error {
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

		if err := s.appUninstallLinux(proc); err != nil {
			return fmt.Errorf("unable to uninstall program: %s", err.Error())
		}
	}

	ResetTerminal()
	if err := RunCommand("apt autoremove"); err != nil {
		return fmt.Errorf("unable to autoremove packages: %s", err.Error())
	}

	return nil
}

// appUninstallLinux will uninstall the program on linux and remove any traces of it.
func (s *NetworkApps) appUninstallLinux(program string) error {
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
