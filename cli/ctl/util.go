package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func isDirectory(name string) (bool, error) {
	info, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

func pathToFileURL(path string) string {
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return ""
		}
	}
	return fmt.Sprintf("file://%s", filepath.ToSlash(path))
}
