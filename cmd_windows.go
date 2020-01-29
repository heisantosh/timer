package main

import (
	"os"
	"path/filepath"
)

// getSoundsDir returns the directory storing added sounds.
func getSoundsDir() string {
	return filepath.Join(os.Getenv("HOME"), "AppData", "timer", "sounds")
}
