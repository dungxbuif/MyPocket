package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// LoadLocalEnv reads literal KEY=value pairs, never evaluates shell expressions,
// and never overrides an explicitly supplied process environment variable.
func LoadLocalEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return errors.New("cannot read local environment file")
	}
	defer f.Close()
	values := map[string]string{}
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" || strings.ContainsAny(key, " \t\r\n") {
			return errors.New("invalid local environment entry")
		}
		values[key] = strings.TrimSpace(value)
	}
	if scan.Err() != nil {
		return errors.New("cannot read local environment file")
	}
	for key, value := range values {
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return errors.New("invalid local environment key")
			}
		}
	}
	return nil
}
