package cluster

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove stored clusters",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.Remove(filepath.Join(viper.GetString("bitgodineDir"), "clusters.csv")); err != nil {
			logger.Error("Cluster rm", err, logger.Params{})
		}
	},
}

// func init() {
// }
