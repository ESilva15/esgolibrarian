package main

import (
	"fmt"
	"os"
	"path/filepath"

	"librarian/cmd"
	"librarian/utils"
)

var (
	CfgDir = "/.config/golibrarian/"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home dir: %v\n", err)
		os.Exit(1)
	}
	CfgDir = filepath.Join(homeDir, CfgDir)

	// Check that CfgDir exists
	if exists, _ := utils.PathExists(CfgDir); !exists {
		if err := os.Mkdir(CfgDir, os.ModePerm); err != nil {
			fmt.Printf("Failed to create config dir: %s\n%v\n", CfgDir, err)
			os.Exit(1)
		}
	}

	cmd.Execute()
}
