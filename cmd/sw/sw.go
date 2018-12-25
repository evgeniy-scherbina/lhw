package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"io/ioutil"
	"os"
)

const (
	defaultHost     = "localhost:18443"
	defaultUser     = "user"
	defaultPassword = "password"
)

func checkErr(err error, funcName string) {
	if err != nil {
		fmt.Printf("funcName: %v, err: %v\n", funcName, err)
		os.Exit(1)
	}
}

func assert(ok bool) {
	if !ok {
		panic("assertion failed")
	}
}

func main() {
	init := flag.Bool("init", false, "")
	encodedAddrP := flag.String("addr", "", "")
	publish := flag.Bool("publish", false, "")
	sigInHex := flag.String("signature", "", "")
	flag.Parse()

	cfg := &rpcclient.ConnConfig{
		Host:         defaultHost,
		User:         defaultUser,
		Pass:         defaultPassword,
		DisableTLS:   true,
		HTTPPostMode: true,
	}

	client, err := rpcclient.New(cfg, nil)
	checkErr(err, "rpcclient.New")

	if *init {
		// send coins to lhw
		_, err = client.Generate(110)
		checkErr(err, "client.Generate")

		addr, err := btcutil.DecodeAddress(*encodedAddrP, &chaincfg.RegressionNetParams)
		checkErr(err, "btcutil.DecodeAddress")

		amt := int64(1e8)
		txHash, err := client.SendToAddress(addr, btcutil.Amount(amt))
		checkErr(err, "client.SendToAddress")
		fmt.Printf("txHash: %v\n", txHash)

		_, err = client.Generate(6)
		checkErr(err, "client.Generate")

		// try to spend newly sent coins
		tx, err := client.GetRawTransaction(txHash)
		checkErr(err, "client.GetRawTransaction")

		// find output index
		listTxOut := tx.MsgTx().TxOut
		outputIndex := uint32(0)
		subScript := make([]byte, 0)
		assert(len(listTxOut) == 2)
		for index, txOut := range listTxOut {
			if txOut.Value == amt {
				outputIndex = uint32(index)
				subScript = txOut.PkScript
			}
		}
		fmt.Printf("outputIndex: %v\n", outputIndex)

		// construct spending tx
		txIn := []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{
					Hash:  *txHash,
					Index: outputIndex,
				},
				Sequence: 0xFFFFFFFF,
			},
		}

		txOut := []*wire.TxOut{
			{
				Value:    amt - 1e4, // subtract fee
				PkScript: addr.ScriptAddress(),
			},
		}

		spendingTx := &wire.MsgTx{
			Version:  0,
			TxIn:     txIn,
			TxOut:    txOut,
			LockTime: 0,
		}

		buff := &bytes.Buffer{}
		checkErr(spendingTx.Serialize(buff), "spendingTx.Serialize")
		fmt.Printf("tx_for_sign: %v\n", hex.EncodeToString(buff.Bytes()))
		fmt.Printf("sub_script: %v\n", hex.EncodeToString(subScript))

		checkErr(ioutil.WriteFile("spending_tx.data", buff.Bytes(), 0644), "ioutil.WriteFile")
		return
	}

	if *publish {
		encodedSpendingTx, err := ioutil.ReadFile("spending_tx.data")
		checkErr(err, "ioutil.ReadFile")

		spendingTx := wire.MsgTx{}
		buff := bytes.NewBuffer(encodedSpendingTx)
		checkErr(spendingTx.Deserialize(buff))

		signature, err := hex.DecodeString(sigInHex)
		checkErr(err, "hex.DecodeString")
		spendingTx.TxIn[0].SignatureScript = signature

		allowHighFees := false
		spendingTxHash, err := client.SendRawTransaction(&spendingTx, allowHighFees)
		checkErr(err, "client.SendRawTransaction")
		fmt.Printf("spendingTxHash: %v\n", spendingTxHash)
	}
}