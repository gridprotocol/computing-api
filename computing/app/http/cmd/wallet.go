package cmd

import (
	"fmt"
	"log"

	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/keystore"
	"github.com/urfave/cli/v2"
)

// init config
func init() {
	// parse config file
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}
}

var WalletCmd = &cli.Command{
	Name:  "wallet",
	Usage: "wallet management",
	Subcommands: []*cli.Command{
		initCmd,
		importCmd,
		showCmd,
		useCmd,
	},
}

// create a new wallet
var initCmd = &cli.Command{
	Name:  "init",
	Usage: "create a new wallet",
	Flags: []cli.Flag{

		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "set the keystore path",
			Value:   "./.keystore",
		},

		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"pw"},
			Usage:   "set the key password",
			Value:   "computing",
		},
	},
	Action: func(ctx *cli.Context) error {

		p := ctx.String("path")
		pw := ctx.String("pw")

		if p == "" {
			return fmt.Errorf("the repo path must be given")
		}
		if pw == "" {
			return fmt.Errorf("the password of the wallet must be given")
		}

		logger.Debug("new keystore")
		// keystore
		ks, err := keystore.NewKeyStore(p)
		if err != nil {
			return err
		}

		logger.Debug("new key")
		// key
		ki, err := keystore.NewKey()
		if err != nil {
			return err
		}

		logger.Debug("put key")
		// store key into keyjson
		ks.Put(ki.Address(), pw, *ki)

		// set address
		config.GetConfig().Addr.Addr = ki.Address()
		// update config file
		config.WriteConf(config.GetConfig())

		return nil
	},
}

// import a wallet with an sk
var importCmd = &cli.Command{
	Name:  "import",
	Usage: "import a wallet with an sk",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "set the keystore path",
			Value:   "./.keystore",
		},
		&cli.StringFlag{
			Name:    "secretkey",
			Aliases: []string{"sk"},
			Usage:   "private key",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"pw"},
			Usage:   "set the key password",
			Value:   "computing",
		},
	},
	Action: func(ctx *cli.Context) error {
		p := ctx.String("path")
		sk := ctx.String("sk")
		pw := ctx.String("pw")

		if p == "" {
			return fmt.Errorf("the repo path must be given")
		}
		if sk == "" {
			return fmt.Errorf("a sk must be given")
		}
		if pw == "" {
			return fmt.Errorf("the password of the wallet must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(p)
		if err != nil {
			return err
		}

		// key
		ki, err := keystore.Import(sk)
		if err != nil {
			return err
		}

		// store key into keyjson
		ks.Put(ki.Address(), pw, *ki)

		// set address
		config.GetConfig().Addr.Addr = ki.Address()
		// update config file
		config.WriteConf(config.GetConfig())

		return nil
	},
}

// select an address to use
var useCmd = &cli.Command{
	Name:  "use",
	Usage: "select an wallet address",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "set the keystore path",
			Value:   "./.keystore",
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"addr"},
			Usage:   "wallet address",
			Value:   "",
		},
	},
	Action: func(ctx *cli.Context) error {
		p := ctx.String("path")
		addr := ctx.String("address")

		if p == "" {
			return fmt.Errorf("the repo path must be given")
		}
		if addr == "" {
			return fmt.Errorf("an wallet address must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(p)
		if err != nil {
			return err
		}

		// check if name exists
		b, err := ks.Exist(addr)
		if err != nil {
			return err
		}

		if !b {
			return fmt.Errorf("this wallet is not exists: %s", addr)
		}

		// set address
		config.GetConfig().Addr.Addr = addr
		// update config file
		config.WriteConf(config.GetConfig())

		return nil
	},
}

var showCmd = &cli.Command{
	Name:  "show",
	Usage: "show the sk of an account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "set the keystore path",
			Value:   "./.keystore",
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"addr"},
			Usage:   "address",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"pw"},
			Usage:   "set the key password",
			Value:   "computing",
		},
	},
	Action: func(ctx *cli.Context) error {
		p := ctx.String("path")
		addr := ctx.String("address")
		pw := ctx.String("pw")

		if addr == "" {
			return fmt.Errorf("a sk must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(p)
		if err != nil {
			return err
		}

		// key
		ki, err := ks.Get(addr, pw)
		if err != nil {
			return err
		}

		fmt.Println("sk: ", ki.SK())

		return nil
	},
}
