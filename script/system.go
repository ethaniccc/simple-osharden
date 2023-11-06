package script

import "github.com/ethaniccc/simple-osharden/utils"

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
	return utils.WriteOptsToFile(map[string]string{
		"fs.suid_dumpable":          "0",
		"kernel.randomize_va_space": "2",
		"kernel.exec-shield":        "1",
	}, " = ", "/etc/sysctl.conf")
}
