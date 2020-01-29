// +build windows

package main

import (
	"os"
	"testing"
)

func TestConfigPath(t *testing.T) {
	want, got := os.Getenv("HOME")+"AppData\\timer\\sounds", getSoundsDir()
	if want != got {
		t.Errorf("want %s got %s", want, got)
	}
}
