package cluster

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/persistent"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint/memory"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// recoveryCmd represents the recovery command
var recoveryCmd = &cobra.Command{
	Use:   "recovery",
	Short: "Recover persistent set structure",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Cluster recovery", "recoverying clusters...", logger.Params{})
		set := persistent.NewDisjointSet(dgraph.Instance(nil))
		mSet := memory.NewDisjointSet()

		if err := persistent.RestorePersistentSet(&set); err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			os.Exit(-1)
		}
		if err := set.RecoverPersistentSet(&mSet); err != nil {
			logger.Error("Blockchain", err, logger.Params{})
			os.Exit(-1)
		}

		logger.Info("Cluster recovery", "Cluster Successfully recovered", logger.Params{})
	},
}

func init() {
}
