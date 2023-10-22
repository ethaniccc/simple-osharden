package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/sirupsen/logrus"
)

const (
	OSMac     = "darwin"
	OSWindows = "windows"
	OSLinux   = "linux"
)

var mediaExtensions = []string{
	"mp3", "mp4", "txt", "png",
	"jpg", "jpeg", "gif", "wav",
	"mov", "avi", "wmv", "flv",
	"m4a", "m4v", "webm", "mkv",
	"doc", "docx", "xls", "xlsx",
	"ppt", "pptx", "pdf",
}

var unauthorizedPrograms = []string{
	"wireshark", "ophcrack", "nmap",
}

var logger *logrus.Logger

func main() {
	logger = logrus.New()

	switch runtime.GOOS {
	case OSWindows:
		f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		if err != nil {
			logger.Warn("Please run this program as administrator.")
			return
		}

		f.Close()
	default:
		if os.Geteuid() != 0 {
			logger.Warn("Please run this program as root.")
			return
		}
	}

	var homeDirectory string
	switch runtime.GOOS {
	case OSWindows:
		homeDirectory = "C:\\Users\\"
	case OSLinux:
		homeDirectory = "/home/"
	case OSMac:
		homeDirectory = "/Users/"
	default:
		logger.Fatalf("Unsupported OS: %s", runtime.GOOS)
		return
	}

	customDirectory := os.Getenv("TEST_DIR")
	if customDirectory != "" {
		homeDirectory = customDirectory
	}

	debug.SetMemoryLimit(256 * 1024 * 1024)
	debug.SetGCPercent(-1)

	os.Remove("results.txt")
	f, err := os.OpenFile("results.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	logger.Info("Searching for media files...")
	if err != nil {
		logger.Error("Unable to open results file, output will be printed in the terminal.")
	} else {
		logger.SetOutput(f)
	}
	t := time.Now()
	found := findMediaFiles(homeDirectory)
	logger.SetOutput(os.Stdout) // Reset the output to the terminal.
	logger.Infof("Found %d media files in %v milliseconds", found, time.Since(t).Milliseconds())

	logger.Info("Disabling unnecessary services...")
	disableServices()
	logger.Info("Done! Some services may need to be disabled manually.")

	logger.Info("Attempting to remove unauthorized programs...")
	removeUnauthorizedPrograms()
	logger.Info("Done! Some programs may need to be removed manually.")

	logger.Info("Attempting to update programs...")
	updatePrograms()
	logger.Info("Done! Some programs may need to be updated manually.")

	logger.Info("Attempting to enable firewall...")
	enableFirewall()
	logger.Info("Done! Some firewall rules may need to be added manually.")

	logger.Info("Starting validation for users on this machine...")
	verifyCurrentUsers()
	logger.Info("Done!")

	logger.Info("Entering user into file editor to disable SSH root access...")
	disableRootSSH()
	logger.Info("Done!")

	logger.Info("Setting password policy...")
	setPasswordPolicy()
	logger.Info("Done!")

	logger.Info("Script completed. Some tasks may need to be completed manually. This script will terminate in 10 seconds.")
	<-time.After(time.Second * 10)

	// Reset the terminal due to a bug in the prompt library.
	createCommand("reset").Run()
}

// findMediaFiles recursively searches for files with media extensions in the given directory.
func findMediaFiles(path string) (found int) {
	files, err := os.ReadDir(path)
	if err != nil {
		logger.Errorf("Unable to read files from %s: %s", path, err.Error())
		return -1
	}

	for _, file := range files {
		if !file.IsDir() {
			split := strings.Split(file.Name(), ".")
			if len(split) < 2 {
				continue
			}

			extension := split[len(split)-1]
			for _, badExtension := range mediaExtensions {
				if extension != badExtension {
					continue
				}

				logger.Warnf("Found %s", path+file.Name())
				found++
			}

			continue
		}

		found += findMediaFiles(path + file.Name() + "/")
	}

	return found
}

// removeUnauthorizedPrograms removes unauthorized programs on the system. This is only supported on Linux.
func removeUnauthorizedPrograms() {
	if runtime.GOOS != OSLinux {
		logger.Error("This functionally is not supported on this OS. Please remove unauthorized programs manually.")
		return
	}

	for _, program := range unauthorizedPrograms {
		logger.Infof("Attempting to remove %s...", program)
		createCommand("apt", "remove", program).Run()
	}

	createCommand("apt", "autoremove").Run()
}

// enableFirewall enables the firewall, based on the given OS.
func enableFirewall() {
	switch runtime.GOOS {
	case OSLinux, OSMac:
		// Install/update uncomplicated firewall.
		createCommand("apt", "install", "ufw").Run()

		// Enable the firewall.
		createCommand("ufw", "enable").Run()
		createCommand("ufw", "default", "deny", "incoming").Run()
		createCommand("ufw", "default", "allow", "outgoing").Run()
	case OSWindows:
		// Enable the firewall.
		createCommand("netsh", "advfirewall", "set", "allprofiles", "state", "on").Run()
		createCommand("netsh", "advfirewall", "set", "allprofiles", "firewallpolicy", "blockinbound,allowoutbound").Run()
	}
}

// disableServices disables services that may increase the attack surface of the system.
func disableServices() {
	switch runtime.GOOS {
	case OSLinux:
		// Disable the FTP server.
		createCommand("systemctl", "disable", "vsftpd").Run()
		if verifyUninstallPrompt("vsftpd") {
			createCommand("apt", "remove", "vsftpd").Run()
		}

		// Disable the Telnet server.
		createCommand("systemctl", "disable", "telnetd").Run()
		if verifyUninstallPrompt("telnetd") {
			createCommand("apt", "remove", "telnetd").Run()
		}

		// Disable the nginx server.
		createCommand("systemctl", "disable", "nginx").Run()
		if verifyUninstallPrompt("nginx") {
			createCommand("apt", "remove", "nginx").Run()
		}

		// Disable the apache server
		createCommand("systemctl", "disable", "apache2").Run()
		if verifyUninstallPrompt("apache2") {
			createCommand("apt", "remove", "apache2").Run()
		}
	default:
		logger.Error("This functionally is not yet supported on this OS. Please disable uneeded services manually.")
	}
}

// updatePrograms updates necessary programs based on the given OS.
func updatePrograms() {
	switch runtime.GOOS {
	case OSLinux, OSMac:
		logger.Info("Update repository list...")
		createCommand("apt", "update").Run()

		// Update all packages on the system (includes systemd).
		logger.Info("Updating all packages in 5 seconds: please note that this may take a while.")
		<-time.After(time.Second * 5)
		createCommand("apt", "upgrade").Run()
	default:
		logger.Error("This functionally is not yet supported on this OS. Please update any programs manually.")
	}
}

// verifyCurrentUsers verifies that the current users are authorized. This will require manual input
// from the user to verify that they are authorized. This will also remove the administator permissions
// of the account if they are not considered admins.
func verifyCurrentUsers() {
	switch runtime.GOOS {
	case OSLinux:
		users, err := os.ReadDir("/home/")
		if err != nil {
			logger.Errorf("Unable to read users from /home/: %s", err.Error())
			return
		}

		var sudoersList map[string]bool
		res, err := exec.Command("getent", "group", "sudo").Output()
		if err != nil {
			logger.Errorf("Unable to get sudoers list: %s", err.Error())
		} else {
			split := strings.Split(string(res), ":")
			if len(split) < 4 {
				logger.Errorf("Unable to parse sudoers list (got %s)", string(res))
			} else {
				sudoersList = make(map[string]bool)
				split = strings.Split(split[3:][0], ",")
				for _, user := range split {
					user = strings.TrimSpace(user)
					sudoersList[user] = true
					logger.Infof("Found administator %s", user)
				}
			}
		}

		// We assume all users have a home directory in the /home/ directory.
		for _, user := range users {
			if !user.IsDir() || user.Name() == "." || user.Name() == ".." {
				continue
			}

			// Have the user verify that the user given is authorized on the system.
			if !verifyAuthorizedUserPrompt(user.Name()) {
				logger.Warnf("Confirmed %s is not an authorized user of the system. User will be deleted shortly.", user.Name())
				createCommand("deluser", "--remove-home", user.Name()).Run()
				logger.Infof("Successfully deleted user %s", user.Name())
				continue
			}
			logger.Infof("Confirmed %s as an authorized user of the system.", user.Name())

			// Check if the user is in the sudoers list. If they are not, we can skip the admin check.
			if _, ok := sudoersList[user.Name()]; !ok {
				continue
			}

			// Have the user verify that the user in question is an admin on the system.
			if !verifyAdminUserPrompt(user.Name()) {
				logger.Warnf("Confirmed %s is not an administator of the system. User will be removed from sudoers list shortly.", user.Name())
				createCommand("deluser", user.Name(), "sudo").Run()
				logger.Infof("Successfully removed user %s from sudoers list", user.Name())
				continue
			}

			logger.Infof("Confirmed %s as an administator of the system.", user.Name())
		}
	case OSWindows:
		var adminList map[string]bool

		res, err := exec.Command("net", "localgroup", "Administrators").Output()
		if err != nil {
			logger.Errorf("Unable to get administrators list: %s", err.Error())
		} else {
			adminList = make(map[string]bool)
			// This is... very ugly. But it works. If it doesn't and it crashes because we assumed wrong - oh well!
			actualList := strings.Replace(strings.Split(string(res), "-------------------------------------------------------------------------------")[1], "The command completed successfully.", "", 1)
			split := strings.Split(actualList, "\n")[1:]

			for _, admin := range split {
				admin = strings.TrimSpace(admin)
				if admin == "Administrator" || admin == "" {
					continue
				}

				adminList[admin] = true
				logger.Infof("Found administator %s", admin)
			}
		}

		users, err := os.ReadDir("C:\\Users\\")
		if err != nil {
			logger.Errorf("Unable to read users from C:\\Users\\: %s", err.Error())
			return
		}

		for _, user := range users {
			if !user.IsDir() || user.Name() == "." || user.Name() == ".." {
				continue
			}

			// Have the user verify that the user given is authorized on the system.
			if !verifyAuthorizedUserPrompt(user.Name()) {
				logger.Warnf("Confirmed %s is not an authorized user of the system. User will be deleted shortly.", user.Name())
				cmd := createCommand("net", "user", user.Name(), "/delete")
				if err := cmd.Run(); err != nil {
					logger.Errorf("Unable to delete user %s: %s", user.Name(), err.Error())
					continue
				}

				logger.Infof("Successfully deleted user %s", user.Name())
				continue
			}
			logger.Infof("Confirmed %s as an authorized user of the system.", user.Name())

			if _, ok := adminList[user.Name()]; !ok {
				continue
			}

			// Have the user verify that the user in question is an admin on the system.
			if !verifyAdminUserPrompt(user.Name()) {
				logger.Warnf("Confirmed %s is not an administator of the system. User will be removed from administrators list shortly.", user.Name())
				cmd := createCommand("net", "localgroup", "Administrators", user.Name(), "/delete")
				if err := cmd.Run(); err != nil {
					logger.Errorf("Unable to remove user %s from administrators list: %s", user.Name(), err.Error())
					continue
				}

				logger.Infof("Successfully removed user %s from administrators list", user.Name())
				continue
			}

			logger.Infof("Confirmed %s as an administator of the system.", user.Name())
		}
	default:
		logger.Error("This functionally is not yet supported on this OS. Please verify users manually.")
	}
}

// disableRootSSH opens the sshd_config file in nano for the user to edit.
func disableRootSSH() {
	if runtime.GOOS != OSLinux {
		logger.Error("This functionally is not supported on this OS. Please disable root SSH access manually.")
		return
	}

	createCommand("nano", "/etc/ssh/sshd_config").Run()
}

// setPasswordPolicy opens the login.defs file in nano for the user to edit.
func setPasswordPolicy() {
	switch runtime.GOOS {
	case OSLinux:
		logger.Info("Installing password policy library...")
		createCommand("apt", "install", "libpam-cracklib").Run()

		// Have the user set the password policy manually.
		createCommand("nano", "/etc/login.defs").Run()
		createCommand("nano", "/etc/pam.d/common-password").Run()
	case OSWindows:
		logger.Info("Setting mininum password length to 8 characters...")
		createCommand("net", "accounts", "/minpwlen:8").Run()
		logger.Info("Setting maximum password age to 90 days...")
		createCommand("net", "accounts", "/maxpwage:90").Run()
		logger.Info("Setting minimum password age to 10 day...")
		createCommand("net", "accounts", "/minpwage:10").Run()
		logger.Info("Setting password history to 5 passwords...")
		createCommand("net", "accounts", "/uniquepw:5").Run()
		logger.Info("Setting lockout threshold to 5 attempts...")
		createCommand("net", "accounts", "/lockoutthreshold:5").Run()
		logger.Info("Setting lockout duration to 30 minutes...")
		createCommand("net", "accounts", "/lockoutduration:30").Run()
	default:
		logger.Error("This functionally is not supported on this OS. Please set the password policy manually.")
	}
}

// createCommand creates a command and sets the IO to the standard IO.
func createCommand(c string, args ...string) *exec.Cmd {
	cmd := exec.Command(c, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// verifyUninstallPrompt creates a verification prompt for the user to verify that they want to uninstall a program.
func verifyUninstallPrompt(program string) bool {
	logger.Warn("Please respond to this prompt with caution - as programs (when you answer \"n\" or \"no\") will be uninstalled.")
	res := prompt.Input(fmt.Sprintf("Do you want to uninstall %s? (y/n): ", program), dummyPromptCompletor, dummyPromptOption)
	switch strings.ToLower(res) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		logger.Warn("Invalid input, please try again.")
		return verifyUninstallPrompt(program)
	}
}

// verifyAuthorizedUserPrompt creates a verification prompt for the user to verify that another user is authorized.
func verifyAuthorizedUserPrompt(user string) bool {
	logger.Warn("Please respond to this prompt with caution - as unauthorized users (when you answer \"n\" or \"no\") will be deleted.")
	res := prompt.Input(fmt.Sprintf("Is %s allowed to be on this machine? (y/n): ", user), dummyPromptCompletor, dummyPromptOption)
	switch strings.ToLower(res) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		logger.Warn("Invalid input, please try again.")
		return verifyAuthorizedUserPrompt(user)
	}
}

// verifyAdminUserPrompt creates a verification prompt for the user to verify that they are an admin.
func verifyAdminUserPrompt(user string) bool {
	logger.Warn("Please respond to this prompt with caution - as non-admins (when you answer \"n\" or \"no\") lose their administrator permissions.")
	res := prompt.Input(fmt.Sprintf("Is %s supposed to have admin permissions? (y/n): ", user), dummyPromptCompletor, dummyPromptOption)
	switch strings.ToLower(res) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		logger.Warn("Invalid input, please try again.")
		return verifyAdminUserPrompt(user)
	}
}

var dummyPromptOption prompt.Option = func(p *prompt.Prompt) error {
	return nil
}

var dummyPromptCompletor prompt.Completer = func(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}
