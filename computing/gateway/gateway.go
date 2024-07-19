package gateway

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

func StopTask() {
	// delete: ingress + service + deployment + ReplicaSetsController + pods
}
