package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number and build information",
	Long:  `Print the version number and build information`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-terminology v%s (%s)\n", Version, Build)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// override RootCmd version
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// override RootCmd version
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
