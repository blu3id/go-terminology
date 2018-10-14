package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var compactIndexCmd = &cobra.Command{
	Use:   "compact <data-dir>",
	Short: "Manually compact Bleve leveldb search index files",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := filepath.Join(args[0], "bleve_index", "store")
		// Overide index path if --index set to alternate directory
		if index != "" {
			path = filepath.Join(index, "store")
		}

		fmt.Printf("%+v\n", path)
		options := opt.Options{CompactionTableSizeMultiplier: 2}
		db, err := leveldb.OpenFile(path, &options)
		defer db.Close()
		if err != nil {
			return err
		}
		err = db.CompactRange(util.Range{})
		if err != nil {
			return err
		}

		var st leveldb.DBStats
		db.Stats(&st)
		fmt.Printf("%+v\n", st)
		return db.Close()
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
	indexCmd.AddCommand(compactIndexCmd)
}
