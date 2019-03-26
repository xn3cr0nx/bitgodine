package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/parser"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Parses the blockchain to sync it",
	Long: `Parses the blockchain, from the last point,
if the synced is being previously performed.
The parsing stores blocks and transaction and creates clusters to provide
data representation to analyze the blockchain.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Sync", "sync called", logger.Params{})

		b := blockchain.Instance(BitcoinNet)
		b.Read()

		cltz := visitor.NewClusterizer()

		c := make(chan os.Signal, 1)
		interrupt := make(chan int)
		done := make(chan int)
		signal.Notify(c, os.Interrupt)
		go handleInterrupt(cltz, c, interrupt, done)

		parser.Walk(b, cltz, interrupt, done)
		cltzCount, err := cltz.Done()
		if err != nil {
			logger.Error("Blockchain test", err, logger.Params{})
		}
		fmt.Printf("Exported Clusters: %v\n", cltzCount)

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func handleInterrupt(cltz visitor.BlockchainVisitor, c chan os.Signal, interrupt, done chan int) {
	for sig := range c {
		logger.Info("Sync", "Killing the application", logger.Params{"signal": sig})
		interrupt <- 1
		cltzCount, err := cltz.Done()
		if err != nil {
			logger.Error("Sync", err, logger.Params{})
		}
		logger.Info("Sync", fmt.Sprintf("Exported Clusters: %v\n", cltzCount), logger.Params{})
		done <- 1
		os.Exit(1)
	}
}
