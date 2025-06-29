package main

import (
	"fmt"
	"os"
	"path/filepath"

	"librarian/cmd"
	"librarian/logger"
	"librarian/utils"
)

var (
	CfgDir = "/.config/golibrarian/"
)

func main() {
	// Find the HOMEDIR of the user and create the config directory there
	// will fail if the user doesn't have a .config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home dir: %v\n", err)
		os.Exit(1)
	}
	CfgDir = filepath.Join(homeDir, CfgDir)

	// Check that CfgDir exists, if it doesn't: create it
	if exists, _ := utils.PathExists(CfgDir); !exists {
		if err := os.Mkdir(CfgDir, os.ModePerm); err != nil {
			fmt.Printf("Failed to create config dir: %s\n%v\n", CfgDir, err)
			os.Exit(1)
		}
	}

	// Set the path for the log file for now
	logger.SetLogpath(CfgDir)

	cmd.Execute()
}
