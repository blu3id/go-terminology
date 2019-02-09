package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dmdCmd represents the dmd command
var dmdCmd = &cobra.Command{
	Use:   "dmd",
	Short: "Commands for dm+d including import of data",
	Long:  `Commands for dm+d including import of data.`,
}

var importDmdCmd = &cobra.Command{
	Use:   "import <data-dir> <dmdv2-dir>",
	Short: "Import dm+d data files from specified directory",
	Long: `Import dm+d data files from specified directory. The directory should be the` + "\n" +
		`"nhsbsa_dmd_<version>" folder from the TRUD distribution. Version information` + "\n" +
		`is taken from the directory name so do NOT rename the folder.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("must specify input file(s)")
		}
		return mds.PerformImport(args[1])
	},
}

var testDmdCmd = &cobra.Command{
	Use:   "test <data-dir>",
	Short: "run test",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := mds.GetRoutesForVTM(90332006)
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n", r)
		for _, route := range r {
			f, err := mds.GetFormForVTMRoute(90332006, route)
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", f)

			t, err := mds.GetTypeForVTMRoute(90332006, route)
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", t)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dmdCmd)
	dmdCmd.AddCommand(importDmdCmd, testDmdCmd)
}
