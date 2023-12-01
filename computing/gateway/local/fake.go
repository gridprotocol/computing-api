package local

import (
	"computing-api/computing/model"
	"fmt"
)

type FakeImplementofLocalProcess struct {
	fakeDB map[string]string
}

func NewFakeImplementofLocalProcess() *FakeImplementofLocalProcess {
	return &FakeImplementofLocalProcess{
		fakeDB: make(map[string]string),
	}
}

func (filp *FakeImplementofLocalProcess) VerifyAccessibility(address string, api_key string, needKey bool) bool {
	key := prefixKey(address, leasePrefix)
	if _, ok := filp.fakeDB[string(key)]; !ok {
		return false
	}
	return true
}
func (filp *FakeImplementofLocalProcess) AssessPower() model.Resources {
	return model.Resources{}
}
func (filp *FakeImplementofLocalProcess) Authorize(user string, lease model.Lease) error {
	key := prefixKey(user, leasePrefix)
	filp.fakeDB[string(key)] = ""
	return nil
}
func (filp *FakeImplementofLocalProcess) Deploy(user string, task string) error {
	key := prefixKey(user, entrancePrefix)
	filp.fakeDB[string(key)] = task
	return nil
}
func (filp *FakeImplementofLocalProcess) GetEntrance(user string) (string, error) {
	key := prefixKey(user, entrancePrefix)
	if ent, ok := filp.fakeDB[string(key)]; !ok {
		return "", fmt.Errorf("entrance is not found in test map")
	} else {
		return ent, nil
	}
}
func (filp *FakeImplementofLocalProcess) Terminate() error {
	return nil
}
func (filp *FakeImplementofLocalProcess) Close() error {
	filp.fakeDB = nil
	return nil
}
