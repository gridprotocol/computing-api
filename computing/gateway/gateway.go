package gateway

import "computing-api/computing/model"

// Namespace, appName, how to call and execute
type ComputingGateway struct {
	GatewayLocalProcessAPI
	GatewayRemoteProcessAPI
}

func NewComputingGateway(glp GatewayLocalProcessAPI, grp GatewayRemoteProcessAPI) *ComputingGateway {
	return &ComputingGateway{
		GatewayLocalProcessAPI:  glp,
		GatewayRemoteProcessAPI: grp,
	}
}

func (*ComputingGateway) Compute(input *model.ComputingInput, output *model.ComputingOutput) error {
	return nil
}

func StopTask() {
	// ingress + service + deployment + ReplicaSetsController + pods
}
