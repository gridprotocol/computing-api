package computing

import (
	"context"
	"fmt"
	"time"

	"github.com/gridprotocol/computing-api/computing/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ComputingProcessor struct {
	c          proto.ComputeServiceClient
	cCloseFunc func() error
	greetTO    time.Duration
	processTO  time.Duration
}

func NewComputingProcessor() *ComputingProcessor {
	return &ComputingProcessor{
		c:         nil,
		greetTO:   time.Minute,
		processTO: time.Minute,
	}
}

func (cp *ComputingProcessor) NewClient(targetURL string) error {
	conn, err := grpc.Dial(targetURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	if cp.c != nil {
		if err = cp.cCloseFunc(); err != nil {
			return err
		}
	}
	cp.c = proto.NewComputeServiceClient(conn)
	cp.cCloseFunc = conn.Close
	return nil
}

func (cp *ComputingProcessor) GetTimeout() (time.Duration, time.Duration) {
	return cp.greetTO, cp.processTO
}

func (cp *ComputingProcessor) SetTimeout(to ...time.Duration) {
	if len(to) < 2 {
		cp.greetTO = time.Minute
		cp.processTO = time.Minute
	} else {
		cp.greetTO = to[0]
		cp.processTO = to[1]
	}
}

func (cp *ComputingProcessor) CloseClient() error {
	err := cp.cCloseFunc()
	cp.c = nil
	cp.cCloseFunc = nil
	return err
}

func (cp *ComputingProcessor) Greet(msgType int32, input string, opts map[string]string) (string, error) {
	if cp.c == nil {
		return "", fmt.Errorf("no client provided")
	}

	// customized timeout
	ctx, cancel := context.WithTimeout(context.Background(), cp.greetTO)
	defer cancel()

	res, err := cp.c.Greet(ctx, &proto.GreetFromClient{Input: input, MsgType: msgType, Opts: opts})
	if err != nil {
		return "", err
	}
	return res.GetResult(), nil
}

func (cp *ComputingProcessor) Process(address string, apikey string, httpReq []byte) ([]byte, error) {
	if cp.c == nil {
		return nil, fmt.Errorf("no client provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cp.processTO)
	defer cancel()

	res, err := cp.c.Process(ctx, &proto.Request{ApiKey: apikey, Address: address, Request: httpReq})
	if err != nil {
		return nil, err
	}
	return res.GetResponse(), nil
}
