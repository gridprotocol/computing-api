package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gridprotocol/computing-api/common/version"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/gateway/local"
	"github.com/gridprotocol/computing-api/computing/gateway/remote"
	"github.com/gridprotocol/computing-api/computing/proto"
	"github.com/gridprotocol/computing-api/computing/server/rpcserver"

	"google.golang.org/grpc"
)

var gw *gateway.ComputingGateway

var (
	test = false // for local test, no k8s deployment required
)

// make a geteway object with remote and local
func init() {
	if version.CheckVersion() {
		os.Exit(0)
	}

	err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}

	// remote
	grp := remote.NewGatewayRemoteProcess()

	// local
	var glp gateway.GatewayLocalProcessAPI
	if test {
		glp = local.NewFakeImplementofLocalProcess()
	} else {
		glp = local.NewGatewayLocalProcess()
	}

	// gw
	gw = gateway.NewComputingGateway(glp, grp)
}

func main() {
	lis, err := net.Listen("tcp", config.GetConfig().Grpc.Listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	srv := rpcserver.InitEntranceService(gw)
	proto.RegisterComputeServiceServer(s, srv)

	go func() {
		if err = s.Serve(lis); err != nil {
			log.Fatalf("fail to start serving: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gateway...")
	s.GracefulStop()
	gw.Close()
}
