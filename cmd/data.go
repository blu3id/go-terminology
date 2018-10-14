package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// dataCmd represents the data command
var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Commands for import, export and management of data",
	Long:  `Commands for import, export and management of data.`,
}

var importCmd = &cobra.Command{
	Use:   "import <data-dir> <REF2-dir> [REF2-dir2...]",
	Short: "Import SNOMED-CT data files from specified directories",
	Long:  `Import SNOMED-CT data files from specified directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("must specify input file(s)")
		}
		for i, filename := range args {
			if i == 0 {
				continue //skip data-dir
			}
			sct.PerformImport(filename)
		}
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export <data-dir>",
	Short: "Export expanded descriptions in delimited protobuf format",
	Long:  `Export expanded descriptions in delimited protobuf format to stdout.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return sct.Export()
	},
}

var indexCmd = &cobra.Command{
	Use:   "index <data-dir>",
	Short: "Build search index from currently loaded data",
	Long:  `Build the search index from currently loaded data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := sct.Index()
		time.Sleep(60 * time.Second) // Dirty hack for Bleve Moss storage to ensure data is written to disk
		return err
	},
}

var precomputeCmd = &cobra.Command{
	Use:   "precompute <data-dir>",
	Short: "Perform precomputations and optimisations",
	Long:  `Perform precomputations and optimisations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sct.PerformPrecomputations()
		return nil
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset <data-dir>",
	Short: "Clear precomputations and optimisations",
	Long:  `Clear precomputations and optimisations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sct.ClearPrecomputations()
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info <data-dir>",
	Short: "Print datastore statistics",
	Long:  `Print datastore statistics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := sct.GetStatistics()
		if err != nil {
			return err
		}
		fmt.Printf("%v", stats)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dataCmd)
	dataCmd.AddCommand(importCmd, exportCmd, indexCmd, precomputeCmd, resetCmd, infoCmd)
}
