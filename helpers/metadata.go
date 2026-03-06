package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/smtdfc/dix/parser"
)

func SaveMetadata(metadata *parser.Metadata, name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	outputPath := filepath.Join(cwd, ".dix", name)

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

func LoadMetadata(path string) (*parser.Metadata, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read metadata file: %w", err)
	}

	metadata := &parser.Metadata{}

	err = json.Unmarshal(bytes, metadata)
	if err != nil {
		return nil, fmt.Errorf("cannot decode json : %w", err)
	}

	return metadata, nil
}
