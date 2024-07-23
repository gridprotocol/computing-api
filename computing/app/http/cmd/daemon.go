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
	"github.com/gridprotocol/computing-api/common/version"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/gateway/local"
	"github.com/gridprotocol/computing-api/computing/gateway/remote"
	"github.com/gridprotocol/computing-api/computing/server/httpserver"
	"github.com/gridprotocol/computing-api/lib/logc"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

var logger = logc.Logger("cmd")

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
	},
	Action: func(ctx *cli.Context) error {
		test := ctx.Bool("test")
		chain := ctx.String("chain")

		// check version
		if version.CheckVersion() {
			os.Exit(0)
		}
		log.Println("Current Version:", version.CurrentVersion())

		// parse config file
		err := config.InitConfig()
		if err != nil {
			log.Fatalf("failed to init the config: %v", err)
		}

		// chain select for remote gw
		var chain_endpoint string
		switch chain {
		case "local":
			chain_endpoint = eth.Endpoint

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
			chain_endpoint = eth.Endpoint2

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
		}

		// remote gw
		grp := remote.NewGatewayRemoteProcess(chain_endpoint)
		// local gw
		var glp gateway.GatewayLocalProcessAPI

		// check for fake
		if test {
			glp = local.NewFakeImplementofLocalProcess()
		} else {
			glp = local.NewGatewayLocalProcess()
		}

		// make a gw object with local process and remote process
		gw := gateway.NewComputingGateway(glp, grp)
		defer gw.Close()

		logger.Debug("endpoint: ", config.GetConfig().Http.Listen)

		// make an httpserver with endpoint and gw object
		svr := httpserver.NewServer(config.GetConfig().Http.Listen, gw)
		// listen
		go func() {
			if err = svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("fail to start serving: %v", err)
			}
		}()

		// chan
		quit := make(chan os.Signal, 1)
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

/*
func main() {
	// init
	checkFlag()

	// check version
	if version.CheckVersion() {
		os.Exit(0)
	}
	log.Println("Current Version:", version.CurrentVersion())

	// parse config file
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}

	// chain select for remote gw
	var chain_endpoint string
	switch *chain {
	case "local":
		chain_endpoint = eth.Endpoint

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
		chain_endpoint = eth.Endpoint2

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
	}

	// remote gw
	grp := remote.NewGatewayRemoteProcess(chain_endpoint)
	// local gw
	var glp gateway.GatewayLocalProcessAPI

	// check for fake
	if *test {
		glp = local.NewFakeImplementofLocalProcess()
	} else {
		glp = local.NewGatewayLocalProcess()
	}

	// make a gw object with local process and remote process
	gw := gateway.NewComputingGateway(glp, grp)
	defer gw.Close()

	// make an httpserver with ip:port and gw object
	srv := httpserver.NewServer(config.GetConfig().Http.Listen, gw)
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("fail to start serving: %v", err)
		}
	}()

	// exit signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gateway...")
	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(cctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}
*/
