package block

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/dbblocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gosuri/uiprogress"
)

// skippedCmd represents the skipped command
var skippedCmd = &cobra.Command{
	Use:   "skipped",
	Short: "Skipped stored blocks operations",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			logger.Error("Block skipped", errors.New("Missing command"), logger.Params{})
			os.Exit(1)
		}

		if args[0] == "recovery" {
			skipped := make(map[chainhash.Hash]blocks.Block)
			skippedBlocksStorage, err := dbblocks.NewDbBlocks(&db.Config{
				Dir: viper.GetString("dbDir"),
			})
			if err != nil {
				logger.Error("Bitgodine", err, logger.Params{})
				os.Exit(-1)
			}
			b := blockchain.Instance(chaincfg.MainNetParams)
			b.Read()

			head, err := b.Head()
			if err != nil {
				logger.Error("Block skipped", err, logger.Params{})
				os.Exit(-1)
			}

			var rawChain [][]uint8
			for _, ref := range b.Maps {
				rawChain = append(rawChain, []uint8(ref))
			}
			if err := recoverSkipped(&head, skippedBlocksStorage, &rawChain, &skipped); err != nil {
				logger.Error("Block skipped", err, logger.Params{})
				os.Exit(-1)
			}
		}

	},
}

func recoverSkipped(head *blocks.Block, db *dbblocks.DbBlocks, chain *[][]uint8, skipped *map[chainhash.Hash]blocks.Block) error {
	fmt.Println("")
	uiprogress.Start()
	bar := uiprogress.AddBar(len((*chain)[0])).AppendCompleted().PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Recovering skipped blocks: %v/%v", b.Current(), len((*chain)[0]))
	})

	for _, slice := range *chain {
		for len(slice) > 0 {
			initLen := len(slice)
			block, err := blocks.Parse(&slice)
			if err != nil {
				return err
			}

			if block.Hash().IsEqual(head.Hash()) {
				return nil
			}

			if _, err = dgraph.GetBlockFromHash(block.Hash().String()); err != nil {
				if err.Error() == "Block not found" {
					if err := db.StoreBlockPrevHash(block); err != nil {
						return err
					}
				} else {
					return err
				}
			}

			for i := 0; i < initLen-len(slice); i++ {
				bar.Incr()
			}
		}
	}
	uiprogress.Stop()
	return errors.New("Parsed the entire chain, head not found")
}
