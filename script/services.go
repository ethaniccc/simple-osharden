package script

import (
	"fmt"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
	"github.com/ethaniccc/simple-osharden/utils"
)

func init() {
	RegisterScript(&ServiceConfiguration{})
}

type ServiceConfiguration struct {
}

func (s *ServiceConfiguration) Name() string {
	return "servicecfg"
}

func (s *ServiceConfiguration) Description() string {
	return "Configures services to be more secure."
}

func (s *ServiceConfiguration) RunOnLinux() error {
	if prompts.Confirm("Would you like to configure SSH?") {
		if err := s.configureSSH(); err != nil {
			return err
		}
	}

	if prompts.Confirm("Would you like to configure FTP?") {
		if err := s.configureFTP(); err != nil {
			return err
		}
	}

	if prompts.Confirm("Would you like to configure Apache2?") {
		if err := s.configureApache(); err != nil {
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
	RunCommand("systemctl start " + service)
	res, err := GetCommandOutput("systemctl status " + service)
	if err != nil {
		logger.Warnf("unable to get status of %s: it is possible the service does not exist on this machine", service)
		return false, nil
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

	return utils.WriteOptsToFile(ftpOpts, "=", "/etc/vsftpd.conf")
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

	return utils.WriteOptsToFile(sshOpts, " ", "/etc/ssh/sshd_config")
}

func (s *ServiceConfiguration) configureApache() error {
	c, err := s.initService("apache2")
	if err != nil {
		return fmt.Errorf("unable to initialize apache2 service: %s", err.Error())
	}

	if !c {
		return nil
	}

	// Allow Apache through the firewall.
	if err := RunCommandWithArgs("ufw", "allow", "Apache Secure"); err != nil {
		return fmt.Errorf("unable to allow apache2 through firewall: %s", err.Error())
	}

	apacheOpts := map[string]string{}
	if prompts.Confirm("Would you like to set Apache's response header to prod?") {
		apacheOpts["ServerTokens"] = "Prod"
	}

	if prompts.Confirm("Would you like to disable Apache's server signature?") {
		apacheOpts["ServerSignature"] = "Off"
	} else {
		apacheOpts["ServerSignature"] = "On"
	}

	return utils.WriteOptsToFile(apacheOpts, " ", "/etc/apache2/conf-enabled/security.conf")
}
