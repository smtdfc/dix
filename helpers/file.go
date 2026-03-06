package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteTextFile(text string, filePath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	bytes := []byte(text)

	outputPath := filepath.Join(cwd, filePath)

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(outputPath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
