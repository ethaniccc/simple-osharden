package script

import (
	"time"

	"github.com/ethaniccc/simple-osharden/prompts"
	"github.com/ethaniccc/simple-osharden/utils"
)

func init() {
	RegisterScript(&NetworkSetup{})
}

// NetworkSetup is a script that installs and enables UFW, and sets other network settings.
// By default, it allows SSH connections, and denies all other incoming connections.
// All outgoing connections are allowed by default.
type NetworkSetup struct {
}

func (s *NetworkSetup) Name() string {
	return "net-setup"
}

func (s *NetworkSetup) Description() string {
	return "Installs and configures firewall, and sets other network settings."
}

func (s *NetworkSetup) Run() error {
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
