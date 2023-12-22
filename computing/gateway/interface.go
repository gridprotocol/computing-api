package gateway

import "computing-api/computing/model"

type ComputingGatewayAPI interface {
	GatewayLocalProcessAPI
	GatewayRemoteProcessAPI

	Compute(entrance string, input *model.ComputingInput, output *model.ComputingOutput) error
}

// Local function. Processing locally without having to connect to blockchain.
type GatewayLocalProcessAPI interface {
	// verify address and api_key
	VerifyAccessibility(address string, api_key string, needKey bool) bool
	AssessPower() model.Resources
	//CalculateReward()
	Authorize(user string, lease model.Lease) error
	Deploy(user string, task string, local bool) error
	GetEntrance(user string) (string, error)
	Terminate() error
	Close() error
}

// Blockchain related. Mainly relate to smart contract.
type GatewayRemoteProcessAPI interface {
	// Register the service on a contract, which can be showned to users.
	Register(ability model.Resources) error
	// Check the settlement contract to decide whether to offer the service.
	CheckContract(contract string) bool
	CheckPayee(contract string) (bool, string)
	SetWatcher(contract string) error
	// Retrieve remuneration.
	Settle() error
}
