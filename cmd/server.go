package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wardle/go-terminology/server"
)

var (
	port    int
	rpc     bool
	address string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server <data-dir>",
	Short: "Runs the terminology server",
	Long:  `The server command runs the terminology server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server.Serve(sct, mds, address+":"+strconv.Itoa(port))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to use when running server")
	serverCmd.Flags().StringVarP(&address, "interface", "i", "", "interface to bind to when running server (default :)")
	//serverCmd.Flags().BoolVar(&rpc, "rpc", false, "run RPC server")
}
