package block

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove stored blocks",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := ioutil.ReadDir(viper.GetString("dbDir"))
		if err != nil {
			logger.Error("transactions rm", err, logger.Params{})
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
