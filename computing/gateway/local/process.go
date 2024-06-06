package local

import (
	"computing-api/computing/config"
	"computing-api/computing/deploy"
	"computing-api/computing/model"
	"computing-api/lib/kv"
	"computing-api/lib/logs"
	"fmt"
)

var logger = logs.Logger("local")

const (
	testWhitelistMsg = "cheat"
)

// TODO: add cache
type GatewayLocalProcess struct {
	db         *kv.Database
	signExpire int64
}

func NewGatewayLocalProcess() *GatewayLocalProcess {
	glp := new(GatewayLocalProcess)
	db, err := kv.NewDatabase(config.GetConfig().Local.DBPath)
	if err != nil {
		logger.Error("Fail to open up the database, err: ", err)
		panic(err)
	}
	glp.db = db
	glp.signExpire = int64(config.GetConfig().Local.SignExpire)
	return glp
}

// TODO: cache
func (glp *GatewayLocalProcess) VerifyAccessibility(ainfo *model.AuthInfo) bool {
	// check msg (time), if input=cheat, always ok
	if ainfo.Msg == testWhitelistMsg {
		return true
	}

	// check expire for the signature in type2
	if ok, err := checkExpire(ainfo.Msg, glp.signExpire); err != nil {
		logger.Error("Invalid time", err)
		return false
	} else {
		if !ok {
			logger.Error("Expired time", ainfo.Msg)
			return false
		}
	}

	// check sig and address
	if len(ainfo.Address) == 0 || len(ainfo.Sig) == 0 {
		logger.Error("Fail sig or address is nil")
		return false
	}
	dat, err := glp.db.Get(prefixKey(ainfo.Address, leasePrefix))
	if err != nil {
		logger.Error("Fail to get address's lease, err: ", err)
		return false
	}

	var l model.Lease
	if err = l.Decode(dat); err != nil {
		logger.Error("Fail to decode, err: ", err)
		return false
	}
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
	if ok, err := glp.db.Has(prefixKey(user, leasePrefix)); err != nil {
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
	err = glp.db.Put(prefixKey(user, leasePrefix), lb)
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
func (glp *GatewayLocalProcess) Deploy(user string, yaml string, local bool) error {
	// k8s deploy service

	var ep *deploy.EndPoint
	var err error

	// deploy with yaml and create NodePort service
	if local {
		ep, err = deploy.DeployLocal(yaml)
	} else {
		ep, err = deploy.Deploy(yaml)
	}
	if err != nil {
		logger.Error("fail to deploy: %v", err)
		return err
	}

	// use the NodePort to make an entrance
	entrance := fmt.Sprintf("http://localhost:%d", ep.NodePort)
	fmt.Println("entrance:", entrance)

	// record entrance
	err = glp.db.Put(prefixKey(user, entrancePrefix), []byte(entrance))
	if err != nil {
		// should delete deployment or pod
		return err
	}
	return nil
}

func (glp *GatewayLocalProcess) GetEntrance(user string) (string, error) {
	ent, err := glp.db.Get(prefixKey(user, entrancePrefix))
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
	err := glp.db.MultiDelete(keys)
	if err != nil {
		return err
	}
	return nil
}

func (glp *GatewayLocalProcess) Close() error {
	return glp.db.Close()
}
