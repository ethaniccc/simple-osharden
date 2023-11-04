package script

import "github.com/sirupsen/logrus"

var logger *logrus.Logger

func init() {
	logger = logrus.New()

	RegisterScript("firewall", &Firewall{})
}

type Script interface {
	// Name returns the name of the script.
	Name() string
	// Description returns the description of the script.
	Description() string

	// Run executes the script.
	Run() error
}

var scriptPool = map[string]Script{}

// AvailableScripts returns all the scripts in the script pool.
func AvailableScripts() map[string]Script {
	return scriptPool
}

// GetScript returns a script from the script pool.
func GetScript(name string) Script {
	s, ok := scriptPool[name]
	if !ok {
		return nil
	}

	return s
}

// RegisterScript registers a script to the script pool.
func RegisterScript(name string, s Script) {
	scriptPool[name] = s
}
