package cluster

import (
	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show list of stored clusters",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// func init() {
// }
