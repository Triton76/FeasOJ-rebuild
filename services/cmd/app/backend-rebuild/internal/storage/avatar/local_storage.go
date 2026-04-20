package avatarstorage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	dir string
}

func NewLocalStorage(dir string) (*LocalStorage, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("avatar storage dir is empty")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{dir: dir}, nil
}

func (s *LocalStorage) Save(ctx context.Context, name string, data []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		return "", fmt.Errorf("invalid avatar file name")
	}

	tmp, err := os.CreateTemp(s.dir, base+".tmp-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	finalPath := filepath.Join(s.dir, base)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return base, nil
}
