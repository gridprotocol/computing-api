package gateway

type ComputingGatewayAPI interface {
	GatewayNetworkAPI
	GatewayLocalProcessAPI
	GatewayRemoteProcessAPI
}

type GatewayNetworkAPI interface {
	ReceiveRequest()
	ReturnResponse()
}

type GatewayLocalProcessAPI interface {
	VerifyAccessibility()
	AssessAbility()
	CalculateReward()
	Authorize()
	Terminate()
}

type GatewayRemoteProcessAPI interface {
	Register()
	Settle()
}
