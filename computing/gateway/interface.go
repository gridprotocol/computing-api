package gateway

import (
	"github.com/grid/contracts/go/market"
	"github.com/gridprotocol/computing-api/computing/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

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
	Deploy(user string, deps []*appsv1.Deployment, svcs []*corev1.Service, local bool) ([]*appsv1.Deployment, error)
	GetEntrance(user string) (string, error)
	Terminate(user string) error
	Close() error
}

// Blockchain related. Mainly relate to smart contract.
type GatewayRemoteProcessAPI interface {
	// Register the service on a contract, which can be showned to users.
	Register(ability model.Resources) error

	// Check the settlement contract to decide whether to offer the service.
	StaticCheck(orderInfo market.MarketOrder) (bool, error)
	// check expire
	ExpireCheck(orderInfo market.MarketOrder) (bool, error)
	// provider confirm an order
	ProviderConfirm(user string) error
	// provider activate an order
	Activate(user string) error

	// check the order's payee to be the provider itself
	PayeeCheck(orderInfo market.MarketOrder) (bool, error)
	SetWatcher(contract string) error
	// Retrieve remuneration.
	Settle() error

	// get order with user and cp
	GetOrder(user string, cp string) (*market.MarketOrder, error)
}
