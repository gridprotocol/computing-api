package gateway

import (
	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/gateway/local"
	"github.com/gridprotocol/computing-api/computing/gateway/remote"
	"github.com/gridprotocol/computing-api/lib/kv"
	"github.com/gridprotocol/computing-api/lib/logc"
)

// Namespace, appName, how to call and execute
type ComputingGateway struct {
	GatewayLocalProcessAPI
	GatewayRemoteProcessAPI

	DB *kv.Database
}

var logger = logc.Logger("gateway")

// func NewComputingGateway(glp GatewayLocalProcessAPI, grp GatewayRemoteProcessAPI) *ComputingGateway {
func NewComputingGateway(ep string, test bool) *ComputingGateway {
	// new kv db for gw
	db, err := kv.NewDatabase(config.GetConfig().Local.DBPath)
	if err != nil {
		logger.Error("Fail to open up the database, err: ", err)
		panic(err)
	}

	// remote gw
	grp := remote.NewGatewayRemoteProcess(ep, db)
	// local gw
	var glp GatewayLocalProcessAPI
	// check for fake
	if test {
		glp = local.NewFakeImplementofLocalProcess()
	} else {
		glp = local.NewGatewayLocalProcess(db)
	}

	return &ComputingGateway{
		GatewayLocalProcessAPI:  glp,
		GatewayRemoteProcessAPI: grp,

		DB: db,
	}
}

func StopTask() {
	// delete: ingress + service + deployment + ReplicaSetsController + pods
}

// close db for gw
func (gw *ComputingGateway) Close() error {
	return gw.DB.Close()
}
