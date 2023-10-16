package local

import (
	"computing-api/computing/config"
	"computing-api/computing/model"
	"computing-api/lib/kv"
	"computing-api/lib/logs"
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
	}
	glp.db = db
	return glp
}

func (glp *GatewayLocalProcess) VerifyAccessibility(address string, api_key string) bool {
	return true
}

// One approach is to record in a structure or in database
func (glp *GatewayLocalProcess) AssessPower() model.Resources {
	return model.Resources{}
}

func (glp *GatewayLocalProcess) Authorize() error {
	return nil
}

func (glp *GatewayLocalProcess) Terminate() error {
	return nil
}

func (glp *GatewayLocalProcess) Close() error {
	return glp.db.Close()
}
