package clusterizer

import (
	"io/ioutil"

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

func (c Clusterizer) VisitTransactionInput(txIn wire.TxIn, block *visitor.BlockItem, txItem *visitor.TransactionItem, oItem visitor.OutputItem) {
	// ignore coinbase
	if zeroHash, _ := chainhash.NewHash(make([]byte, 32)); txIn.PreviousOutPoint.Hash.IsEqual(zeroHash) {
		return
	}
	if oItem != nil {
		(*txItem).Add(oItem)
	}
}

// TODO: this fuction should be tested
func (c Clusterizer) VisitTransactionOutput(txOut wire.TxOut, blockItem *visitor.BlockItem, txItem *visitor.TransactionItem) (visitor.OutputItem, error) {
	// txscript.GetScriptClass(txOut.Script).String()
	// _, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.Script, &blockchain.Instance().Network)
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, &chaincfg.MainNetParams)
	return addresses[0], err
}

func (c Clusterizer) VisitTransactionEnd(tx btcutil.Tx, blockItem *visitor.BlockItem, txItem visitor.TransactionItem) {
	// skip transactions with just one input
	if txItem.Size() > 1 {
		txInputs := txItem.Values()
		lastAddress := txInputs[txItem.Size()-1].(btcutil.Address)
		c.clusters.MakeSet(lastAddress)
		for _, address := range txInputs {
			c.clusters.MakeSet(address.(btcutil.Address))
			c.clusters.Union(lastAddress, address.(btcutil.Address))
			lastAddress = address.(btcutil.Address)
		}
	}
}

func (c Clusterizer) Done() (visitor.DoneItem, error) {
	c.clusters.Finalize()
	logger.Info("Clusterizer", "Exporting clusters to CSV", logger.Params{"size": string(c.clusters.Size())})
	for address, tag := range c.clusters.HashMap {
		ioutil.WriteFile("clusters.csv", append(address.ScriptAddress(), byte(c.clusters.Parent[tag])), 0777)
	}

	logger.Info("Clusterizer", "Exported clusters to CSV", logger.Params{"size": string(c.clusters.Size())})
	return visitor.DoneItem(c.clusters.Size()), nil
}
