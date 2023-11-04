package main

import (
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

	for {
		script.RunCommand("reset")
		res := mainPrompt()
		if res == "exit" {
			script.RunCommand("reset")
			return
		} else if res == "reboot" {
			script.RunCommand("shutdown -r 0")
			return
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
	}
}

func mainPrompt() string {
	return prompt.Input("Enter a command: ", func(d prompt.Document) []prompt.Suggest {
		list := []prompt.Suggest{
			{Text: "exit", Description: "Quit Simple-OSHarden."},
			{Text: "reboot", Description: "Reboot the machine."},
		}

		for _, s := range script.AvailableScripts() {
			list = append(list, prompt.Suggest{
				Text:        s.Name(),
				Description: s.Description(),
			})
		}

		return prompt.FilterHasPrefix(list, d.GetWordBeforeCursor(), true)
	}, prompts.DummyPromptOption)
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
