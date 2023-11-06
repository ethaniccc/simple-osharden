package script

import (
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
}

type Script interface {
	// Name returns the name of the script.
	Name() string
	// Description returns the description of the script.
	Description() string
}

// LinuxSupportedScript is an interface for a script that supports Linux.
type LinuxSupportedScript interface {
	// RunOnLinux executes the script.
	RunOnLinux() error
}

// WindowsSupportedScript is an interface for a script that supports Windows.
type WindowsSupportedScript interface {
	// RunOnWindows executes the script on Windows.
	RunOnWindows() error
}

var scriptPool = map[string]Script{}

// AvailableScripts returns all the scripts in the script pool.
func AvailableScripts() map[string]Script {
	return scriptPool
}

// AvailableLinuxScripts returns all the scripts in the script pool that support Windows.
func AvailableLinuxScripts() map[string]Script {
	scripts := map[string]Script{}
	for name, s := range scriptPool {
		if _, ok := s.(LinuxSupportedScript); ok {
			scripts[name] = s
		}
	}

	return scripts
}

// AvailableWindowsScripts returns all the scripts in the script pool that support Windows.
func AvailableWindowsScripts() map[string]Script {
	scripts := map[string]Script{}
	for name, s := range scriptPool {
		if _, ok := s.(WindowsSupportedScript); ok {
			scripts[name] = s
		}
	}

	return scripts
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

// RunScript runs a script.
func RunScript(s Script) error {
	var runFunc func() error
	switch runtime.GOOS {
	case "windows":
		ws, ok := s.(WindowsSupportedScript)
		if !ok {
			return fmt.Errorf("not supported on windows")
		}

		runFunc = ws.RunOnWindows
	case "linux":
		ls, ok := s.(LinuxSupportedScript)
		if !ok {
			return fmt.Errorf("not supported on linux")
		}

		runFunc = ls.RunOnLinux
	}

	return runFunc()
}
