package server

import (
	"computing-api/computing/gateway"
	"computing-api/computing/proto"
	"computing-api/lib/logs"
	"context"
	"fmt"
)

var logger = logs.Logger("server")

type EntranceService struct {
	proto.UnimplementedComputeServiceServer

	gw gateway.ComputingGatewayAPI
}

func InitEntranceService(gw gateway.ComputingGatewayAPI) *EntranceService {
	return &EntranceService{
		gw: gw,
	}
}

func (es *EntranceService) Greet(ctx context.Context, gfc *proto.GreetFromClient) (*proto.GreetFromServer, error) {
	switch gfc.MsgType {
	case 0: // contract
		logger.Debug("Greet - contract")
		if es.gw.CheckContract(gfc.GetInput()) {
			return &proto.GreetFromServer{Result: "[ACK] the contract is acceptable"}, nil
		} else {
			return &proto.GreetFromServer{Result: "[Fail] the contract is not acceptable"}, nil
		}
	case 1: // apply for authority
		logger.Debug("Greet - apply for authority")
		if !es.gw.CheckContract(gfc.GetInput()) {
			return &proto.GreetFromServer{Result: "[Fail] the contract is not acceptable"}, nil
		}
		// check payee (send activate tx if necessary)
		es.gw.CheckPayee(gfc.GetInput())
		// authorize and record in database and set a contract watcher
		es.gw.Authorize()
		es.gw.SetWatcher(gfc.GetInput())
		return &proto.GreetFromServer{Result: "[ACK] authorized ok"}, nil
	case 2: // check authority
		logger.Debug("Greet - check authority")
		es.gw.VerifyAccessibility(gfc.GetInput(), "")
		return &proto.GreetFromServer{Result: "[ACK] already authorized"}, nil
	default:
		logger.Debug("Greet - unsupported type")
		return &proto.GreetFromServer{Result: fmt.Sprintf("[Fail] Unsupported message type: %d", gfc.MsgType)}, nil
	}
}

func (es *EntranceService) Process(ctx context.Context, gfc *proto.Request) (*proto.Response, error) {
	logger.Debug("Process")
	es.gw.VerifyAccessibility(gfc.GetAddress(), gfc.GetApiKey())
	// TODO: deploy (flexiable, enable image change in the future)
	es.gw.Compute(nil, nil)
	return &proto.Response{Result: "[Process] ok"}, nil
}
