package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// GetOptsFromFile will return the options specified in the given file.
func GetOptsFromFile(sep string, file string) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %s", file, err.Error())
	}

	buffer, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %s", file, err.Error())
	}

	lines := strings.Split(string(buffer), "\n")
	opts := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}

		line := strings.ReplaceAll(line, "#", "")
		split := strings.Split(line, sep)
		if len(split) < 2 {
			continue
		}

		opts[strings.TrimSpace(split[0])] = strings.Join(split[1:], sep)
	}

	return opts, nil
}

// WriteOptsToFile will write the options specified to the given file.
func WriteOptsToFile(opts map[string]string, sep string, file string) error {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("unable to read %s: %s", file, err.Error())
	}

	lines := strings.Split(string(buffer), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}

		line := strings.ReplaceAll(line, "#", "")
		split := strings.Split(line, sep)
		if len(split) < 2 {
			continue
		}

		if newVal, ok := opts[strings.TrimSpace(split[0])]; ok {
			lines[i] = fmt.Sprintf("%s%s%s", split[0], sep, newVal)
			delete(opts, strings.TrimSpace(split[0]))
		}
	}

	// Add any missing options not set because they do not exist on the config.
	for opt, val := range opts {
		lines = append(lines, fmt.Sprintf("%s%s%s", opt, sep, val))
	}

	// Write to the file.
	if err := os.WriteFile(file, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("unable to write to %s: %s", file, err.Error())
	}

	return nil
}
