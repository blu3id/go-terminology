package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/wardle/go-terminology/analysis"
)

var print, dof bool
var reduceDof, minDistance int

var analysisCmd = &cobra.Command{
	Use:   "analysis <data-dir> <input> [input2...]",
	Short: "Run analysis tools against specified files",
	Long:  `Run analysis tools against specified files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("must specify input file(s)")
		}
		if !print && !dof {
			return fmt.Errorf("at least one of --print or --dof must be specified")
		}

		if print {
			for i, filename := range args {
				if i == 0 {
					continue //skip data-dir
				}
				f, err := os.Open(filename)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
				reader := bufio.NewReader(f)
				if err := analysis.Print(sct, reader); err != nil {
					log.Fatal(err)
				}
			}
		}

		if dof {
			for i, filename := range args {
				if i == 0 {
					continue //skip data-dir
				}
				f, err := os.Open(filename)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()
				reader := bufio.NewReader(f)
				if reduceDof > 0 {
					r := analysis.NewReducer(sct, reduceDof, minDistance)
					if err := r.Reduce(reader, os.Stdout); err != nil {
						log.Fatal(err)
					}
				} else {
					factors, err := analysis.NumberFactors(reader)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(factors)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(analysisCmd)

	analysisCmd.Flags().BoolVar(&print, "print", false, "print information for each identifier in file(s) specified")
	analysisCmd.Flags().BoolVar(&dof, "dof", false, "dimensionality analysis and reduction for file(s) specified")
	analysisCmd.Flags().IntVarP(&reduceDof, "reduce", "r", 0, "reduce number of factors to specified number (default 0)")
	analysisCmd.Flags().IntVarP(&minDistance, "minimumDistance", "d", 3, "minimum distance from root")
}
