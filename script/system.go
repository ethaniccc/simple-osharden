package script

import (
	"fmt"

	"github.com/ethaniccc/simple-osharden/utils"
)

func init() {
	RegisterScript(&SystemConfiguration{})
}

// SystemConfiguration sets certain settings in /etc/sysctl.conf to further secure the system.
type SystemConfiguration struct {
}

func (s *SystemConfiguration) Name() string {
	return "syscfg"
}

func (s *SystemConfiguration) Description() string {
	return "Configures settings in sysctl.conf to secure the system."
}

func (s *SystemConfiguration) RunOnLinux() error {
	if err := utils.WriteOptsToFile(map[string]string{
		"fs.suid_dumpable":          "0",
		"kernel.randomize_va_space": "2",
		"kernel.exec-shield":        "1",
	}, " = ", "/etc/sysctl.conf"); err != nil {
		return fmt.Errorf("unable to write to /etc/sysctl.conf: %s", err.Error())
	}

	if err := utils.WriteOptsToFile(map[string]string{
		"local_events": "yes",
	}, " = ", "/etc/audit/auditd.conf"); err != nil {
		return fmt.Errorf("unable to write to /etc/audit/auditd.conf: %s", err.Error())
	}

	return nil
}
