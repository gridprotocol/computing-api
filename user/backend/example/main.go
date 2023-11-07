package main

import (
	"bufio"
	"bytes"
	"computing-api/computing/proto"
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr     = flag.String("addr", "localhost:12345", "remote address of the server")
	contract = "0xd46e8dd67c5d32be8058bb8eb970870f07244567"
	account  = "0x683642c22feDE752415D4793832Ab75EFdF6223c"
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
	// 1. check lease contract
	res1, err := c.Greet(ctx, &proto.GreetFromClient{Input: contract, MsgType: 0})
	if err != nil {
		log.Fatalf("fail to greet: %v", err)
	}
	log.Printf("[Greet] %v\n", res1.GetResult())
	// 2. check payee and apply for authority
	res2, err := c.Greet(ctx, &proto.GreetFromClient{Input: contract, MsgType: 1})
	if err != nil {
		log.Fatalf("fail to greet: %v", err)
	}
	log.Printf("[Greet] %v\n", res2.GetResult())
	// 3. check authority
	res3, err := c.Greet(ctx, &proto.GreetFromClient{Input: account, MsgType: 2})
	if err != nil {
		log.Fatalf("fail to greet: %v", err)
	}
	log.Printf("[Greet] %v\n", res3.GetResult())

	// process
	testReq, err := http.NewRequest("GET", "https://example/", nil)
	if err != nil {
		log.Fatalf("fail to create a request: %v", err)
	}
	buf := new(bytes.Buffer)
	testReq.WriteProxy(buf)
	resP, err := c.Process(ctx, &proto.Request{Address: account, Request: buf.Bytes()})
	if err != nil {
		log.Fatalf("fail to process: %v", err)
	}
	log.Printf("[Process] %v\n", len(resP.GetResponse()))

	bufResp := bytes.NewBuffer(resP.GetResponse())
	resp, err := http.ReadResponse(bufio.NewReader(bufResp), testReq)
	if err != nil {
		log.Fatalf("fail to read response: %v", err)
	}
	log.Printf("[Response] %v\n", resp)
}
