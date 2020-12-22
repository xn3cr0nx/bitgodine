package transaction

import (
	"github.com/spf13/cobra"
)

// TransactionCmd represents the Transaction command
var TransactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Manage transactions",
	Long:  "",
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	TransactionCmd.AddCommand(lsCmd)
	TransactionCmd.AddCommand(rmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// TransactionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// TransactionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
