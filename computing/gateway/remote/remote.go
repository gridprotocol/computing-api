package remote

import (
	"fmt"
	"math/big"

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
	pl_url         string
}

func NewGatewayRemoteProcess(ep string, db *kv.Database, pl_url string) *GatewayRemoteProcess {
	return &GatewayRemoteProcess{
		chain_endpoint: ep,
		wallet:         config.GetConfig().Remote.Wallet,
		pl_url:         pl_url,
	}
}

func (grp *GatewayRemoteProcess) Register(ability model.Resources) error {
	return nil
}

// set the app name in contract
func (grp *GatewayRemoteProcess) SetApp(id uint64, app string) error {
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
	tx, err := marketIns.SetApp(authProvider, id, app)
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
	// orderInfo, err := marketIns.GetOrder(&bind.CallOpts{}, id)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// set the avail status for a node
func (grp *GatewayRemoteProcess) SetAvail(nodeid uint64, avail bool) error {
	// connect to an eth node with ep
	backend, chainID := eth.ConnETH(grp.chain_endpoint)
	logger.Debug("chain id:", chainID)

	// get contract instance
	regIns, err := registry.NewRegistry(RegistryAddr, backend)
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
	tx, err := regIns.SetAvail(authProvider, common.HexToAddress(cp), nodeid, avail)
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("setavail gas used:", receipt.GasUsed)

	return nil
}

// user extend an order
func (grp *GatewayRemoteProcess) Extend(userSK string, id uint64, dur string) error {

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

	logger.Debug("user extend an order")
	tx, err := marketIns.Extend(authUser, id, _dur)
	if err != nil {
		return err
	}

	logger.Debug("waiting for tx to be ok")
	err = eth.CheckTx(grp.chain_endpoint, tx.Hash(), "")
	if err != nil {
		return err
	}

	receipt := eth.GetTransactionReceipt(grp.chain_endpoint, tx.Hash())
	logger.Debug("extend order gas used:", receipt.GasUsed)

	return nil
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

// get an order with user and cp
func (grp *GatewayRemoteProcess) GetOrder(id uint64) (*utils.Order, error) {
	orderInfo, err := utils.SendGetOrderRequest(id)
	if err != nil {
		return orderInfo, err
	}

	return orderInfo, nil
}

// process the order check
func (grp *GatewayRemoteProcess) OrderCheck(id uint64) (bool, error) {
	// get order info with params
	orderInfo, err := grp.GetOrder(id)
	if err != nil {
		return false, fmt.Errorf("get order failed: %s", err.Error())
	}

	logger.Debug("order info:", orderInfo)

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

	return true, nil
}
