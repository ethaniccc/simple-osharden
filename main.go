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

var list = []prompt.Suggest{
	{Text: "help", Description: "Display a list of commands."},
	{Text: "reboot", Description: "Reboot the machine."},
	{Text: "exit", Description: "Quit Simple-OSHarden."},
}

func main() {
	if runtime.GOOS != "linux" {
		log.Fatal("This program is only supported on Linux.")
	}

	// Run the inital prompt.
	log = logrus.New()

	// Make sure the user running this script is an admin.
	adminCheck()
	// Handle ctrl+c in a goroutine.
	go handleInterrupt()

	log.Info("Updating repositories...")
	script.RunCommand("apt update")
	script.RunCommand("reset")

	log.Info("Loading scripts...")
	for _, s := range script.AvailableScripts() {
		list = append(list, prompt.Suggest{
			Text:        s.Name(),
			Description: s.Description(),
		})
	}

	for {
		script.RunCommand("reset")
		res := mainPrompt()
		if res == "exit" {
			script.RunCommand("reset")
			return
		} else if res == "reboot" {
			script.RunCommand("shutdown -r 0")
			return
		} else if res == "help" {
			for _, s := range script.AvailableScripts() {
				log.Infof("%s - %s\n", s.Name(), s.Description())
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

		script.RunCommand("reset")
		if err := s.Run(); err != nil {
			log.Errorf("Error running script: %s", err.Error())
		} else {
			log.Info("The script ran successfully! Returning to main menu in 3 seconds...")
		}
		<-time.After(time.Second * 3)
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

	script.RunCommand("reset")
	os.Exit(1)
}

func adminCheck() {
	if os.Getuid() != 0 {
		log.Fatal("This program must be run as root.")
	}
}
