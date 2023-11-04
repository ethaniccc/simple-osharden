package script

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethaniccc/simple-osharden/prompts"
)

// Network is a script that installs and enables UFW, and sets other network settings.
// By default, it allows SSH connections, and rejects all other incoming connections.
// All outgoing connections are allowed by default.
type Network struct {
}

func (f *Network) Name() string {
	return "network-setup"
}

func (f *Network) Description() string {
	return "Installs and configures firewall, and sets other network settings."
}

func (f *Network) Run() error {
	if err := ExecuteLoggedCommands([]LoggedCommand{
		{"Installing UFW", "apt install ufw", true},
		{"Enabling UFW Firewall", "ufw enable", false},
		{"Allowing SSH through firewall", "ufw allow ssh", false},
		{"Setting option to reject incoming connections by default", "ufw default reject incoming", false},
		{"Setting option to allow outgoing connections by default", "ufw default allow outgoing", false},
	}); err != nil {
		return err
	}

	networkOpts := map[string]string{}

	buffer, err := os.ReadFile("/etc/sysctl.conf")
	if err != nil {
		return fmt.Errorf("unable to read /etc/sysctl.conf: %s", err.Error())
	}
	data := string(buffer)

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

		if err := RunCommand("ufw disable ipv6"); err != nil {
			return fmt.Errorf("unable to disable IPv6 in UFW: %s", err.Error())
		}
	}

	lines := strings.Split(data, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		// The current option on the line.
		cOpt := strings.Split(strings.ReplaceAll(line, " ", ""), "=")[0]
		if newVal, ok := networkOpts[cOpt]; ok {
			lines[i] = cOpt + " = " + newVal
			delete(networkOpts, cOpt)
		}
	}

	// Add any new options that didn't previously exist on sysctl.conf
	for opt, val := range networkOpts {
		lines = append(lines, opt+" = "+val)
	}

	if err := os.WriteFile("/etc/sysctl.conf", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write new data to /etc/sysctl.conf: %s", err.Error())
	}

	logger.Warnf("--------------- IMPORTANT ---------------")
	logger.Warnf("For changes to be applied, please restart the machine.")
	logger.Warnf("--------------- IMPORTANT ---------------")
	<-time.After(time.Second * 3)

	return nil
}
