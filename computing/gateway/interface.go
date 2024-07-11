package gateway

import "github.com/gridprotocol/computing-api/computing/model"

type ComputingGatewayAPI interface {
	GatewayLocalProcessAPI
	GatewayRemoteProcessAPI

	Compute(entrance string, input *model.ComputingInput, output *model.ComputingOutput) error
}

// Local function. Processing locally without having to connect to blockchain.
type GatewayLocalProcessAPI interface {
	// verify address and api_key
	VerifyAccessibility(*model.AuthInfo) bool
	AssessPower() model.Resources
	//CalculateReward()
	Authorize(user string, lease model.Lease) error
	Deploy(user string, task string, local bool) error
	GetEntrance(user string) (string, error)
	Terminate(user string) error
	Close() error
}

// Blockchain related. Mainly relate to smart contract.
type GatewayRemoteProcessAPI interface {
	// Register the service on a contract, which can be showned to users.
	Register(ability model.Resources) error

	// Check the settlement contract to decide whether to offer the service.
	StaticCheck(user string, cp string) (bool, string, error)
	// check expire
	ExpireCheck(user string, cp string) (bool, string, error)
	// provider confirm an order
	ProviderConfirm(user string) error
	// provider activate an order
	Activate(user string) error

	CheckPayee(contract string) (bool, string)
	SetWatcher(contract string) error
	// Retrieve remuneration.
	Settle() error
}
