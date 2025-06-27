package cmd

import (
	"fmt"

	"librarian/librarian"

	"github.com/spf13/cobra"
)

func validateMedia(cmd *cobra.Command, args []string) {
	librarian.ValidateFiles(args)
}

// shelf
var validateMediaCmd = &cobra.Command{
	Use:   "validate",
	Short: "",
	Long:  ``,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least one argument: [source file]")
		}

		return nil
	},
	Run: validateMedia,
}

func init() {
	rootCmd.AddCommand(validateMediaCmd)
}
