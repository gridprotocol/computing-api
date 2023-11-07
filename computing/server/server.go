package server

import (
	"computing-api/computing/gateway"
	"computing-api/computing/model"
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

// Greet for service setup
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
	case 3: // deploy
		// TODO: deploy (flexiable, enable image change in the future)
		return &proto.GreetFromServer{Result: "[ACK] already deployed"}, nil
	default:
		logger.Debug("Greet - unsupported type")
		return &proto.GreetFromServer{Result: fmt.Sprintf("[Fail] Unsupported message type: %d", gfc.MsgType)}, nil
	}
}

// Process for service usage
func (es *EntranceService) Process(ctx context.Context, gfc *proto.Request) (*proto.Response, error) {
	logger.Debug("Process")

	// check authority
	es.gw.VerifyAccessibility(gfc.GetAddress(), gfc.GetApiKey())

	// acquire entrance from recording
	entrance := "baidu.com" // temp for test
	in := model.ComputingInput{Request: gfc.Request}
	out := model.ComputingOutput{Response: nil}
	err := es.gw.Compute(entrance, &in, &out)
	if err != nil {
		logger.Error("Bad request: ", err)
		return &proto.Response{Response: nil}, err
	}

	return &proto.Response{Response: out.Response}, nil
}
