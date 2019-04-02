package visitor

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/xn3cr0nx/bitgodine_code/internal/blocks"
)

func TestVisitTransactionOutput(t *testing.T) {
	f, _ := ioutil.ReadFile("/home/xn3cr0nx/.bitcoin/blocks/blk00000.dat")
	block, _ := blocks.Parse(&f)

	script := block.Transactions()[0].MsgTx().TxOut[0].PkScript
	fmt.Printf("block: %v, script: %v\n", block.Hash(), script)

	scriptClass, addresses, reqSigs, err := txscript.ExtractPkScriptAddrs(
		script, &chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Script Class:", scriptClass)
	fmt.Println("Addresses:", addresses)
	fmt.Println("Required Signatures:", reqSigs)

	addr := addresses[0].EncodeAddress()
	// addrHash := addr.AddressPubKeyHash()
	// fmt.Println("PubkKey", addr, "PubKeyHash", addrHash)
	fmt.Println("PubKeyHash", addr)

	// assert.Equal(t, "pubkeyhash", scriptClass.String())
	// assert.Equal(t, "12gpXQVcCL2qhTNQgyLVdCFG2Qs2px98nV", addresses[0].String())
	// assert.Equal(t, 1, reqSigs)
}
