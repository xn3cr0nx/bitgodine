package clusterizer

import (
	"io/ioutil"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/xn3cr0nx/bitgodine/internal/blockchain"
	"github.com/xn3cr0nx/bitgodine/internal/disjoint"
	"github.com/xn3cr0nx/bitgodine/internal/transactions"
	"github.com/xn3cr0nx/bitgodine/internal/visitor"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

type Clusterizer struct {
	clusters *disjoint.DisjointSet
}

func NewClusterizer() Clusterizer {
	return Clusterizer{
		clusters: disjoint.NewDisjointSet(),
	}
}

func (c *Clusterizer) visitBlockBegin(block visitor.BlockItem, height uint64) {}

func (c *Clusterizer) visitTransactionBegin(block visitor.BlockItem) visitor.TransactionItem {
	return hashset.New()
}

func (c *Clusterizer) visitTransactionInput(txIn tx.TxInput, block visitor.BlockItem, txItem visitor.TransactionItem, oItem visitor.OutputItem) {
	// ignore coinbase
	if zeroHash, _ := chainhash.NewHash(make([]byte, 32)); txIn.PrevHash.IsEqual(zeroHash) {
		return
	}
	if oItem != nil {
		txItem.Add(oItem)
	}
}

// TODO: this fuction should be tested
func (c *Clusterizer) visitTransactionOutput(txOut tx.TxOutput, blockItem visitor.BlockItem, txItem visitor.TransactionItem) (visitor.OutputItem, error) {
	// txscript.GetScriptClass(txOut.Script).String()
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(txOut.Script, &blockchain.Instance().Network)
	return addresses[0], err
}

func (c *Clusterizer) visitTransactionEnd(tx tx.Tx, blockItem visitor.BlockItem, txItem visitor.TransactionItem) {
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

func (c *Clusterizer) done() (uint, error) {
	c.clusters.Finalize()
	logger.Info("Clusterizer", "Exporting clusters to CSV", logger.Params{"size": string(c.clusters.Size())})
	for address, tag := range c.clusters.HashMap {
		ioutil.WriteFile("clusters.csv", append(address.ScriptAddress(), byte(c.clusters.Parent[tag])), 0777)
	}

	logger.Info("Clusterizer", "Exported clusters to CSV", logger.Params{"size": string(c.clusters.Size())})
	return c.clusters.Size(), nil
}
