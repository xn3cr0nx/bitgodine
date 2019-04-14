package analysis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/analysis"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var (
	blocksRange string
	from, to    int32
	plot        bool
)

// AnalysisCmd represents the analysis command
var AnalysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "Analyze transactions",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		chain := blockchain.Instance(chaincfg.Params{})
		chainHeight := chain.Height()
		var a [][]bool
		var start, end int
		var err error

		if viper.GetString("analysis.range") != "full" {
			heights := strings.Split(viper.GetString("analysis.range"), "-")
			start, err = strconv.Atoi(heights[0])
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				return
			}
			end, err = strconv.Atoi(heights[1])
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				return
			}
			if start > end {
				logger.Error("Analysis", errors.New("Starting height in range can't be major than end height"), logger.Params{"start": heights[0], "end": heights[1]})
				return
			}
			if end > int(chainHeight) {
				logger.Error("Analysis", errors.New("The chain is not synced to that end point"), logger.Params{"start": heights[0], "end": heights[1], "chain_height": chainHeight})
				return
			}
			a, err = analysis.Range(int32(start), int32(end))
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				return
			}
		} else {
			start, end = viper.GetInt("analysis.from"), viper.GetInt("analysis.to")
			if start > end {
				logger.Error("Analysis", errors.New("Starting height should be minor than end height"), logger.Params{"chain_height": chainHeight})
				return
			}
			if int(chainHeight) < end {
				logger.Error("Analysis", errors.New("The chain is not synced to that end point"), logger.Params{"chain_height": chainHeight})
				return
			}
			a, err = analysis.Range(int32(start), int32(end))
			if err != nil {
				logger.Error("Analysis", err, logger.Params{})
				return
			}
		}

		if len(a) == 0 {
			logger.Error("Analysis", errors.New("No output to produce. No transaction in the range analyzed. It means none on them has at least two outputs"), logger.Params{})
			return
		}

		if viper.GetBool("analysis.plot") {
			analysis.Plot(a, start, end)
		} else {
			fmt.Printf("%d transactions analyzed\n", len(a))
			analysis.Percentages(a)
		}
	},
}

func init() {
	AnalysisCmd.AddCommand(txCmd)
	AnalysisCmd.AddCommand(peelingCmd)
	AnalysisCmd.AddCommand(forwardCmd)
	AnalysisCmd.AddCommand(backwardCmd)

	AnalysisCmd.PersistentFlags().StringVar(&blocksRange, "range", "full", "Specify the range the analysis should be applied on in the form \"start-end\"")
	viper.SetDefault("analysis.range", "full")
	viper.BindPFlag("analysis.range", AnalysisCmd.PersistentFlags().Lookup("range"))

	AnalysisCmd.PersistentFlags().Int32Var(&from, "from", 1, "Specify the block height the analysis should start from")
	viper.SetDefault("analysis.from", 1)
	viper.BindPFlag("analysis.from", AnalysisCmd.PersistentFlags().Lookup("from"))

	AnalysisCmd.PersistentFlags().Int32Var(&to, "to", 1000, "Specify the block height the analysis should reach to")
	viper.SetDefault("analysis.to", 1000)
	viper.BindPFlag("analysis.to", AnalysisCmd.PersistentFlags().Lookup("to"))

	AnalysisCmd.PersistentFlags().BoolVar(&plot, "plot", false, "Specify whether show the resulting plot of the analysis")
	viper.SetDefault("analysis.plot", false)
	viper.BindPFlag("analysis.plot", AnalysisCmd.PersistentFlags().Lookup("plot"))
}
