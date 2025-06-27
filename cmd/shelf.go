package cmd

import (
	"fmt"

	"librarian/librarian"

	"github.com/spf13/cobra"
)

func shelfMedia(cmd *cobra.Command, args []string) {
	dest := args[len(args)-1]
	sources := args[:len(args)-1]

	librarian.MoveFiles(sources, dest)
}

func verifyArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("requires at least two arguments: [source file] and dest")
	}

	return nil
}

// shelf
var shelfCmd = &cobra.Command{
	Use:   "shelf",
	Short: "",
	Long:  ``,
	Args:  verifyArgs,
	Run:   shelfMedia,
}

func init() {
	rootCmd.AddCommand(shelfCmd)
}
