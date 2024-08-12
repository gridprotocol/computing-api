package cmd

import (
	"fmt"
	"log"

	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/keystore"
	"github.com/urfave/cli/v2"
)

var RepoPath string

// init config
func init() {
	// load repo path from config
	RepoPath = config.GetConfig().Remote.KeyStore
	if RepoPath == "" {
		log.Fatalf("the repo path must be given")
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
		listCmd,
		delCmd,
	},
}

// create a new wallet
var initCmd = &cli.Command{
	Name:  "init",
	Usage: "create a new wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"pw"},
			Usage:   "set the key password",
			Value:   "computing",
		},
	},
	Action: func(ctx *cli.Context) error {
		pw := ctx.String("pw")

		if pw == "" {
			return fmt.Errorf("the password of the wallet must be given")
		}

		logger.Debug("new keystore")
		// keystore
		ks, err := keystore.NewKeyStore(RepoPath)
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
		config.GetConfig().Remote.Wallet = ki.Address()
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
		sk := ctx.String("sk")
		pw := ctx.String("pw")

		if sk == "" {
			return fmt.Errorf("a sk must be given")
		}
		if pw == "" {
			return fmt.Errorf("the password of the wallet must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(RepoPath)
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
		config.GetConfig().Remote.Wallet = ki.Address()
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
			Name:    "address",
			Aliases: []string{"addr"},
			Usage:   "wallet address",
			Value:   "",
		},
	},
	Action: func(ctx *cli.Context) error {
		addr := ctx.String("address")

		if addr == "" {
			return fmt.Errorf("an wallet address must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(RepoPath)
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
		config.GetConfig().Remote.Wallet = addr
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
		addr := ctx.String("address")
		pw := ctx.String("pw")

		if addr == "" {
			return fmt.Errorf("a sk must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(RepoPath)
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

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "list all wallets",
	Flags: []cli.Flag{},
	Action: func(ctx *cli.Context) error {
		// keystore
		ks, err := keystore.NewKeyStore(RepoPath)
		if err != nil {
			return err
		}

		// key
		list, err := ks.List()
		if err != nil {
			return err
		}

		// list all wallets
		for _, v := range list {
			fmt.Println(v)
		}

		return nil
	},
}

var delCmd = &cli.Command{
	Name:  "delete",
	Usage: "delete a wallet",
	Flags: []cli.Flag{
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
		addr := ctx.String("address")
		pw := ctx.String("pw")

		if addr == "" {
			return fmt.Errorf("an wallet address must be given")
		}

		// keystore
		ks, err := keystore.NewKeyStore(RepoPath)
		if err != nil {
			return err
		}

		// delete
		err = ks.Delete(addr, pw)
		if err != nil {
			return err
		}

		return nil
	},
}
