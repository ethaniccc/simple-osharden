package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// CreateTempFileFrom will create a temp file from the given file.
func CreateTempFileFrom(f string) (*os.File, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %s", f, err.Error())
	}
	defer file.Close()

	tmpFile, err := os.CreateTemp("", "osharden-temp")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %s", err.Error())
	}

	buffer, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %s", f, err.Error())
	}

	if _, err := tmpFile.Write(buffer); err != nil {
		return nil, fmt.Errorf("unable to write to temp file: %s", err.Error())
	}

	return tmpFile, nil
}

// WriteOptsToFile will write the options specified to the given file.
func WriteOptsToFile(opts map[string]string, sep string, file string) error {
	// Create a temp file, so in the case that the program crashes, we don't lose the original file.
	f, err := CreateTempFileFrom(file)
	if err != nil {
		return fmt.Errorf("unable to create temp file: %s", err.Error())
	}
	defer f.Close()
	defer os.Remove(f.Name())

	buffer, err := os.ReadFile(f.Name())
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
