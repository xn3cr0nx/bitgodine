package main

import (
	"fmt"

	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/parser"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Generate clusters file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Cluster", "cluster called", logger.Params{})

		b := blockchain.Instance(BitcoinNet)
		b.Read()
		// if len(b.Maps) == 0 {
		// 	fmt.Println("You need to sync the blockchain, call bitgodine sync")
		// 	return
		// }
		cltz := visitor.NewClusterizer()
		parser.Walk(b, cltz)
		cltzCount, err := cltz.Done()
		if err != nil {
			logger.Error("Blockchain test", err, logger.Params{})
		}
		fmt.Printf("Exported Clusters: %v\n", cltzCount)
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
