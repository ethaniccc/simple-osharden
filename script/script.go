package script

import "github.com/sirupsen/logrus"

var logger *logrus.Logger

type Script interface {
	// Run executes the script.
	Run() error
}

var scriptPool = map[string]Script{}

// GetScript returns a script from the script pool.
func GetScript(name string) Script {
	s, ok := scriptPool[name]
	if !ok {
		return nil
	}

	return s
}

// RegisterScript registers a script to the script pool.
func init() {
	logger = logrus.New()
	scriptPool["firewall"] = &Firewall{}
}
