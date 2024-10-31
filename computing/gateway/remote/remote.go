package remote

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/grid/contracts/eth"
	"github.com/grid/contracts/go/market"
	"github.com/grid/contracts/go/registry"
	com "github.com/gridprotocol/computing-api/common"
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/model"
	"github.com/gridprotocol/computing-api/keystore"
	"github.com/gridprotocol/computing-api/lib/kv"
	"github.com/gridprotocol/computing-api/lib/logc"
	"github.com/gridprotocol/computing-api/lib/utils"
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

func NewGatewayRemoteProcess(ep string, db *kv.Database) *GatewayRemoteProcess {
	return &GatewayRemoteProcess{
		chain_endpoint: ep,
		wallet:         config.GetConfig().Remote.Wallet,
	}
}

func (grp *GatewayRemoteProcess) Register(ability model.Resources) error {
	return nil
}

// static check for an order
func (grp *GatewayRemoteProcess) StaticCheck(orderInfo market.IMarketOrder) (bool, error) {
	// // calc value with resource and price
	// fee, err := grp.fee(user, provider)
	// if err != nil {
	// 	return false, err
	// }

	// logger.Debug("order fee: ", fee.String())

	// // check remain value
	// remain := orderInfo.Remain
	// if remain.Cmp(big.NewInt(0)) < 0 || remain.Cmp(orderInfo.TotalValue) > 0 {
	// 	return false, fmt.Errorf("remain value of this order is invalid")
	// }

	// // check remain add remueration, should equal to total value
	// v := new(big.Int).Add(orderInfo.Remain, orderInfo.Remuneration)
	// if v.Cmp(orderInfo.TotalValue) != 0 {
	// 	return false, fmt.Errorf("remuneration and remain value of this order is invalid, they should equal to the total value")
	// }

	return true, nil
}

// calc the total fee of an order
func (grp *GatewayRemoteProcess) fee(user common.Address, provider common.Address) (*big.Int, error) {
	// get node resource

	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	// get contract instance
	regIns, err := registry.NewRegistry(RegistryAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	// market
	marIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	// get order info
	order, err := marIns.GetOrder(&bind.CallOpts{}, user, provider)
	if err != nil {
		return nil, fmt.Errorf("get order failed: %s", err.Error())
	}

	logger.Debug("provider set the app name for this order")
	node, err := regIns.GetNode(&bind.CallOpts{}, common.Address(provider), order.NodeId)
	if err != nil {
		return nil, err
	}

	logger.Debug("node info:", node)

	// get order duration
	dur := order.Duration

	// calc orde fee with node resource and dur
	vcpu := node.Cpu.PriceSec
	vgpu := node.Gpu.PriceSec
	vmem := new(big.Int).Mul(new(big.Int).SetUint64(node.Mem.Num), node.Mem.PriceSec)
	vdisk := new(big.Int).Mul(new(big.Int).SetUint64(node.Disk.Num), node.Disk.PriceSec)

	v1 := new(big.Int).Add(vcpu, vgpu)
	v2 := new(big.Int).Add(vmem, vdisk)
	v3 := new(big.Int).Add(v1, v2)

	total := new(big.Int).Mul(v3, dur)

	return total, nil
}

// set the app name in contract
func (grp *GatewayRemoteProcess) SetApp(user string, app string) error {
	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	// get wallet
	cp := config.GetConfig().Remote.Wallet
	// get sk with password
	repo := keystore.Repo
	pw := com.Password
	ki, err := repo.Get(cp, pw)
	if err != nil {
		return err
	}
	sk := ki.SK()

	// make auth for sending transaction
	authProvider, err := eth.MakeAuth(chainID, sk)
	if err != nil {
		return err
	}

	// gas
	authProvider.GasLimit = 1000000
	// 50 gwei
	authProvider.GasPrice = new(big.Int).SetUint64(50000000000)

	logger.Debug("provider set the app name for this order")
	tx, err := marketIns.SetApp(authProvider, common.Address(common.HexToAddress(user)), app)
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("set app name gas used:", receipt.GasUsed)

	// get order
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, common.HexToAddress(user), common.HexToAddress(cp))
	if err != nil {
		return err
	}
	logger.Debug("app name:", orderInfo.AppName)

	return nil
}

// user renew an order
func (grp *GatewayRemoteProcess) Renew(userAddr string, userSK string, dur string) error {

	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	logger.Debug("market address:", MarketAddr)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	// make auth for sending transaction
	authUser, err := eth.MakeAuth(chainID, userSK)
	if err != nil {
		return err
	}

	// gas
	authUser.GasLimit = 1000000
	// 50 gwei
	authUser.GasPrice = new(big.Int).SetUint64(50000000000)

	_dur, err := utils.StringToUint64(dur)
	if err != nil {
		return err
	}

	logger.Debug("user renews an order")
	tx, err := marketIns.Extend(authUser, common.Address(common.HexToAddress(com.CP)), _dur)
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("renew order gas used:", receipt.GasUsed)

	return nil
}

