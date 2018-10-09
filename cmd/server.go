package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wardle/go-terminology/server"
)

var port int
var rpc bool

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server <data-dir>",
	Short: "Runs the terminology server",
	Long:  `The server command runs the terminology server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch rpc {
		case true:
			server.RunRPCServer(sct, port)
		case false:
			server.RunServer(sct, port)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "port to use when running server")
	serverCmd.Flags().BoolVar(&rpc, "rpc", false, "run RPC server")
}
