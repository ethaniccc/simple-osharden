package script

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethaniccc/simple-osharden/prompts"
)

func init() {
	RegisterScript(&PasswordSetup{})
}

type PasswordSetup struct {
}

func (s *PasswordSetup) Name() string {
	return "pwd-setup"
}

func (s *PasswordSetup) Description() string {
	return "Setup the password policy for the machine."
}

func (s *PasswordSetup) Run() error {
	loginDefOpts := map[string]string{}
	pwQualityOpts := map[string]string{}

	loginDefOpts["PASS_MIN_DAYS"] = prompts.RawResponseWithDefaultPrompt("What should the minimum password age be? (recommended is 7)", "7")
	loginDefOpts["PASS_MAX_DAYS"] = prompts.RawResponseWithDefaultPrompt("What should the maximum password age be? (recommended is 30)", "30")
	loginDefOpts["ENCRYPT_METHOD"] = prompts.RawResponseWithDefaultPrompt("What should the encryption method be? (recommended is SHA512)", "SHA512")
	loginDefOpts["LOGIN_RETRIES"] = prompts.RawResponseWithDefaultPrompt("How many login retries should be allowed? (recommended is 3)", "3")

	pwQualityOpts["minlen"] = prompts.RawResponseWithDefaultPrompt("What should the minimum password length be? (recommended is 8)", "8")
	if prompts.Confirm("Should password complexity checks be enabled?") {
		pwQualityOpts["dcredit"] = "-1"
		pwQualityOpts["ucredit"] = "-1"
		pwQualityOpts["ocredit"] = "-1"
		pwQualityOpts["lcredit"] = "-1"
	} else {
		pwQualityOpts["dcredit"] = "0"
		pwQualityOpts["ucredit"] = "0"
		pwQualityOpts["ocredit"] = "0"
		pwQualityOpts["lcredit"] = "0"
	}

	if prompts.Confirm("Should the password disctionary check be enabled?") {
		pwQualityOpts["dictcheck"] = "1"
	} else {
		pwQualityOpts["dictcheck"] = "0"
	}

	if prompts.Confirm("Should the password username check be enabled (check if the username is in the password)?") {
		pwQualityOpts["usercheck"] = "1"
	} else {
		pwQualityOpts["usercheck"] = "0"
	}

	// Update /etc/login.defs
	buffer, err := os.ReadFile("/etc/login.defs")
	if err != nil {
		return fmt.Errorf("unable to read /etc/login.defs: %s", err.Error())
	}

	// Go through each line on the file and search for options that we want to change.
	lines := strings.Split(string(buffer), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		line := strings.ReplaceAll(line, "#", "")
		split := strings.Split(line, " ")
		if len(split) < 2 {
			continue
		}

		if newVal, ok := loginDefOpts[split[0]]; ok {
			lines[i] = fmt.Sprintf("%s %s", split[0], newVal)
			delete(loginDefOpts, split[0])
		}
	}

	// Add any missing options not set because they do not exist on the config.
	for opt, val := range loginDefOpts {
		lines = append(lines, fmt.Sprintf("%s %s", opt, val))
	}

	// Write to /etc/login.defs
	if err := os.WriteFile("/etc/login.defs", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write to /etc/login.defs: %s", err.Error())
	}
	logger.Info("Updated /etc/login.defs")

	buffer, err = os.ReadFile("/etc/security/pwquality.conf")
	if err != nil {
		return fmt.Errorf("unable to read /etc/security/pwquality.conf: %s", err.Error())
	}

	// Go through each line on the file and search for options that we want to change.
	lines = strings.Split(string(buffer), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		line := strings.ReplaceAll(line, "#", "")
		split := strings.Split(line, "=")
		if len(split) < 2 {
			continue
		}

		if newVal, ok := pwQualityOpts[strings.TrimSpace(split[0])]; ok {
			lines[i] = fmt.Sprintf("%s = %s", split[0], newVal)
			delete(pwQualityOpts, strings.TrimSpace(split[0]))
		}
	}

	// Add any missing options not set because they do not exist on the config.
	for opt, val := range pwQualityOpts {
		lines = append(lines, fmt.Sprintf("%s = %s", opt, val))
	}

	// Write to /etc/security/pwquality.conf
	if err := os.WriteFile("/etc/security/pwquality.conf", []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write to /etc/security/pwquality.conf: %s", err.Error())
	}
	logger.Info("Updated /etc/security/pwquality.conf")

	return nil
}
