package main

import (
	"computing-api/common/version"
	"computing-api/computing/config"
	"computing-api/computing/gateway"
	"computing-api/computing/gateway/local"
	"computing-api/computing/gateway/remote"
	"computing-api/computing/server/httpserver"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	test *bool
)

func checkFlag() {
	test = flag.Bool("test", false, "deploy or direct forward")
}

func main() {
	// init
	checkFlag()
	if version.CheckVersion() {
		os.Exit(0)
	}
	log.Println("Current Version:", version.CurrentVersion())
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}
	grp := remote.NewGatewayRemoteProcess()
	var glp gateway.GatewayLocalProcessAPI
	if *test {
		glp = local.NewFakeImplementofLocalProcess()
	} else {
		glp = local.NewGatewayLocalProcess()
	}
	gw := gateway.NewComputingGateway(glp, grp)
	defer gw.Close()

	// server
	srv := httpserver.NewServer(config.GetConfig().Http.Listen, gw)
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("fail to start serving: %v", err)
		}
	}()

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
