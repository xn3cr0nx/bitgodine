package block

import (
	"errors"
	"os"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xn3cr0nx/bitgodine_code/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine_code/internal/db"
	"github.com/xn3cr0nx/bitgodine_code/internal/db/dbblocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/dgraph"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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

			if err := restoreSkipped(skippedBlocksStorage, &skipped); err != nil {
				logger.Error("Block skipped", err, logger.Params{})
				os.Exit(-1)
			}

			var rawChain [][]uint8
			for _, ref := range b.Maps {
				rawChain = append(rawChain, []uint8(ref))
			}
			if err := recoverSkipped(skippedBlocksStorage, &rawChain, &skipped); err != nil {
				logger.Error("Block skipped", err, logger.Params{})
				os.Exit(-1)
			}
		}

	},
}

func recoverSkipped(db *dbblocks.DbBlocks, chain *[][]uint8, skipped *map[chainhash.Hash]blocks.Block) error {
	for _, slice := range *chain {
		for len(slice) > 0 {
			block, err := blocks.Parse(&slice)
			if err != nil {
				return err
			}
			if _, err = dgraph.GetBlockFromHash(block.Hash().String()); err != nil {
				if err.Error() == "Block not found" {
					return nil
				}
				return err
			}

			if _, ok := (*skipped)[block.MsgBlock().Header.PrevBlock]; !ok {
				fmt.Println("Storing missing block", block.Hash().String())
				(*skipped)[block.MsgBlock().Header.PrevBlock] = *block
				if err := db.StoreBlockPrevHash(block); err != nil {
					return err
				}
			}
		}
	}
	return nil
}


func restoreSkipped(db *dbblocks.DbBlocks, skipped *map[chainhash.Hash]blocks.Block) error {
	cachedSkipped, err := db.GetAll()
	if err != nil {
		return err
	}
	logger.Info("Blockchain", "Restoring skipped blocks", logger.Params{"n_blocks": len(cachedSkipped)})
	for _, skip := range cachedSkipped {
		(*skipped)[skip.MsgBlock().Header.PrevBlock] = skip
	}
	return nil
}
