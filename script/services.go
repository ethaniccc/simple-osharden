package script

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

type ServiceConfiguration struct {
}

func (s *ServiceConfiguration) Name() string {
	return "service-config"
}

func (s *ServiceConfiguration) Description() string {
	return "Configures services to be more secure."
}

func (s *ServiceConfiguration) Run() error {
	if prompts.Confirm("Would you like to configure the SSH service?") {
		if err := s.configureSSH(); err != nil {
			return err
		}
	}

	if prompts.Confirm("Would you like to configure the FTP service?") {
		if err := s.configureFTP(); err != nil {
			return err
		}
	}

	return nil
}

// initService initializes a service.
func (s *ServiceConfiguration) initService(service string) (bool, error) {
	// Check if the user wants the service to be running.
	if !prompts.Confirm(fmt.Sprintf("Should %s be enabled on this machine?", service)) {
		logger.Warnf("stopping %s", service)
		RunCommand("systemctl stop " + service)

		logger.Warnf("disabling %s", service)
		RunCommand("systemctl disable " + service)

		return false, nil
	}

	RunCommand("systemctl enable " + service)
	res, err := GetCommandOutput("systemctl status " + service)
	if err != nil {
		return false, fmt.Errorf("unable to get status of %s: %s", service, err.Error())
	}

	if !strings.Contains(res, "active (running)") {
		logger.Warnf("starting %s (currently detected as not running)", service)
		RunCommand("systemctl start " + service)
	}

	return true, nil
}

func (s *ServiceConfiguration) configureFTP() error {
	c, err := s.initService("vsftpd")
	if err != nil {
		return fmt.Errorf("unable to initialize vsftpd service: %s", err.Error())
	}

	if !c {
		return nil
	}

	if err := RunCommand("ufw allow vsftpd"); err != nil {
		return fmt.Errorf("unable to allow ftp through firewall: %s", err.Error())
	}

	buffer, err := os.ReadFile("/etc/vsftpd.conf")
	if err != nil {
		return fmt.Errorf("unable to read vsftpd.conf: %s", err.Error())
	}

	ftpOpts := map[string]string{}
	if prompts.Confirm("Would you like to allow anonymous users?") {
		ftpOpts["anonymous_enable"] = "YES"
	} else {
		ftpOpts["anonymous_enable"] = "NO"
	}

	if prompts.Confirm("Should the FTP use TLS?") {
		ftpOpts["ssl_enable"] = "YES"
		ftpOpts["ssl_tlsv1"] = "YES"
		ftpOpts["ssl_sslv2"] = "YES"
		ftpOpts["ssl_sslv3"] = "YES"
	} else {
		ftpOpts["ssl_enable"] = "NO"
		ftpOpts["ssl_tlsv1"] = "NO"
		ftpOpts["ssl_sslv2"] = "NO"
		ftpOpts["ssl_sslv3"] = "NO"
	}

	if prompts.Confirm("Should anonymous TLS/SSL be enabled?") {
		ftpOpts["allow_anon_ssl"] = "YES"
	} else {
		ftpOpts["allow_anon_ssl"] = "NO"
	}

	if prompts.Confirm("Should a passive port range be set?") {
		minPort := prompts.RawResponsePrompt("What should the minimum port be?")
		maxPort := prompts.RawResponsePrompt("What should the maximum port be?")
		ftpOpts["pasv_min_port"] = minPort
		ftpOpts["pasv_max_port"] = maxPort

		RunCommand("ufw allow " + minPort + ":" + maxPort + "/tcp")
	}

	lines := strings.Split(string(buffer), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		cOpt := strings.ReplaceAll(strings.Split(line, "=")[0], "#", "")
		if newVal, ok := ftpOpts[cOpt]; ok {
			lines[i] = fmt.Sprintf("%s=%s", cOpt, newVal)
			delete(ftpOpts, cOpt)
		}
	}

	// Add any new options that were not already in the config file.
	for opt, val := range ftpOpts {
		lines = append(lines, opt+"="+val)
	}

	if err := os.WriteFile("/etc/vsftpd.conf", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write vsftpd.conf: %s", err.Error())
	}
	logger.Info("Data written to vsftpd.conf")

	return nil
}

func (s *ServiceConfiguration) configureSSH() error {
	c, err := s.initService("ssh")
	if err != nil {
		return fmt.Errorf("unable to initialize ssh service: %s", err.Error())
	}

	// Don't continue. This is because the user does not want to use this service on their machine.
	if !c {
		return nil
	}

	if err := RunCommand("ufw allow openssh"); err != nil {
		return fmt.Errorf("unable to allow ssh through firewall: %s", err.Error())
	}

	buffer, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return fmt.Errorf("unable to read sshd_config: %s", err.Error())
	}

	sshOpts := map[string]string{}
	if prompts.Confirm("Would you like to use root login?") {
		sshOpts["PermitRootLogin"] = "yes"
	} else {
		sshOpts["PermitRootLogin"] = "no"
	}

	if prompts.Confirm("Would you like to use password authentication?") {
		sshOpts["PasswordAuthentication"] = "yes"
	} else {
		sshOpts["PasswordAuthentication"] = "no"
	}

	if res := prompts.RawResponsePrompt("What port should SSH listen on? (default is 22)"); res != "" {
		sshOpts["Port"] = res
	}

	if len(sshOpts) == 0 {
		return nil
	}

	lines := strings.Split(string(buffer), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		split := strings.Split(line, " ")
		if len(split) < 2 {
			continue
		}

		key := strings.ReplaceAll(split[0], "#", "")
		if _, ok := sshOpts[key]; !ok {
			continue
		}

		lines[i] = fmt.Sprintf("%s %s", key, sshOpts[key])
		delete(sshOpts, key)
	}

	for opt, val := range sshOpts {
		lines = append(lines, fmt.Sprintf("%s %s", opt, val))
	}

	if err := os.WriteFile("/etc/ssh/sshd_config", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write sshd_config: %s", err.Error())
	}
	logger.Info("Data written to sshd_config")

	if prompts.Confirm("Would you like to restart the SSH service?") {
		RunCommand("systemctl restart sshd")
		logger.Info("SSH service restarted")
	}

	return nil
}
