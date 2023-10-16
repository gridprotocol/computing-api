package main

import (
	"computing-api/computing/config"
	"computing-api/computing/gateway"
	"computing-api/computing/gateway/local"
	"computing-api/computing/gateway/remote"
	"computing-api/computing/proto"
	"computing-api/computing/server"
	"log"
	"net"

	"google.golang.org/grpc"
)

var gw *gateway.ComputingGateway

func init() {
	grp := remote.NewGatewayRemoteProcess()
	glp := local.NewGatewayLocalProcess()
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

	if err = s.Serve(lis); err != nil {
		log.Fatalf("fail to start serving: %v", err)
	}
}
