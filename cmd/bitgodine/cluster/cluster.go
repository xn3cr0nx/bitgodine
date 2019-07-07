package cluster

import (
	"github.com/spf13/cobra"
)

// ClusterCmd represents the cluster command
var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters",
	Long:  "",
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	ClusterCmd.AddCommand(lsCmd)
	ClusterCmd.AddCommand(rmCmd)
	ClusterCmd.AddCommand(exportCmd)
	ClusterCmd.AddCommand(storeCmd)
	ClusterCmd.AddCommand(tagCmd)
}
