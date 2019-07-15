package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// sizeCmd represents the size command
var sizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Show size of cluster set",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cluster, err := dgraph.GetClusters()
		if err != nil {
			logger.Error("Cluster size", err, logger.Params{})
		}

		logger.Info("Cluster size", fmt.Sprintf("Cluster size: %v", cluster.Size), logger.Params{})
	},
}

// func init() {
// }
