package analyze

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/heuristics/analyze"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

var (
	blocksRange string
	from, to    int32
	plot        bool
)

// AnalyzeCmd represents the Analyze command
var AnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze transactions",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("fucking viper")
		// fmt.Println("args", strings.Join(viper.AllKeys(), " "))
		// fmt.Println("flags", strings.Join(cmd.LocalFlags().Args(), " "))
		chain := blockchain.Instance(chaincfg.Params{})
		chainHeight := chain.Height()
		var analysis [][]bool

		if viper.GetString("analyze.range") != "full" {
			heights := strings.Split(viper.GetString("analyze.range"), "-")
			start, err := strconv.Atoi(heights[0])
			if err != nil {
				logger.Error("Analyze", err, logger.Params{})
				return
			}
			end, err := strconv.Atoi(heights[1])
			if err != nil {
				logger.Error("Analyze", err, logger.Params{})
				return
			}
			if start > end {
				logger.Error("Analyze", errors.New("Starting height in range can't be major than end height"), logger.Params{"start": heights[0], "end": heights[1]})
				return
			}
			if end > int(chainHeight) {
				if start > end {
					logger.Error("Analyze", errors.New("The chain is not synced to that end point"), logger.Params{"start": heights[0], "end": heights[1], "chain_height": chainHeight})
					return
				}
			}
			analysis, err = analyze.Range(int32(start), int32(end))
			if err != nil {
				logger.Error("Analyze", err, logger.Params{})
				return
			}
		} else {
			var err error
			if int(chainHeight) < viper.GetInt("analyze.to") {
				logger.Error("Analyze", errors.New("The chain is not synced to that end point"), logger.Params{"chain_height": chainHeight})
				return
			}
			analysis, err = analyze.Range(int32(viper.GetInt("analyze.from")), int32(viper.GetInt("analyze.to")))
			if err != nil {
				logger.Error("Analyze", err, logger.Params{})
				return
			}
		}

		if viper.GetBool("analyze.plot") {
			fmt.Println("Implementeremo il plot")
		} else {
			analyze.Percentages(analysis)
		}
	},
}

func init() {
	AnalyzeCmd.AddCommand(txCmd)
	AnalyzeCmd.AddCommand(peelingCmd)
	AnalyzeCmd.AddCommand(forwardCmd)
	AnalyzeCmd.AddCommand(backwardCmd)

	AnalyzeCmd.PersistentFlags().StringVar(&blocksRange, "range", "full", "Specify the range the analysis should be applied on in the form \"start-end\"")
	viper.SetDefault("analyze.range", "full")
	viper.BindPFlag("analyze.range", AnalyzeCmd.PersistentFlags().Lookup("range"))

	AnalyzeCmd.PersistentFlags().Int32Var(&from, "from", 1, "Specify the block height the analysis should start from")
	viper.SetDefault("analyze.from", 1)
	viper.BindPFlag("analyze.from", AnalyzeCmd.PersistentFlags().Lookup("from"))

	AnalyzeCmd.PersistentFlags().Int32Var(&to, "to", 1000, "Specify the block height the analysis should reach to")
	viper.SetDefault("analyze.to", 1000)
	viper.BindPFlag("analyze.to", AnalyzeCmd.PersistentFlags().Lookup("to"))

	AnalyzeCmd.PersistentFlags().BoolVar(&plot, "plot", false, "Specify whether show the resulting plot of the analysis")
	viper.SetDefault("analyze.plot", false)
	viper.BindPFlag("analyze.plot", AnalyzeCmd.PersistentFlags().Lookup("plot"))
}
