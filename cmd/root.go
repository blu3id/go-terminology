package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wardle/go-terminology/medicine"
	"github.com/wardle/go-terminology/terminology"
)

var sct *terminology.Svc
var mds *medicine.Svc
var profilecpu, index, version, build string

// cleanExit closes open datastores from the terminology and dm+d services and
// ends CPU Profile recording if enabled. Called on SIGTERM or as PostRun
// commmand from Cobra if not in a longrunning loop.
func cleanExit() error {
	if sct != nil {
		if err := sct.Close(); err != nil {
			return err
		}
	}
	if mds != nil {
		if err := mds.Close(); err != nil {
			return err
		}
	}
	if profilecpu != "" {
		pprof.StopCPUProfile()
	}
	return nil
}

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
		options := terminology.Options{
			Index:         args[0],
			IndexReadOnly: true,
		}
		// Overide options for index path it --index set to alternate directory
		if index != "" {
			options.Index = index
		}
		// Set readOnly to false if command in following map
		readWriteCommands := map[string]bool{"import": true, "precompute": true, "reset": true}
		if _, ok := readWriteCommands[cmd.CalledAs()]; ok {
			readOnly = false
		}
		// Special case for index command
		if cmd.CalledAs() == "index" {
			readOnly = true
			options.IndexReadOnly = false
		}

		// Create new terminology service
		sct, err = terminology.New(args[0], readOnly, options)
		if err != nil {
			return fmt.Errorf("couldn't open terminology datastore: %v", err)
		}
		sct.Version() // Print terminology version info

		// Create new dm+d service
		mds, err = medicine.New(sct, args[0], readOnly)
		if err != nil {
			return fmt.Errorf("couldn't open dm+d datastore: %v", err)
		}
		mds.Version() // Print dm+d version info

		// Graceful cleanup on exit (ctrl-c)
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			err := cleanExit()
			if err != nil {
				log.Fatalf("error cleaning up: %v", err)
			}
			os.Exit(1)
		}()

		return nil
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Receives version and build strings from main
func Execute(versionArg string, buildArg string) {
	version = versionArg
	build = buildArg
	defer cleanExit() // Run cleanup when executed command exits
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		cleanExit() // Attempt to cleanup and exit cleanly may throw more errors
		os.Exit(-1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profilecpu, "profile-cpu", "", "write cpu profile to `file` specified")
	rootCmd.PersistentFlags().StringVar(&index, "index", "", "use specified `directory` for search index instead of defaulting to <data-dir>")
}
