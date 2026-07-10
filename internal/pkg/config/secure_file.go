package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func readRegularFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("refusing non-regular file %q", path)
	}
	return os.ReadFile(path)
}

func writeFileAtomic(path string, data []byte) (retErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".monogit-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		if retErr != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(0600); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
