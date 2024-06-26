package main

import (
	"computing-api/common/version"
	"computing-api/computing/config"
	"computing-api/computing/gateway"
	"computing-api/computing/gateway/local"
	"computing-api/computing/gateway/remote"
	"computing-api/computing/proto"
	"computing-api/computing/server"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

var gw *gateway.ComputingGateway

var (
	test = false // for local test, no k8s deployment required
)

func init() {
	if version.CheckVersion() {
		os.Exit(0)
	}

	err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}

	grp := remote.NewGatewayRemoteProcess()
	var glp gateway.GatewayLocalProcessAPI
	if test {
		glp = local.NewFakeImplementofLocalProcess()
	} else {
		glp = local.NewGatewayLocalProcess()
	}
	gw = gateway.NewComputingGateway(glp, grp)
}

func main() {
	lis, err := net.Listen("tcp", config.GetConfig().Grpc.Listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	srv := server.InitEntranceService(gw)
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
