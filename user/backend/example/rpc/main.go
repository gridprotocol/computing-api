package main

import (
	"bufio"
	"bytes"
	"computing-api/common/version"
	"computing-api/computing/proto"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// rpc server address
	addr     = flag.String("addr", "localhost:12345", "remote address of the server")
	contract = "0xd46e8dd67c5d32be8058bb8eb970870f07244567"
	// account  = "0x683642c22feDE752415D4793832Ab75EFdF6223c"
	//entrance = "baidu.com"
)

func main() {
	if version.CheckVersion() {
		return
	}
	// flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot connect to the server: %v", err)
	}
	defer conn.Close()
	c := proto.NewComputeServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	once := true

	// greet
	if once {
		// 1. check lease contract
		fmt.Println("Greet 0")
		res1, err := c.Greet(ctx, &proto.GreetFromClient{Input: contract, MsgType: 0})
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res1.GetResult())

		// 2. check payee and apply for authority
		fmt.Println("Greet 1")
		res2, err := c.Greet(ctx, &proto.GreetFromClient{Input: contract, MsgType: 1})
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res2.GetResult())

		// 3. check authority
		fmt.Println("Greet 2")
		res3, err := c.Greet(ctx, &proto.GreetFromClient{Input: contract, MsgType: 2})
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res3.GetResult())

		// 4. deploy (set entrance)
		fmt.Println("Greet 3")
		yamlUrl := "https://k8s.io/examples/service/load-balancer-example.yaml"
		res4, err := c.Greet(ctx, &proto.GreetFromClient{Input: yamlUrl, Opts: map[string]string{"address": contract}, MsgType: 3})
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res4.GetResult())
	}

	// process
	testReq, err := http.NewRequest("GET", "http://example/", nil)
	if err != nil {
		log.Fatalf("fail to create a request: %v", err)
	}
	buf := new(bytes.Buffer)
	testReq.WriteProxy(buf)
	resP, err := c.Process(ctx, &proto.Request{Address: contract, Request: buf.Bytes()})
	if err != nil {
		log.Fatalf("fail to process: %v", err)
	}
	log.Printf("[Process] %v\n", len(resP.GetResponse()))

	bufResp := bytes.NewBuffer(resP.GetResponse())
	resp, err := http.ReadResponse(bufio.NewReader(bufResp), testReq)
	if err != nil {
		log.Fatalf("fail to read response: %v", err)
	}

	// read body from response
	body := make([]byte, 100)
	resp.Body.Read(body)
	log.Printf("[Response] %s\n", body)
}
