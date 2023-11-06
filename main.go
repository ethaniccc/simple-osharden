package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/ethaniccc/simple-osharden/prompts"
	"github.com/ethaniccc/simple-osharden/script"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

var list []prompt.Suggest

func main() {
	log = logrus.New()

	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		log.Fatal("Simple-OSHarden is currently only supported on Windows and Linux.")
		return
	}

	// Make sure the user running this script is an admin.
	adminCheck()
	// Handle ctrl+c in a goroutine.
	go handleInterrupt()

	log.Info("Updating repositories...")
	script.RunCommand("apt update")
	script.ResetTerminal()

	// Load all the available scripts into the prompt suggestion list.
	log.Info("Loading scripts...")
	list = []prompt.Suggest{
		{Text: "help", Description: "Display a list of commands."},
		{Text: "reboot", Description: "Reboot the machine."},
		{Text: "exit", Description: "Quit Simple-OSHarden."},
	}

	scripts := map[string]script.Script{}
	if runtime.GOOS == "windows" {
		scripts = script.AvailableWindowsScripts()
	} else if runtime.GOOS == "linux" {
		scripts = script.AvailableLinuxScripts()
	}

	for _, s := range scripts {
		list = append(list, prompt.Suggest{Text: s.Name(), Description: s.Description()})
	}

	// Run the main prompt.
	for {
		script.ResetTerminal()
		res := mainPrompt()

		// Hardcode commands because I am very very very very very lazy :)
		if res == "exit" {
			script.ResetTerminal()
			return
		} else if res == "reboot" {
			if runtime.GOOS == "windows" {
				script.RunCommand("shutdown /r /t 0")
			} else if runtime.GOOS == "linux" {
				script.RunCommand("shutdown -r 0")
			}

			return
		} else if res == "help" {
			scripts := map[string]script.Script{}
			if runtime.GOOS == "windows" {
				scripts = script.AvailableWindowsScripts()
			} else if runtime.GOOS == "linux" {
				scripts = script.AvailableLinuxScripts()
			}

			if len(scripts) == 0 {
				log.Error("At the moment, no scripts are supported on your operating system.")
			} else {
				for _, s := range scripts {
					log.Infof("%s - %s\n", s.Name(), s.Description())
				}
			}

			prompts.RawResponsePrompt("Press enter to continue")
			continue
		}

		s := script.GetScript(res)
		if s == nil {
			log.Errorf("\"%s\" is not a valid script.", res)
			<-time.After(time.Second * 3)
			continue
		}

		// Run the script
		if err := script.RunScript(s); err != nil {
			prompts.RawResponsePrompt(fmt.Sprintf("The script failed to run due to an error: %s\n[Press enter to continue]", err.Error()))
		} else {
			prompts.RawResponsePrompt("Script finished running successfully\n[Press enter to continue]")
		}

		runtime.GC()
	}
}

func mainPrompt() string {
	msg := `          
 _____ _____ _____           _         
|     |   __|  |  |___ ___ _| |___ ___ 
|  |  |__   |     | .'|  _| . | -_|   |
|_____|_____|__|__|__,|_| |___|___|_|_|			

- @ethaniccc						


Simple-OSHarden is a tool that can be used to harden your machine. 
As of Nov. 4, 2023, this tool is in beta, and only supports linux.

If for any reason, you'd like to contact me, please send an email
to benjaminscyber@skiff.com. I'll try to respond as soon as I can.

Github: https://www.github.com/ethaniccc/
Source code: https://github.com/ethaniccc/simple-osharden
`

	fmt.Println(msg)
	return prompt.Input("Enter a command >> ", func(d prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(list, d.GetWordBeforeCursor(), true)
	}, prompt.OptionMaxSuggestion(16))
}

func handleInterrupt() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	<-sigchan

	script.ResetTerminal()
	os.Exit(1)
}

func adminCheck() {
	switch runtime.GOOS {
	case "windows":
		f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		if err != nil {
			log.Fatal("You must run this script as an administrator.")
			return
		}
		f.Close()
	case "linux":
		if os.Geteuid() != 0 {
			log.Fatal("You must run this script as an administrator.")
		}
	}
}
