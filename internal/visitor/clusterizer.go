package visitor

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/viper"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint"
	txs "github.com/xn3cr0nx/bitgodine_code/internal/transactions"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

// Clusterizer struct containing the disjoint set data structure
type Clusterizer struct {
	clusters disjoint.DisjointSet
}

// NewClusterizer returns a new clusterizer
func NewClusterizer(set disjoint.DisjointSet) Clusterizer {
	return Clusterizer{
		clusters: set,
	}
}

func (c Clusterizer) VisitBlockBegin(block *blocks.Block, height int32) BlockItem {
	return nil
}

func (c Clusterizer) VisitBlockEnd(block *blocks.Block, height int32, blockItem BlockItem) {}

func (c Clusterizer) VisitTransactionBegin(block *BlockItem) TransactionItem {
	return hashset.New()
}

func (c Clusterizer) VisitTransactionInput(txIn wire.TxIn, block *BlockItem, txItem *TransactionItem, utxo Utxo) {
	// ignore coinbase
	if zeroHash, _ := chainhash.NewHash(make([]byte, 32)); txIn.PreviousOutPoint.Hash.IsEqual(zeroHash) {
		return
	}
	if utxo != "" {
		(*txItem).Add(utxo)
	}
}

func (c Clusterizer) VisitTransactionOutput(txOut wire.TxOut, blockItem *BlockItem, txItem *TransactionItem) (Utxo, error) {
	// txscript.GetScriptClass(txOut.Script).String()
	// _, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.Script, &blockchain.Instance().Network)
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	if len(addresses) > 0 {
		// EncodeAddress returns always the address' P2PKH version
		return Utxo(addresses[0].EncodeAddress()), nil
	}
	return "", errors.New("Not able to extract address from PkScript")
}

// VisitTransactionEnd implements first heuristic (all input are from the same user) and clusterize the input in the disjoint set
func (c Clusterizer) VisitTransactionEnd(tx txs.Tx, blockItem *BlockItem, txItem *TransactionItem) {
	// skip transactions with just one input
	if (*txItem).Size() > 1 && !tx.IsCoinjoin() {
		txInputs := (*txItem).Values()
		lastAddress := txInputs[0].(Utxo)
		logger.Debug("Clusterizer", "Enhancing disjoint set", logger.Params{"last_address": lastAddress})
		c.clusters.MakeSet(lastAddress)
		for _, address := range txInputs {
			c.clusters.MakeSet(address.(Utxo))
			c.clusters.Union(lastAddress, address.(Utxo))
			lastAddress = address.(Utxo)
		}
	}
}

// Done finalizes the operations of the clusterizer exporting its content to a csv file
func (c Clusterizer) Done() (DoneItem, error) {
	c.clusters.Finalize()
	logger.Info("Clusterizer", "Exporting clusters to CSV", logger.Params{"size": c.clusters.Size()})
	file, err := os.Create(fmt.Sprintf("%s/clusters.csv", viper.GetString("outputDir")))
	if err != nil {
		logger.Error("Clusterizer", err, logger.Params{})
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for address, tag := range c.clusters.GetHashMap() {
		// fmt.Printf("	tag %v, element %v\n", tag, c.clusters.Parent[tag])
		writer.Write([]string{string(address.(Utxo)), strconv.Itoa(c.clusters.GetParent(tag))})
	}

	logger.Info("Clusterizer", "Exported clusters to CSV", logger.Params{"size": c.clusters.Size()})
	return DoneItem(c.clusters.Size()), nil
}
