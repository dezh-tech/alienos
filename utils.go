package main

import (
	"fmt"
	"os"
)

func Mkdir(path string) error {
	// create the directory
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("could not create directory %s", path)
	}

	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}
