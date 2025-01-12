package local

import (
	"fmt"

	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/deploy"
	"github.com/gridprotocol/computing-api/computing/model"

	"github.com/gridprotocol/computing-api/lib/kv"
	"github.com/gridprotocol/computing-api/lib/logc"
	"github.com/gridprotocol/computing-api/lib/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var logger = logc.Logger("local")

const (
	testWhitelistMsg = "cheat"
)

// TODO: add cache
type GatewayLocalProcess struct {
	signExpire int64

	Platfor_Url string
	Wallet      string

	DB *kv.Database
}

func NewGatewayLocalProcess(db *kv.Database, pl_url string) *GatewayLocalProcess {
	glp := new(GatewayLocalProcess)

	glp.signExpire = int64(config.GetConfig().Local.SignExpire)
	glp.DB = db
	glp.Wallet = config.GetConfig().Remote.Wallet

	return glp
}

// TODO: cache
// verify auth info, signature and it's expire
func (glp *GatewayLocalProcess) CheckAuthInfo(ainfo *model.AuthInfo) bool {
	// check msg (time), if input=cheat, always ok
	if ainfo.Msg == testWhitelistMsg {
		return true
	}

	// check the expire of signature in a cookie, must within ts+signExpire
	if ok, err := checkExpire(ainfo.Msg, glp.signExpire); err != nil {
		logger.Error("Invalid time", err)
		return false
	} else {
		if !ok {
			logger.Error("Expired time", ainfo.Msg)
			return false
		}
	}

	// not nil
	if len(ainfo.Address) == 0 || len(ainfo.Sig) == 0 {
		logger.Error("Fail sig or address is nil")
		return false
	}

	// check signature
	ok, err := checkSignature(ainfo.Sig, ainfo.Address, ainfo.Msg)
	if err != nil {
		logger.Error("Bad signature, err: ", err)
		return false
	}

	return ok
}

// One approach is to record in a structure or in database
func (glp *GatewayLocalProcess) AssessPower() model.Resources {
	return model.Resources{}
}

func (glp *GatewayLocalProcess) Authorize(user string, lease model.Lease) error {
	if len(user) == 0 {
		return fmt.Errorf("user should not be empty")
	}
	if ok, err := glp.DB.Has(prefixKey(user, leasePrefix)); err != nil {
		logger.Error("Error occurs when reading db, err:", err)
		return err
	} else {
		if ok {
			return nil
		}
	}

	// set account -> lease
	lb, err := lease.Encode()
	if err != nil {
		return err
	}
	err = glp.DB.Put(prefixKey(user, leasePrefix), lb)
	if err != nil {
		return err
	}
	// add account to tasklist (db)
	// add contract address to watcherlist (db)
	return nil
}

// (flexiable, enable image change in the future, describe in the task file)
// TODO: 1. consider the edge case: already deployed, but fail to put into database
// TODO: 2. user -> lease -> resources -> yaml, which limits the resources a deployment uses
func (glp *GatewayLocalProcess) Deploy(deps []*appsv1.Deployment, svcs []*corev1.Service, user string, oid uint64, nodeid uint64) error {
	// k8s deploy service

	var ep *deploy.EndPoint
	var err error

	// deploy and create NodePort service
	ep, err = deploy.Deploy(deps, svcs, user, nodeid)
	if err != nil {
		logger.Error("fail to deploy: ", err)
		return err
	}

	// check svc
	if ep == nil {
		logger.Info("no svc for this deploy, skip endpoint store")
		return nil
	}

	// use the service's NodePort to make an entrance
	entrance := fmt.Sprintf("http://localhost:%d", ep.NodePort)
	fmt.Println("entrance:", entrance)

	user_oid := fmt.Sprintf("%s-%d", user, oid)
	key := prefixKey(entrancePrefix, user_oid)
	fmt.Printf("key: %s", key)
	// record entrance
	err = glp.DB.Put(key, []byte(entrance))
	if err != nil {
		// should delete deployment or pod
		return err
	}

	return nil
}

func (glp *GatewayLocalProcess) GetEntrance(user string, oid uint64) (string, error) {
	user_oid := fmt.Sprintf("%s-%d", user, oid)
	ent, err := glp.DB.Get(prefixKey(entrancePrefix, user_oid))
	if err != nil {
		return "", err
	}
	return string(ent), nil
}

// delete outdated or canceled record
// TODO: delete deployment and pod/service
func (glp *GatewayLocalProcess) Terminate(user string) error {
	keys := [][]byte{
		prefixKey(user, entrancePrefix),
		prefixKey(user, leasePrefix),
	}
	err := glp.DB.MultiDelete(keys)
	if err != nil {
		return err
	}
	return nil
}

func (glp *GatewayLocalProcess) Close() error {
	return glp.DB.Close()
}

// set the app name of an order
func (glp *GatewayLocalProcess) SetAppName(oid uint64, app string) error {
	url := fmt.Sprintf("%s/v1/order/%d/app/%s", glp.Platfor_Url, oid, app)
	fmt.Println("url:", url)
	utils.SendPost(url)
	return nil
}

// set the node avail true/false
func (glp *GatewayLocalProcess) SetNodeAvail(nodeid uint64, avail string) error {
	url := fmt.Sprintf("%s/v1/node/%s/%d/avail/%s", glp.Platfor_Url, glp.Wallet, nodeid, avail)
	fmt.Println("url:", url)
	utils.SendPost(url)
	return nil
}

// get an order with user and cp
func (glp *GatewayLocalProcess) GetOrder(id uint64) (*utils.Order, error) {
	orderInfo, err := utils.SendGetOrderRequest(id)
	if err != nil {
		return orderInfo, err
	}

	return orderInfo, nil
}

// process the order check
func (glp *GatewayLocalProcess) OrderCheck(id uint64) (bool, error) {
	// get order info with params
	orderInfo, err := glp.GetOrder(id)
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
