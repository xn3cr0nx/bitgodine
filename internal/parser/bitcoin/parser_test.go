package bitcoin_test

import (
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	. "github.com/xn3cr0nx/bitgodine/internal/parser/bitcoin"
	"github.com/xn3cr0nx/bitgodine/internal/storage"
	"github.com/xn3cr0nx/bitgodine/internal/storage/badger"
	"github.com/xn3cr0nx/bitgodine/internal/utxoset"
	"github.com/xn3cr0nx/bitgodine/pkg/cache"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

// Integration tests on blockchain parsing. Taking into consideration to have bitcoin data dir
var _ = Describe("Blockchain", func() {
	var (
		// bp   *Parser
		b    *bitcoin.Blockchain
		utxo *utxoset.UtxoSet
		db   storage.DB
	)

	BeforeSuite(func() {
		logger.Setup()

		ca, err := cache.NewCache(nil)
		Expect(err).ToNot(HaveOccurred())
		conf := &badger.Config{
			Dir: filepath.Join(".", "test"),
		}
		bdg, err := badger.NewBadger(conf, false)
		db, err = badger.NewKV(bdg, ca)
		Expect(err).ToNot(HaveOccurred())
		Expect(db).ToNot(BeNil())
		Expect(err).ToNot(HaveOccurred())

		skippedBlocksStorage := NewSkipped()
		utxo = utxoset.Instance(utxoset.Conf("", false))
		b = bitcoin.NewBlockchain(db, chaincfg.MainNetParams)
		b.Read("")

		NewParser(b, nil, db, skippedBlocksStorage, utxo, ca, nil, nil)
	})

	AfterEach(func() {
		viper.SetDefault("dbDir", filepath.Join(".", "test"))
		err := db.Empty()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Parsing data files", func() {
		It("Parsing a specific block extracted from files", func() {
			target := "000000000000021a070be6856ee21aaa432aa5d4daf4e754f8c2068af9ab3a6e"
			var blockTarget *bitcoin.Block
			var chain [][]uint8
			for _, ref := range b.Maps {
				chain = append(chain, []uint8(ref))
			}

			for _, slice := range chain {
				for len(slice) > 0 {
					block, err := bitcoin.Parse(&slice)
					Expect(err).ToNot(HaveOccurred())
					if block.Hash().String() == target {
						blockTarget = block
						break
					}
				}
				if blockTarget != nil {
					break
				}
			}
			Expect(blockTarget).ToNot(BeNil())

			err := blockTarget.Store(db)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("Testing walk the blockchain", func() {

			It("should correctly parse consequently stored blocks", func() {
			})

		})

	})
})
