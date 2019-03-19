package block

import (
	"github.com/spf13/cobra"
)

// BlockCmd represents the block command
var BlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage block",
	Long:  "",
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	BlockCmd.AddCommand(lsCmd)
	BlockCmd.AddCommand(rmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// BlockCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// BlockCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
