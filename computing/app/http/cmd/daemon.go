package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grid/contracts/eth"
	"github.com/grid/contracts/eth/contracts"
	com "github.com/gridprotocol/computing-api/common"
	"github.com/gridprotocol/computing-api/common/version"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/gateway/remote"
	"github.com/gridprotocol/computing-api/computing/server/httpserver"
	"github.com/gridprotocol/computing-api/keystore"
	"github.com/gridprotocol/computing-api/lib/kv"
	"github.com/gridprotocol/computing-api/lib/logc"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

var (
	logger = logc.Logger("cmd")
	// quit chan
	quit = make(chan os.Signal, 1)

	// user db records
	userDB kv.Database
)

var DaemonCmd = &cli.Command{
	Name:  "daemon",
	Usage: "platform daemon",
	Subcommands: []*cli.Command{
		runCmd,
		stopCmd,
	},
}

// run daemon
var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run server",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "deploy or direct forward",
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "chain",
			Aliases: []string{"c"},
			Usage:   "chain to interactivate, local: use local test chain, sepo: use sepo test chain",
			Value:   "local",
		},
		&cli.StringFlag{
			Name:    "password",
			Aliases: []string{"pw"},
			Usage:   "password of current wallet",
			Value:   "grid",
		},
	},
	Action: func(ctx *cli.Context) error {
		test := ctx.Bool("test")
		chain := ctx.String("chain")
		pw := ctx.String("pw")

		// get wallet and sk
		repo := keystore.Repo
		wallet := config.GetConfig().Remote.Wallet
		ki, err := repo.Get(wallet, pw)
		if err != nil {
			panic(err)
		}

		// save all info into common
		com.Password = pw
		com.CP = wallet
		com.SK = ki.SK()

		// check version
		if version.CheckVersion() {
			os.Exit(0)
		}
		log.Println("Current Version:", version.CurrentVersion())

		// chain select for remote gw
		var chain_endpoint string
		switch chain {
		case "local":
			chain_endpoint = eth.Ganache

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			l := contracts.Local{}
			l.Load()
			logger.Debugf("%+v\n", l)

			if l.Market == "" || l.Access == "" || l.Credit == "" || l.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(l.Market)
			remote.AccessAddr = common.HexToAddress(l.Access)
			remote.CreditAddr = common.HexToAddress(l.Credit)
			remote.RegistryAddr = common.HexToAddress(l.Registry)

		case "sepo":
			chain_endpoint = eth.Sepolia

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			s := contracts.Sepo{}
			s.Load()
			logger.Debugf("%+v\n", s)

			if s.Market == "" || s.Access == "" || s.Credit == "" || s.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(s.Market)
			remote.AccessAddr = common.HexToAddress(s.Access)
			remote.CreditAddr = common.HexToAddress(s.Credit)
			remote.RegistryAddr = common.HexToAddress(s.Registry)

		case "dev":
			chain_endpoint = eth.DevChain

			// load all addresses from json
			logger.Debug("load addresses")
			// loading contracts
			d := contracts.Dev{}
			d.Load()
			logger.Debugf("%+v\n", d)

			if d.Market == "" || d.Access == "" || d.Credit == "" || d.Registry == "" {
				logger.Debug("all contract addresses must exist in json file")
			}
			// save address
			remote.MarketAddr = common.HexToAddress(d.Market)
			remote.AccessAddr = common.HexToAddress(d.Access)
			remote.CreditAddr = common.HexToAddress(d.Credit)
			remote.RegistryAddr = common.HexToAddress(d.Registry)
		}

		// make a gw object
		gw := gateway.NewComputingGateway(chain_endpoint, test)
		// close db
		defer gw.Close()

		logger.Debug("listen address: ", config.GetConfig().Http.Listen)

		// make an httpserver with listen addr and gw object
		svr := httpserver.NewServer(config.GetConfig().Http.Listen, gw)
		// statr server
		go func() {
			if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("fail to start serving: %v", err)
			}
		}()

		// todo: add order expire check for all users

		// notify signal to chan
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// wait for signal and block the app
		<-quit

		// quit signal received adn end app
		log.Println("Shutting down gateway...")

		// ctx
		cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// shutdown server
		if err := svr.Shutdown(cctx); err != nil {
			log.Fatal("Server forced to shutdown: ", err)
		}

		return nil
	},
}

// stop app
var stopCmd = &cli.Command{
	Name:  "stop",
	Usage: "stop server",
	Action: func(_ *cli.Context) error {
		pidpath, err := homedir.Expand("./")
		if err != nil {
			return nil
		}
		pd, _ := os.ReadFile(path.Join(pidpath, "pid"))
		err = kill(string(pd))
		if err != nil {
			return err
		}

		quit <- syscall.SIGTERM

		log.Println("gateway gracefully exit...")

		return nil
	},
}

// kill app
func kill(pid string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("kill", "-15", pid).Run()
	case "windows":
		return exec.Command("taskkill", "/F", "/T", "/PID", pid).Run()
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
}
