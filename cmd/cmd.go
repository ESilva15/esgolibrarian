package cmd

import (
	"os"

	lib "librarian/librarian"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "librarian",
	Short: "CLI tool to help you (actually me I guess) manage your library",
	Long:  `This will help me move and keep the integrity of many files`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("dbstore", "s", false, "Store data in database")
	rootCmd.PersistentFlags().BoolP("cout", "v", false, "Enable console output")
	rootCmd.PersistentFlags().Bool("hwaccel", false, "Enable hardware acceleration")
	rootCmd.PersistentFlags().StringP("format", "f", "table", "Summary format")
	rootCmd.PersistentFlags().StringP("dryrun", "d", "table", "Print actions that would be taken")
}

func loadOptions(cmd *cobra.Command) *lib.LibOptions {
	libOptions := lib.NewLibOptions()

	libOptions.ConsoleOutput, _ = cmd.Flags().GetBool("cout")
	libOptions.UseHWAccel, _ = cmd.Flags().GetBool("hwaccel")
	libOptions.Format, _ = cmd.Flags().GetString("format")
	libOptions.DryRun, _ = cmd.Flags().GetBool("dryrun")
	libOptions.DBStore, _ = cmd.Flags().GetBool("dbstore")

	return libOptions
}
