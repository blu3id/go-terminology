package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
	"github.com/wardle/go-terminology/terminology"
)

var sct *terminology.Svc
var profilecpu string
var Version, Build string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-terminology",
	Short: "A SNOMED-CT terminology server and command line tool",
	Long:  `go-terminology is a command-line SNOMED-CT terminology tool and server.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if profilecpu != "" {
			f, err := os.Create(profilecpu)
			if err != nil {
				log.Fatal(err)
			}
			pprof.StartCPUProfile(f)
		}

		if len(args) < 1 {
			return fmt.Errorf("must specify path to datastore")
		}

		readOnly := true
		// Set readOnly to false if command in following map
		readWriteCommands := map[string]bool{"import": true, "precompute": true, "reset": true}
		if _, ok := readWriteCommands[cmd.CalledAs()]; ok {
			readOnly = false
		}

		// Create new terminology service
		sct, err = terminology.New(args[0], readOnly)
		if err != nil {
			return fmt.Errorf("couldn't open datastore: %v", err)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		err := sct.Close()
		if profilecpu != "" {
			pprof.StopCPUProfile()
		}
		return err
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Receives version and build strings from main
func Execute(version string, build string) {
	Version = version
	Build = build
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profilecpu, "profile-cpu", "", "write cpu profile to `file` specified")
}
