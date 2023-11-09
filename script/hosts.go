package script

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
	"github.com/ethaniccc/simple-osharden/utils"
)

func init() {
	RegisterScript(&VerifyHosts{})
}

// VerifyHosts prompts the user to ensure that all hosts in the host file are allowed.
type VerifyHosts struct {
}

func (s *VerifyHosts) Name() string {
	return "vfhosts"
}

func (s *VerifyHosts) Description() string {
	return "Verify all hosts are allowed on the machine."
}

func (s *VerifyHosts) RunOnLinux() error {
	buffer, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return fmt.Errorf("unable to read /etc/hosts: %s", err.Error())
	}

	badHosts := []string{}
	lines := strings.Split(string(buffer), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		// Get the elements of the line.
		elems := strings.Split(line, " ")
		ip := elems[0]
		// Try to parse the IP address of the first element. If it is not an IP address, we will ignore it.
		if addr := net.ParseIP(ip); addr == nil {
			continue
		}

		if prompts.Confirm(fmt.Sprintf("Should %s redirect to [%s]?", ip, strings.Join(elems[1:], " "))) {
			continue
		}

		// Add the IP to the list of bad hosts.
		badHosts = append(badHosts, ip)
	}

	// Remove the unwanted hosts from /etc/hosts
	if err := utils.DelOptsFromFile(badHosts, " ", "/etc/hosts"); err != nil {
		return err
	}

	return nil
}
