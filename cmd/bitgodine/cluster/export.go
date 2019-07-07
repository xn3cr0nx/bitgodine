package cluster

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/persistent"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var output string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export clusters to the specified output",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Cluster Export", "exporting clusters...", logger.Params{})
		set := persistent.NewDisjointSet(dgraph.Instance(nil))
		if err := persistent.RestorePersistentSet(&set); err != nil {
			if err.Error() != "Cluster not found" {
				logger.Error("Cluster export", err, logger.Params{})
				os.Exit(-1)
			}
		}
		cltz := visitor.NewClusterizer(&set)
		if _, err := cltz.Done(); err != nil {
			logger.Error("Cluster export", err, logger.Params{})
			os.Exit(-1)
		}
	},
}

func init() {
}
