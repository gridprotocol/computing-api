package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grid/contracts/eth"
	"github.com/gridprotocol/computing-api/common/version"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/gateway/local"
	"github.com/gridprotocol/computing-api/computing/gateway/remote"
	"github.com/gridprotocol/computing-api/computing/server/httpserver"
)

var (
	test  *bool
	chain *string
)

func checkFlag() {
	test = flag.Bool("test", false, "deploy or direct forward")
	chain = flag.String("chain", "local", "select a chain to use, local or sepo")
}

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
	case "sepo":
		chain_endpoint = eth.Endpoint2
	}

	// remote gw
	grp := remote.NewGatewayRemoteProcess(chain_endpoint)
	// local gw
	var glp gateway.GatewayLocalProcessAPI
	//
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
