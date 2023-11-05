package script

import (
	"fmt"
	"os"
	"strings"
)

func init() {
	RegisterScript(&SystemConfiguration{})
}

type SystemConfiguration struct {
}

func (s *SystemConfiguration) Name() string {
	return "syscfg"
}

func (s *SystemConfiguration) Description() string {
	return "Configures settings in sysctl.conf to secure the system."
}

func (s *SystemConfiguration) Run() error {
	buffer, err := os.ReadFile("/etc/sysctl.conf")
	if err != nil {
		return fmt.Errorf("unable to open /etc/sysctl.conf: %s", err.Error())
	}

	data := string(buffer)
	options := map[string]string{
		"fs.suid_dumpable":          "0",
		"kernel.randomize_va_space": "2",
		"kernel.exec-shield":        "1",
	}

	logger.Info("Set fs.suid_dumpable to 0")
	logger.Info("Set kernel.randomize_va_space to 2")
	logger.Info("Set kernel.exec-shield to 1")

	lines := strings.Split(data, "\n")
	for i, line := range lines {
		cOpt := strings.ReplaceAll(strings.Split(strings.ReplaceAll(line, " ", ""), "=")[0], "#", "")
		if newVal, ok := options[cOpt]; ok {
			lines[i] = cOpt + " = " + newVal
			delete(options, cOpt)
		}
	}

	for opt, val := range options {
		lines = append(lines, opt+" = "+val)
	}

	if err := os.WriteFile("/etc/sysctl.conf", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write new data to /etc/sysctl.conf: %s", err.Error())
	}

	return nil
}
