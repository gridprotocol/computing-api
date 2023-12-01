package main

import (
	"bufio"
	"bytes"
	"computing-api/user/backend/chain"
	"computing-api/user/backend/computing"
	"log"
	"net/http"
)

// test computing interface: greet and process
func quickTest() {
	comP := computing.NewComputingProcessor()
	chainP := chain.NewChainProcessor()

	// chain test
	err := chainP.FetchList()
	if err != nil {
		log.Fatal(err)
	}
	if len(chainP.CurrentList) == 0 {
		log.Fatal("no available node")
	}
	url := chainP.CurrentList[0]

	// grpc test
	err = comP.NewClient(url)
	if err != nil {
		log.Fatal(err)
	}
	defer comP.CloseClient()
	fakeAddr := "0xd46e8dd67c5d32be8058bb8eb970870f07244567"
	fakeEnt := "baidu.com"

	// greet
	{
		res1, err := comP.Greet(0, fakeAddr, nil)
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res1)

		res2, err := comP.Greet(1, fakeAddr, nil)
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res2)

		res3, err := comP.Greet(2, fakeAddr, nil)
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res3)

		res4, err := comP.Greet(3, fakeEnt, map[string]string{"address": fakeAddr})
		if err != nil {
			log.Fatalf("fail to greet: %v", err)
		}
		log.Printf("[Greet] %v\n", res4)
	}

	// process
	{
		testReq, err := http.NewRequest("GET", "https://example/", nil)
		if err != nil {
			log.Fatalf("fail to create a request: %v", err)
		}
		buf := new(bytes.Buffer)
		testReq.WriteProxy(buf)
		resP, err := comP.Process(fakeAddr, "", buf.Bytes())
		if err != nil {
			log.Fatalf("fail to process: %v", err)
		}
		log.Printf("[Process] %v\n", len(resP))

		bufResp := bytes.NewBuffer(resP)
		resp, err := http.ReadResponse(bufio.NewReader(bufResp), testReq)
		if err != nil {
			log.Fatalf("fail to read response: %v", err)
		}
		log.Printf("[Response] %v\n", resp)
	}
}