// reset order
func (grp *GatewayRemoteProcess) Reset(user string, cp string, prob string, dur string) error {

	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	logger.Debug("market address:", MarketAddr)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	// make auth for sending transaction
	authProvider, err := eth.MakeAuth(chainID, com.SK)
	if err != nil {
		return err
	}

	// gas
	authProvider.GasLimit = 1000000
	// 50 gwei
	authProvider.GasPrice = new(big.Int).SetUint64(50000000000)

	_prob, err := utils.StringToUint64(prob)
	if err != nil {
		return err
	}
	_dur, err := utils.StringToUint64(dur)
	if err != nil {
		return err
	}

	logger.Debug("reset an order")
	tx, err := marketIns.Reset(authProvider, common.Address(common.HexToAddress(user)), common.Address(common.HexToAddress(cp)), _prob, _dur)
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("gas used:", receipt.GasUsed)

	// get order
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, common.HexToAddress(user), common.HexToAddress(cp))
	if err != nil {
		return err
	}

	logger.Debug("order info after reset:", orderInfo)

	return nil
}

// provider settle
func (grp *GatewayRemoteProcess) Settle(user string) error {

	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	logger.Debug("market address:", MarketAddr)

	// get contract instance
	marketIns, err := market.NewMarket(MarketAddr, backend)
	if err != nil {
		return fmt.Errorf("new contract instance failed: %s", err.Error())
	}

	// make auth for sending transaction
	authProvider, err := eth.MakeAuth(chainID, com.SK)
	if err != nil {
		return err
	}

	// gas
	authProvider.GasLimit = 1000000
	// 50 gwei
	authProvider.GasPrice = new(big.Int).SetUint64(50000000000)

	logger.Debug("settle an order")
	tx, err := marketIns.ProSettle(authProvider, common.Address(common.HexToAddress(user)))
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("gas used:", receipt.GasUsed)

	// get order
	orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, common.HexToAddress(user), common.HexToAddress(com.CP))
	if err != nil {
		return err
	}

	logger.Debug("order info after settle:", orderInfo)

	return nil
}

// check the order expire
func (grp *GatewayRemoteProcess) ExpireCheck(orderInfo market.IMarketOrder) (bool, error) {
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
func (grp *GatewayRemoteProcess) PayeeCheck(orderInfo market.IMarketOrder) (bool, error) {
	// if orderInfo.Provider.String() != com.CP {
	// 	return false, fmt.Errorf("the provider in order is invalid")
	// }

	return true, nil
}

func (grp *GatewayRemoteProcess) SetWatcher(contract string) error {
	return nil
}

// func (grp *GatewayRemoteProcess) Settle() error {
// 	return grp.settle(nil)
// }

// func (grp *GatewayRemoteProcess) settle(signer interface{}) error {
// 	// sign a transaction to retrieve remuneration
// 	_ = signer
// 	return nil
// }

// get an order with user and cp
func (grp *GatewayRemoteProcess) GetOrder(user string, cp string) (*market.IMarketOrder, error) {
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

// process the order check
func (grp *GatewayRemoteProcess) OrderCheck(user string, cp string) (bool, error) {

	// get order info with params
	orderInfo, err := grp.GetOrder(user, cp)
	if err != nil {
		return false, fmt.Errorf("get order failed: %s", err.Error())
	}
	logger.Debug("order info:", orderInfo)

	// TODO: cache check

	// static check
	ok, err := grp.StaticCheck(*orderInfo)
	if !ok {

		return false, fmt.Errorf("the order static check failed: %s", err.Error())
	}
	logger.Debug("static check ok")

	// check payee (send activate tx if necessary)
	ok, err = grp.PayeeCheck(*orderInfo)
	if !ok {
		//c.JSON(http.StatusBadRequest, gin.H{"msg": "[Fail] the order payee check failed: " + err.Error()})
		return false, fmt.Errorf("the order payee check failed: %s", err.Error())
	}
	logger.Debug("payee check ok")

	// check status must be activated
	if orderInfo.Status != 2 {
		var status string
		switch orderInfo.Status {
		case 0:
			status = "order not exist"
		case 1:
			status = "order unactive"
		case 3:
			status = "order cancelled"
		case 4:
			status = "order completed"
		}
		return false, fmt.Errorf("only active order can get cookie: %s", status)
	}

	// // check authorize
	// if err := hc.gw.Authorize(user, model.Lease{}); err != nil {
	// 	logger.Error(err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"msg": "[Fail] authorization failed"})
	// 	return
	// }

	// // set watcher for the lease (current ver is empty)
	// grp.SetWatcher(user)

	// // provider confirm this order after all check passed
	// logger.Debug("order check passed, proccess provider confirming")
	// err = hc.gw.ProviderConfirm(user)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"msg": fmt.Sprintf("[Fail] provider confirm failed: %s", err.Error())})
	// 	return
	// }

	return true, nil
}
