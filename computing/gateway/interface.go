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
}

// Local function. Processing locally without having to connect to blockchain.
type GatewayLocalProcessAPI interface {
	// verify address and api_key
	CheckAuthInfo(*model.AuthInfo) bool
	AssessPower() model.Resources
	//CalculateReward()
	Authorize(user string, lease model.Lease) error
	Deploy(deps []*appsv1.Deployment, svcs []*corev1.Service, user string, nodeid uint64) error
	GetEntrance(user string) (string, error)
	// compute app after deployed
	Compute(entrance string, input *model.ComputingInput, output *model.ComputingOutput) error
	Terminate(user string) error
	Close() error
}

// Blockchain related. Mainly relate to smart contract.
type GatewayRemoteProcessAPI interface {
	// Register the service on a contract, which can be showned to users.
	Register(ability model.Resources) error

	// Check the settlement contract to decide whether to offer the service.
	StaticCheck(orderInfo market.IMarketOrder) (bool, error)
	// check expire
	ExpireCheck(orderInfo market.IMarketOrder) (bool, error)
	// provider confirm an order
	//Confirm(user string) error

	// provider activate and deactivate an order
	//Activate(user string) error
	//Deactivate(user string) error

	// provider set the app name when deploy ok
	SetApp(id uint64, app string) error
	//UserCancel(userAddr string, userSK string) error
	// user renew an order
	Extend(userSK string, id uint64, dur string) error

	// reset an order
	Reset(id uint64, prob string, dur string) error

	// provider settle an order to retrieve remueration
	Settle(id uint64) error

	// check the order's payee to be the provider itself
	PayeeCheck(orderInfo market.IMarketOrder) (bool, error)
	SetWatcher(contract string) error

	// get order with user and cp
	GetOrder(id uint64) (*market.IMarketOrder, error)

	// check order
	OrderCheck(id uint64) (bool, error)
}
