package local

import (
	"computing-api/computing/config"
	"computing-api/computing/deploy"
	"computing-api/computing/model"
	"computing-api/lib/kv"
	"computing-api/lib/logs"
	"fmt"
	"log"
)

var logger = logs.Logger("local")

type GatewayLocalProcess struct {
	db *kv.Database
}

func NewGatewayLocalProcess() *GatewayLocalProcess {
	glp := new(GatewayLocalProcess)
	db, err := kv.NewDatabase(config.GetConfig().Local.DBPath)
	if err != nil {
		logger.Error("Fail to open up the database, err: ", err)
		panic(err)
	}
	glp.db = db
	return glp
}

// TODO: api_key is required, and need some way to verify
func (glp *GatewayLocalProcess) VerifyAccessibility(address string, api_key string, needKey bool) bool {
	if len(address) == 0 && len(api_key) == 0 {
		return false
	}
	// addrBytes, err := address2bytes(address)
	// if err != nil {
	// 	logger.Error("Fail to decode address, err: ", err)
	// 	return false
	// }
	// check whether the address is authorized
	if !needKey {
		if ok, err := glp.db.Has(prefixKey(address, leasePrefix)); err != nil {
			logger.Error("Error occurs when reading db, err:", err)
			return false
		} else {
			return ok
		}
	}
	// check api_key with address
	dat, err := glp.db.Get(prefixKey(address, leasePrefix))
	if err != nil {
		logger.Error("Fail to get address's lease, err: ", err)
		return false
	}
	var l model.Lease
	if err = l.Decode(dat); err != nil {
		logger.Error("Fail to decode, err: ", err)
		return false
	}
	if checkExpire(l.Expire) {
		return false
	}
	return checkAPIkey(api_key, l.PublicKey)
}

// One approach is to record in a structure or in database
func (glp *GatewayLocalProcess) AssessPower() model.Resources {
	return model.Resources{}
}

func (glp *GatewayLocalProcess) Authorize(user string, lease model.Lease) error {
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
func (glp *GatewayLocalProcess) Deploy(user string, yamlurl string) error {
	// k8s deploy service

	// deploy with yaml and create NodePort service
	ep, err := deploy.Deploy(yamlurl)

	if err != nil {
		log.Fatalf("fail to greet: %v", err)
	}

	// use the NodePort to make an entrance
	entrance := fmt.Sprintf("localhost:%d", ep.NodePort)
	fmt.Println("entrance:", entrance)

	// record entrance
	glp.db.Put(prefixKey(user, entrancePrefix), []byte(entrance))
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
func (glp *GatewayLocalProcess) Terminate() error {
	return nil
}

func (glp *GatewayLocalProcess) Close() error {
	return glp.db.Close()
}
