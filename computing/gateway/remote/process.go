package remote

import (
	"computing-api/computing/config"
	"computing-api/computing/model"
)

type GatewayRemoteProcess struct {
	wallet string
}

func NewGatewayRemoteProcess() *GatewayRemoteProcess {
	return &GatewayRemoteProcess{
		wallet: config.GetConfig().Remote.Wallet,
	}
}

func (grp *GatewayRemoteProcess) Register(ability model.Resources) error {
	return nil
}

func (grp *GatewayRemoteProcess) CheckContract(contract string) bool {
	return true
}

func (grp *GatewayRemoteProcess) CheckPayee(contract string) (bool, string) {
	return true, contract
}

func (grp *GatewayRemoteProcess) SetWatcher(contract string) error {
	return nil
}

func (grp *GatewayRemoteProcess) Settle() error {
	return grp.settle(nil)
}

func (grp *GatewayRemoteProcess) settle(signer interface{}) error {
	// sign a transaction to retrieve remuneration
	return nil
}
