package block

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// height, err := strconv.Atoi(args[0])
		// if err == nil {
		// }
		if args[0] == "last" {
			if err := blocks.RemoveLast(); err != nil {
				logger.Error("Block rm", err, logger.Params{})
				os.Exit(-1)
			}
			os.Exit(1)
		}

		dir, err := ioutil.ReadDir(viper.GetString("dbDir"))
		if err != nil {
			logger.Error("Block rm", err, logger.Params{})
		}
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{viper.GetString("dbDir"), d.Name()}...))
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
