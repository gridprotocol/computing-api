package cmd

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gridprotocol/computing-api/lib/auth"
	"github.com/gridprotocol/computing-api/lib/utils"
	"github.com/urfave/cli/v2"
)

var SignCmd = &cli.Command{
	Name:  "sign",
	Usage: "sign with a ts to get a cookie",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "sk",
			Usage: "set the sk to sign",
			Value: "",
		},
	},
	Action: func(ctx *cli.Context) error {
		sk := ctx.String("sk")

		// get current timestamp
		timestamp := time.Now().Unix()
		ts := utils.Uint64ToString(uint64(timestamp))

		// enclose ts and calc hash
		hash := auth.Hash([]byte(auth.EncloseEth(ts)))

		// sign with hash and sk
		sig, err := auth.Sign(hash, sk)
		if err != nil {
			panic(err)
		}
		// to string
		strSig := hex.EncodeToString(sig)
		fmt.Printf("ts: %s, sig: %s\n", ts, strSig)

		// get address
		address := hex.EncodeToString(auth.SigToAddress(hash, sig))

		fmt.Println("cookie request:")
		fmt.Printf("http://localhost:12346/greet/cookie?ts=%s&user=0x%s&sig=0x%s\n", ts, address, strSig)

		return nil
	},
}
