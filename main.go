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

	// Handle ctrl+c in a goroutine.
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)

		<-sigchan

		script.CreateCommand("reset").Run()
		os.Exit(1)
	}()

	for {
		res := mainPrompt()
		s := script.GetScript(res)
		if s == nil {
			log.Errorf("\"Z%s\" is not a valid script.", res)
			continue
		}

		if err := s.Run(); err != nil {
			log.Error(err)
		} else {
			log.Info("The script ran successfully! Returning to main menu in 3 seconds...")
		}

		<-time.After(time.Second * 3)
		script.CreateCommand("reset").Run()
	}
}

func mainPrompt() string {
	return prompt.Input("Chose a command >> ", func(d prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix([]prompt.Suggest{
			{Text: "firewall", Description: "Installs and configures the UFW firewall"},
		}, d.GetWordBeforeCursor(), true)
	}, prompts.DummyPromptOption)
}
