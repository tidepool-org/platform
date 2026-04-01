package test

import (
	"fmt"
	"os"
)

func LoadFixture(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return string(data), nil
}
