package tx

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine/internal/storage/dgraph"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove stored transactions",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		dg := dgraph.Instance(dgraph.Conf(), nil)
		if err := dg.Setup(); err != nil {
			logger.Error("Bitgodine", err, logger.Params{})
			os.Exit(-1)
		}

		if err := dg.Empty(); err != nil {
			logger.Error("transactions rm", err, logger.Params{})
			return
		}
	},
}

// func init() {

// 	// Here you will define your flags and configuration settings.

// 	// Cobra supports Persistent Flags which will work for this command
// 	// and all subcommands, e.g.:
// 	// rmCmd.PersistentFlags().String("foo", "", "A help for foo")

// 	// Cobra supports local flags which will only run when this command
// 	// is called directly, e.g.:
// 	// rmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }
