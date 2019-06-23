package tag

import (
	"github.com/spf13/cobra"
)

// TagCmd represents the tag command
var TagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
	Long:  "",
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	TagCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// TagCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// TagCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
