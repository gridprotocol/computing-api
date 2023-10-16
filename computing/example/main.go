package main

import (
	"computing-api/computing/proto"
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:12345", "remote address of the server")
)

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect to the server: %v", err)
	}
	defer conn.Close()
	c := proto.NewComputeServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res1, err := c.Greet(ctx, &proto.GreetFromClient{Input: "test1"})
	if err != nil {
		log.Fatalf("fail to greet: %v", err)
	}
	log.Printf("[Greet] %v\n", res1.GetResult())

	res2, err := c.Process(ctx, &proto.Request{Input: "test2"})
	if err != nil {
		log.Fatalf("fail to greet: %v", err)
	}
	log.Printf("[Process] %s\n", res2.GetResult())
}
