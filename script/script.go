package script

import "github.com/sirupsen/logrus"

var logger *logrus.Logger

func init() {
	logger = logrus.New()

	RegisterScript(&NetworkSetup{})
	RegisterScript(&UpdateDNS{})
	RegisterScript(&NetApps{})

	RegisterScript(&SystemConfig{})
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
func RegisterScript(s Script) {
	scriptPool[s.Name()] = s
}

// UnregisterScript unregisters a script from the script pool.
func UnregisterScript(name string) {
	delete(scriptPool, name)
}
