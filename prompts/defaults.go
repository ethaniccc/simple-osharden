package prompts

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

var DummyPromptOption prompt.Option = func(p *prompt.Prompt) error {
	return nil
}

var DummyPromptCompletor prompt.Completer = func(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

// RawResponsePrompt prompts the user and returns their response w/o any processing.
func RawResponsePrompt(msg string) string {
	return prompt.Input(msg+" >> ", DummyPromptCompletor, DummyPromptOption)
}

// Confirm returns true if the user selected yes, and false if the user selected no.
func Confirm(msg string) bool {
	switch strings.ToLower(RawResponsePrompt(msg + " (y/n)")) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return Confirm(msg)
	}
}
