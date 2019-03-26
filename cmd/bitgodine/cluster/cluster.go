package cluster

import (
	"github.com/spf13/cobra"
)

// ClusterCmd represents the cluster command
var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster",
	Long:  "",
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	ClusterCmd.AddCommand(lsCmd)
	ClusterCmd.AddCommand(rmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ClusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ClusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
