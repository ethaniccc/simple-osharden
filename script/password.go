package script

import (
	"fmt"
	"os"

	"github.com/ethaniccc/simple-osharden/prompts"
	"github.com/ethaniccc/simple-osharden/utils"
)

func init() {
	RegisterScript(&PasswordSetup{})
}

type PasswordSetup struct {
}

func (s *PasswordSetup) Name() string {
	return "pwdsetup"
}

func (s *PasswordSetup) Description() string {
	return "Setup the password policy for the machine."
}

func (s *PasswordSetup) RunOnLinux() error {
	loginDefOpts := map[string]string{}
	pwQualityOpts := map[string]string{}

	loginDefOpts["PASS_MIN_DAYS"] = prompts.RawResponseWithDefaultPrompt("What should the minimum password age be? (recommended is 7)", "7")
	loginDefOpts["PASS_MAX_DAYS"] = prompts.RawResponseWithDefaultPrompt("What should the maximum password age be? (recommended is 90)", "90")
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

	if prompts.Confirm("Should the password dictionary check be enabled?") {
		pwQualityOpts["dictcheck"] = "1"
	} else {
		pwQualityOpts["dictcheck"] = "0"
	}

	if prompts.Confirm("Should the password username check be enabled (check if the username is in the password)?") {
		pwQualityOpts["usercheck"] = "1"
	} else {
		pwQualityOpts["usercheck"] = "0"
	}

	if err := utils.WriteOptsToFile(loginDefOpts, " ", "/etc/login.defs"); err != nil {
		return err
	}

	return utils.WriteOptsToFile(pwQualityOpts, "=", "/etc/security/pwquality.conf")
}

func (s *PasswordSetup) RunOnWindows() error {
	minAge := prompts.RawResponseWithDefaultPrompt("What should the minimum password age be? (recommended is 7)", "7")
	maxAge := prompts.RawResponseWithDefaultPrompt("What should the maximum password age be? (recommended is 30)", "30")
	lockThreshold := prompts.RawResponseWithDefaultPrompt("How many failed login attempts should lock the account? (recommended is 3)", "3")

	// Set the password policy.
	if err := RunCommand(fmt.Sprintf("net accounts /minpwage:%s /maxpwage:%s /lockoutthreshold:%s", minAge, maxAge, lockThreshold)); err != nil {
		return fmt.Errorf("unable to set pw settings: %s", err.Error())
	}

	// Export current security policy to a temporary file
	tmpfile, err := os.CreateTemp("", "secpol")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if err := RunCommand(fmt.Sprintf("secedit /export /cfg %s", tmpfile.Name())); err != nil {
		return err
	}

	pwdOpts := map[string]string{}
	if prompts.Confirm("Enable password complexity checks?") {
		pwdOpts["PasswordComplexity"] = "1"
	} else {
		pwdOpts["PasswordComplexity"] = "0"
	}
	pwdOpts["MinPwdLen"] = prompts.RawResponseWithDefaultPrompt("What should the minimum password length be? (recommended is 8)", "8")

	if err = utils.WriteOptsToFile(pwdOpts, " ", tmpfile.Name()); err != nil {
		return err
	}

	// Import the modified security policy from the temporary file
	return RunCommand("secedit /configure /db secedit.sdb /cfg " + tmpfile.Name() + " /quiet")
}
