package prompts

import "github.com/c-bata/go-prompt"

var DummyPromptOption prompt.Option = func(p *prompt.Prompt) error {
	return nil
}

var DummyPromptCompletor prompt.Completer = func(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

// YesNoPrompt
func YesNoPrompt(msg string) bool {
	switch prompt.Input(msg+" (y/n) >> ", DummyPromptCompletor, DummyPromptOption) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return YesNoPrompt(msg)
	}
}
