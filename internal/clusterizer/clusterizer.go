package clusterizer

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/xn3cr0nx/bitgodine_code/internal/disjoint"
	"github.com/xn3cr0nx/bitgodine_code/internal/visitor"
	"github.com/xn3cr0nx/bitgodine_code/pkg/logger"
)

type Clusterizer struct {
	clusters *disjoint.DisjointSet
}

func NewClusterizer() Clusterizer {
	return Clusterizer{
		clusters: disjoint.NewDisjointSet(),
	}
}

func (c Clusterizer) VisitBlockBegin(block *btcutil.Block, height uint64) visitor.BlockItem {
	return nil
}

func (c Clusterizer) VisitBlockEnd(block *btcutil.Block, height uint64, blockItem visitor.BlockItem) {}

func (c Clusterizer) VisitTransactionBegin(block *visitor.BlockItem) visitor.TransactionItem {
	return hashset.New()
}

func (c Clusterizer) VisitTransactionInput(txIn wire.TxIn, block *visitor.BlockItem, txItem *visitor.TransactionItem, oItem visitor.Utxo) {
	// ignore coinbase
	if zeroHash, _ := chainhash.NewHash(make([]byte, 32)); txIn.PreviousOutPoint.Hash.IsEqual(zeroHash) {
		return
	}
	if oItem != "" {
		(*txItem).Add(oItem)
	}
}

func (c Clusterizer) VisitTransactionOutput(txOut wire.TxOut, blockItem *visitor.BlockItem, txItem *visitor.TransactionItem) (visitor.Utxo, error) {
	// txscript.GetScriptClass(txOut.Script).String()
	// _, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.Script, &blockchain.Instance().Network)
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, &chaincfg.MainNetParams)
	if err != nil {
		return "", err
	}
	if len(addresses) > 0 {
		// EncodeAddress returns always the address' P2PKH version
		return visitor.Utxo(addresses[0].EncodeAddress()), nil
	}
	return "", errors.New("Not able to extract address from PkScript")
}

func (c Clusterizer) VisitTransactionEnd(tx btcutil.Tx, blockItem *visitor.BlockItem, txItem *visitor.TransactionItem) {
	// skip transactions with just one input

	if (*txItem).Size() > 1 {
		txInputs := (*txItem).Values()
		lastAddress := txInputs[0].(visitor.Utxo)
		c.clusters.MakeSet(lastAddress)
		for _, address := range txInputs {
			c.clusters.MakeSet(address.(visitor.Utxo))
			c.clusters.Union(lastAddress, address.(visitor.Utxo))
			lastAddress = address.(visitor.Utxo)
		}
	}
}

func (c Clusterizer) Done() (visitor.DoneItem, error) {
	c.clusters.Finalize()
	logger.Info("Clusterizer", "Exporting clusters to CSV", logger.Params{"size": strconv.Itoa(c.clusters.Size())})
	file, err := os.Create("../../clusters.csv")
	if err != nil {
		logger.Error("Clusterizer", err, logger.Params{})
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for address, tag := range c.clusters.HashMap {
		// fmt.Printf("	tag %v, element %v\n", tag, c.clusters.Parent[tag])
		writer.Write([]string{string(address.(visitor.Utxo)), strconv.Itoa(c.clusters.Parent[tag])})
	}

	logger.Info("Clusterizer", "Exported clusters to CSV", logger.Params{"size": strconv.Itoa(c.clusters.Size())})
	return visitor.DoneItem(c.clusters.Size()), nil
}
