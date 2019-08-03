package skipped

import (
	"github.com/spf13/cobra"
)

// SkippedCmd represents the skipped command
var SkippedCmd = &cobra.Command{
	Use:   "skipped",
	Short: "Skipped stored blocks operations",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	SkippedCmd.AddCommand(rmCmd)
	SkippedCmd.AddCommand(lsCmd)
	SkippedCmd.AddCommand(recoveryCmd)
}
