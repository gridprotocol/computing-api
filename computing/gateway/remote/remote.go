package remote

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/grid/contracts/eth"
	"github.com/grid/contracts/go/market"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/model"
	"github.com/gridprotocol/computing-api/lib/logc"
)

var (
	logger = logc.Logger("remote")

	// market contract addr
	MarketAddr common.Address
	// access contract address
	AccessAddr common.Address
	// credit contract address
	CreditAddr common.Address
	// registry contract address
	RegistryAddr common.Address
)

type GatewayRemoteProcess struct {
	chain_endpoint string
	wallet         string
}

func NewGatewayRemoteProcess(ep string) *GatewayRemoteProcess {
	return &GatewayRemoteProcess{
		chain_endpoint: ep,
		wallet:         config.GetConfig().Remote.Wallet,
	}
}

func (grp *GatewayRemoteProcess) Register(ability model.Resources) error {
	return nil
}

// static check for an order
func (grp *GatewayRemoteProcess) StaticCheck(orderInfo market.MarketOrder) (bool, error) {
	// calc value with resource and price
	valueShould := orderValue(orderInfo)

	logger.Debug("order value should be: ", valueShould.String())
	logger.Debug("order total value: ", orderInfo.TotalValue.String())

	// check all items

	// check user confirm
	if !orderInfo.UserConfirm {
		return false, fmt.Errorf("need user confirm")
	}

	// check total value
	if valueShould.Cmp(orderInfo.TotalValue) != 0 {
		return false, fmt.Errorf("static check error: total value of the order is invalid")
	}

	// check remain value
	remain := orderInfo.Remain
	if remain.Cmp(big.NewInt(0)) < 0 || remain.Cmp(orderInfo.TotalValue) > 0 {
		return false, fmt.Errorf("remain value of this order is invalid")
	}

	// check remain add remueration, should equal to total value
	v := new(big.Int).Add(orderInfo.Remain, orderInfo.Remuneration)
	if v.Cmp(orderInfo.TotalValue) != 0 {
		return false, fmt.Errorf("remuneration and remain value of this order is invalid, they should equal to the total value")
	}

	return true, nil
}

// provider confirm an order
func (grp *GatewayRemoteProcess) ProviderConfirm(user string) error {
	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return err
	}

	// get sk from conf
	sk := config.GetConfig().Key.Key

	// make auth for sending transaction
	authProvider, err := eth.MakeAuth(chainID, sk)
	if err != nil {
		return err
	}

	// gas
	authProvider.GasLimit = 500000
	// 50 gwei
	authProvider.GasPrice = new(big.Int).SetUint64(50000000000)

	// provider call confirm
	logger.Debug("provider confirm an order")
	tx, err := marketIns.ProviderConfirm(authProvider, common.HexToAddress(user))
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("provider confirm order gas used:", receipt.GasUsed)

	// verify
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, eth.Addr1, eth.Addr2)
	if err != nil {
		return err
	}
	if !orderInfo.ProviderConfirm {
		return fmt.Errorf("provider confirm failed")
	}

	return nil
}

// provider activate an order
func (grp *GatewayRemoteProcess) Activate(user string) error {

	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	sk := config.GetConfig().Key.Key
	cp := config.GetConfig().Addr.Addr
	// make auth for sending transaction
	authProvider, err := eth.MakeAuth(chainID, sk)
	if err != nil {
		return err
	}

	// gas
	authProvider.GasLimit = 100000
	// 50 gwei
	authProvider.GasPrice = new(big.Int).SetUint64(50000000000)

	// provider call activate with user as param
	logger.Debug("provider activate an order")
	tx, err := marketIns.Activate(authProvider, common.Address(common.HexToAddress(user)))
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("activate order gas used:", receipt.GasUsed)

	// get order
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, common.HexToAddress(user), common.HexToAddress(cp))
	if err != nil {
		return err
	}
	logger.Debug("order status:", orderInfo.Status)

	// check order status
	if orderInfo.Status != 2 {
		return fmt.Errorf("activate failed, status not 2")
	}

	return nil
}

// call user cancel
func (grp *GatewayRemoteProcess) UserCancel(userAddr string, userSK string) error {

	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	logger.Debug("market address:", MarketAddr)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	cp := config.GetConfig().Addr.Addr
	// make auth for sending transaction
	authUser, err := eth.MakeAuth(chainID, userSK)
	if err != nil {
		return err
	}

	// gas
	authUser.GasLimit = 1000000
	// 50 gwei
	authUser.GasPrice = new(big.Int).SetUint64(50000000000)

	logger.Debug("user cancels an order")
	tx, err := marketIns.UserCancel(authUser, common.Address(common.HexToAddress(cp)))
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("activate order gas used:", receipt.GasUsed)

	// get order
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, common.HexToAddress(userAddr), common.HexToAddress(cp))
	if err != nil {
		return err
	}
	logger.Debug("order status:", orderInfo.Status)

	// check order status
	if orderInfo.Status != 3 {
		return fmt.Errorf("cancel failed, status not 3")
	}

	return nil
}

// check the order expire
func (grp *GatewayRemoteProcess) ExpireCheck(orderInfo market.MarketOrder) (bool, error) {
	activate := orderInfo.ActivateTime
	probation := orderInfo.Probation
	duration := orderInfo.Duration
	// check for overflow
	if activate.Int64() < 0 || activate.Int64() > int64((^uint(0)>>1)) {
		return false, fmt.Errorf("activate time overflow int64 or less than 0")
	}
	if probation.Int64() < 0 || probation.Int64() > int64((^uint(0)>>1)) {
		return false, fmt.Errorf("probation time overflow int64 or less than 0")
	}
	if duration.Int64() < 0 || duration.Int64() > int64((^uint(0)>>1)) {
		return false, fmt.Errorf("duration time overflow int64 or less than 0")
	}

	// now seconds
	now := time.Now().Unix()
	// start seconds
	start := activate.Int64()
	// end seconds
	end := start + probation.Int64() + duration.Int64()
	// expired
	if now > end {
		return false, fmt.Errorf("the order has expired, expire: %d, now: %d", end, now)
	}

	return true, nil
}

// check the order's payee to be the provider itself
func (grp *GatewayRemoteProcess) PayeeCheck(orderInfo market.MarketOrder) (bool, error) {
	// get provider addr from conf
	addr := config.GetConfig().Addr.Addr

	if orderInfo.Provider.String() != addr {
		return false, fmt.Errorf("the provider in order is invalid")
	}

	return true, nil
}

func (grp *GatewayRemoteProcess) SetWatcher(contract string) error {
	return nil
}

func (grp *GatewayRemoteProcess) Settle() error {
	return grp.settle(nil)
}

func (grp *GatewayRemoteProcess) settle(signer interface{}) error {
	// sign a transaction to retrieve remuneration
	_ = signer
	return nil
}

// get an order with user and cp
func (grp *GatewayRemoteProcess) GetOrder(user string, cp string) (*market.MarketOrder, error) {
	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	logger.Debug("market:", MarketAddr)

	// get market instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("new contract instance failed: %v, %s", err, MarketAddr)
	}

	// get order info
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, common.HexToAddress(user), common.HexToAddress(cp))
	if err != nil {
		return nil, fmt.Errorf("getorder failed: %v, %s", err, MarketAddr)
	}

	return &orderInfo, nil
}
