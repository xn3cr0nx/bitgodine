package address

import (
	"github.com/spf13/cobra"
)

// AddressCmd represents the Address command
var AddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Manage Addresses",
	Long:  "",
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func init() {
	AddressCmd.AddCommand(occurencesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// AddressCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// AddressCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
