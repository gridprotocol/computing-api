package rpcserver

import (
	"context"
	"fmt"

	"github.com/gridprotocol/computing-api/computing/gateway"
	"github.com/gridprotocol/computing-api/computing/model"
	"github.com/gridprotocol/computing-api/computing/proto"
	"github.com/gridprotocol/computing-api/lib/logc"
)

var logger = logc.Logger("server")

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
		// // check contract address
		// if es.gw.StaticCheck(gfc.GetInput()) {
		// 	return &proto.GreetFromServer{Result: "[ACK] the contract is acceptable"}, nil
		// } else {
		// 	return &proto.GreetFromServer{Result: "[Fail] the contract is not acceptable"}, nil
		// }

		return nil, nil
	case 1: // apply for authority
		logger.Debug("Greet - apply for authority")
		// if !es.gw.StaticCheck(gfc.GetInput()) {
		// 	return &proto.GreetFromServer{Result: "[Fail] the contract is not acceptable"}, nil
		// }

		// check payee (send activate tx if necessary)
		_, user := es.gw.CheckPayee(gfc.GetInput())
		// authorize and record in database and set a contract watcher
		if err := es.gw.Authorize(user, model.Lease{}); err != nil {
			return &proto.GreetFromServer{Result: "[Fail] Authorize failed"}, err
		}
		es.gw.SetWatcher(gfc.GetInput())
		return &proto.GreetFromServer{Result: "[ACK] authorized ok"}, nil
	case 2: // check authority
		logger.Debug("Greet - check authority")
		ok := es.gw.VerifyAccessibility(&model.AuthInfo{Msg: gfc.GetInput()})
		if !ok {
			return &proto.GreetFromServer{Result: "[Fail] Failed to verify your account"}, nil
		}
		return &proto.GreetFromServer{Result: "[ACK] already authorized"}, nil
	case 3: // deploy
		logger.Debug("Greet - deploy")
		yamlUrl := gfc.GetInput()
		if len(yamlUrl) == 0 {
			return &proto.GreetFromServer{Result: "[Fail] empty deployment"}, nil
		}
		addr, ok := gfc.GetOpts()["address"]
		if !ok {
			return &proto.GreetFromServer{Result: "[Fail] user's address is required"}, nil
		}
		// TODO: api_key verify
		if !es.gw.VerifyAccessibility(&model.AuthInfo{Address: addr}) {
			return &proto.GreetFromServer{Result: "[Fail] user is not authorized"}, nil
		}
		// 'true' for url deploy, 'false' for local
		es.gw.Deploy(addr, yamlUrl, true)
		return &proto.GreetFromServer{Result: "[ACK] deployed ok"}, nil
	default:
		logger.Debug("Greet - unsupported type")
		return &proto.GreetFromServer{Result: fmt.Sprintf("[Fail] Unsupported message type: %d", gfc.MsgType)}, nil
	}
}

// Process for service usage
func (es *EntranceService) Process(ctx context.Context, gfc *proto.Request) (*proto.Response, error) {
	logger.Debug("Process")

	addr := gfc.GetAddress()
	// check authority
	// TODO: temp ignore this part
	if ok := es.gw.VerifyAccessibility(&model.AuthInfo{Address: addr, Msg: gfc.GetApiKey()}); !ok {
		return &proto.Response{Response: nil}, fmt.Errorf("[Fail] Failed to verify your account %s", addr)
	}

	// acquire entrance from recording
	entrance, err := es.gw.GetEntrance(addr)
	if err != nil {
		logger.Error("No Entrance: ", err)
		return &proto.Response{Response: nil}, err
	}
	in := model.ComputingInput{Request: gfc.Request}
	out := model.ComputingOutput{Response: nil}
	err = es.gw.Compute(entrance, &in, &out)
	if err != nil {
		logger.Error("Bad request: ", err)
		return &proto.Response{Response: nil}, err
	}

	return &proto.Response{Response: out.Response}, nil
}
