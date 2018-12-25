package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"io/ioutil"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

const eskFilename = "esk.json"

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//type extendedSecretKey struct {
//	Sk   *btcec.PrivateKey
//	Pk   *btcec.PublicKey
//	Addr string
//}
//
//func newExtendedSecretKey(sk *btcec.PrivateKey, pk *btcec.PublicKey, addr string) *extendedSecretKey {
//	return &extendedSecretKey{
//		Sk:   sk,
//		Pk:   pk,
//		Addr: addr,
//	}
//}

func main() {
	newAddress := flag.Bool("newaddress", false, "")
	sign := flag.Bool("sign", false, "")
	encodedTxForSignInHex := flag.String("tx_for_sign", "", "")
	subScriptInHex := flag.String("sub_script", "", "")
	flag.Parse()

	if *newAddress {
		sk, err := btcec.NewPrivateKey(btcec.S256())
		checkErr(err)
		pk := sk.PubKey()
		pkh := btcutil.Hash160(pk.SerializeCompressed())
		// fmt.Println(hex.EncodeToString(pkh))

		addr, err := btcutil.NewAddressPubKeyHash(pkh, &chaincfg.RegressionNetParams)
		checkErr(err)

		// persist key
		//esk := newExtendedSecretKey(sk, pk, addr.String())
		//encodedEsk, err := json.Marshal(esk)
		//checkErr(err)
		//checkErr(ioutil.WriteFile(eskFilename, encodedEsk, 0644))

		checkErr(ioutil.WriteFile(eskFilename, sk.Serialize(), 0644))

		fmt.Println(addr.String())
		return
	}

	if *sign {
		// restore key
		//encodedEsk, err := ioutil.ReadFile(eskFilename)
		//checkErr(err)
		//esk := extendedSecretKey{}
		//checkErr(json.Unmarshal(encodedEsk, &esk))

		encodedSk, err := ioutil.ReadFile(eskFilename)
		checkErr(err)
		sk, _ := btcec.PrivKeyFromBytes(btcec.S256(), encodedSk)

		// deserialize tx
		txForSign := wire.MsgTx{}
		encodedTxForSign, err := hex.DecodeString(*encodedTxForSignInHex)
		checkErr(err)
		buff := bytes.NewBuffer(encodedTxForSign)
		checkErr(txForSign.Deserialize(buff))

		subScript, err := hex.DecodeString(*subScriptInHex)
		checkErr(err)

		sig, err := txscript.RawTxInSignature(&txForSign, 0, subScript, txscript.SigHashAll, sk)
		checkErr(err)

		fmt.Println(hex.EncodeToString(sig))
	}
}
