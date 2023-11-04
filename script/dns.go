package script

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

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

	buffer, err := os.ReadFile(file)
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
		if err := os.WriteFile(file, []byte(data), 0644); err != nil {
			return fmt.Errorf("unable to write /etc/resolv.conf: %s", err.Error())
		}

		return nil
	}

	first := true
	for {
		if !first && !prompts.Confirm("Would you like to add another DNS server?") {
			break
		}
		first = false

		newServer := prompts.RawResponsePrompt("DNS server IP address")
		if newServer == "" {
			break
		}

		// Validate the IP address.
		if ip := net.ParseIP(newServer); ip == nil {
			logger.Errorf("Invalid IP address: %s", newServer)
			continue
		}

		data += fmt.Sprintf("nameserver %s\n", newServer)
	}

	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		return fmt.Errorf("unable to write /etc/resolv.conf: %s", err.Error())
	}

	return nil
}