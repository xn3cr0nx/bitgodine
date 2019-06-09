package cluster

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/persistent"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

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
				logger.Error("Blockchain", err, logger.Params{})
				os.Exit(-1)
			}
		}
		cltz := visitor.NewClusterizer(&set)
		cltzCount, err := cltz.Done()
		if err != nil {
			logger.Error("Blockchain test", err, logger.Params{})
		}
		fmt.Printf("Exported Clusters: %v\n", cltzCount)
	},
}

// func init() {
// }
