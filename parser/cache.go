package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const cacheSchemaVersion = 1

type cacheFile struct {
	SchemaVersion int       `json:"schema_version"`
	Fingerprint   string    `json:"fingerprint"`
	Metadata      *Metadata `json:"metadata"`
}

func cacheKey(dirs []string, options ScanOptions) (string, error) {
	h := sha256.New()
	h.Write([]byte("dix-cache-schema:" + strconv.Itoa(cacheSchemaVersion)))
	h.Write([]byte("\ngo:" + runtime.Version()))
	h.Write([]byte("\nworkspace:" + boolString(options.Workspace)))

	files := make([]string, 0)
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == ".dix" || entry.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(entry.Name(), ".go") || entry.Name() == "go.mod" || entry.Name() == "go.sum" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	if options.Workspace {
		if workFile, err := findGoWork(dirs[0]); err == nil {
			files = append(files, workFile)
		}
	}
	sort.Strings(files)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		h.Write([]byte("\nfile:" + path + "\n"))
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func loadCache(path, fingerprint string) (*Metadata, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var entry cacheFile
	if json.Unmarshal(data, &entry) != nil || entry.SchemaVersion != cacheSchemaVersion || entry.Fingerprint != fingerprint || entry.Metadata == nil {
		return nil, false
	}
	return entry.Metadata, true
}

func saveCache(path, fingerprint string, metadata *Metadata) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	lockPath := path + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return
	}
	_ = lock.Close()
	defer os.Remove(lockPath)

	data, err := json.MarshalIndent(cacheFile{SchemaVersion: cacheSchemaVersion, Fingerprint: fingerprint, Metadata: metadata}, "", "  ")
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cache-*.tmp")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return
	}
	if err = tmp.Close(); err != nil {
		return
	}
	_ = os.Rename(tmpName, path)
}
