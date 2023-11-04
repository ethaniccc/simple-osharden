package script

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

// Firewall is a script that installs UFW and configures it. By default, it allows SSH
// connections, and rejects all other incoming connections. All outgoing connections
// are allowed by default.
type Firewall struct {
}

func (f *Firewall) Name() string {
	return "firewall"
}

func (f *Firewall) Description() string {
	return "Installs UFW and configures it."
}

func (f *Firewall) Run() error {
	return ExecuteLoggedCommands([]LoggedCommand{
		{"Installing UFW", "apt install ufw", true},
		{"Enabling UFW Firewall", "ufw enable", true},
		{"Allowing SSH through firewall", "ufw allow ssh", false},
		{"Setting option to reject incoming connections by default", "ufw default reject incoming", false},
		{"Setting option to allow outgoing connections by default", "ufw default allow outgoing", false},
	})
}

// UpdateDNS is a script that checks for the current DNS servers. The user is then able
// to confirm wether or not they want to keep that DNS server. If they do not, the DNS server
// is removed. The user is then able to add new DNS servers if they wish to.
type UpdateDNS struct {
}

func (u *UpdateDNS) Name() string {
	return "dns-update"
}

func (u *UpdateDNS) Description() string {
	return "Checks and updates the DNS servers."
}

func (u *UpdateDNS) Run() error {
	file := "/etc/resolv.conf"
	f, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("unable to open /etc/resolv.conf: %s", err.Error())
	}
	defer f.Close()

	// Read the contents of the file.
	buffer, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("unable to read /etc/resolv.conf: %s", err.Error())
	}
	data := string(buffer)

	// Print the current DNS servers.
	for _, line := range strings.Split(data, "\n") {
		if !strings.HasPrefix(line, "nameserver") {
			continue
		}

		logger.Infof("Found DNS server: %s", strings.ReplaceAll(line, "nameserver ", ""))
		if !prompts.Confirm("Would you like to remove this DNS server?") {
			continue
		}

		// Remove the DNS server.
		data = strings.ReplaceAll(data, line+"\n", "")
	}

	// Ask the user if they want to add additional DNS servers.
	if !prompts.Confirm("Would you like to add new DNS servers?") {
		// Write the new data to the file, since DNS servers could have been removed.
		return os.WriteFile(file, []byte(data), 0644)
	}

	newServers := []string{}
	for {
		if len(newServers) != 0 && !prompts.Confirm("Would you like to add another DNS server?") {
			break
		}

		newServer := prompts.RawResponsePrompt("DNS server IP address")
		if newServer == "" {
			break
		}

		// Validate the IP address.
		if ip := net.ParseIP(newServer); ip == nil {
			logger.Errorf("Invalid IP address: %s", newServer)
			continue
		}

		newServers = append(newServers, newServer)
	}

	// Write the new DNS servers to the file.
	for _, server := range newServers {
		data += fmt.Sprintf("nameserver %s\n", server)
	}

	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		return fmt.Errorf("unable to write /etc/resolv.conf: %s", err.Error())
	}

	return nil
}
