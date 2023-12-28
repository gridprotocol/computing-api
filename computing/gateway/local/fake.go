package local

import (
	"computing-api/computing/model"
	"fmt"
	"sync"
)

type FakeImplementofLocalProcess struct {
	mu     sync.RWMutex
	fakeDB map[string]string
}

func NewFakeImplementofLocalProcess() *FakeImplementofLocalProcess {
	return &FakeImplementofLocalProcess{
		fakeDB: make(map[string]string),
	}
}

func (filp *FakeImplementofLocalProcess) VerifyAccessibility(ainfo *model.AuthInfo) bool {
	key := prefixKey(ainfo.Address, leasePrefix)
	if _, ok := filp.get(string(key)); !ok {
		return false
	}
	return true
}

func (filp *FakeImplementofLocalProcess) AssessPower() model.Resources {
	return model.Resources{}
}

func (filp *FakeImplementofLocalProcess) Authorize(user string, lease model.Lease) error {
	key := prefixKey(user, leasePrefix)
	filp.put(string(key), "")
	return nil
}

func (filp *FakeImplementofLocalProcess) Deploy(user string, task string, local bool) error {
	key := prefixKey(user, entrancePrefix)
	filp.put(string(key), task)
	return nil
}

func (filp *FakeImplementofLocalProcess) GetEntrance(user string) (string, error) {
	key := prefixKey(user, entrancePrefix)
	if ent, ok := filp.get(string(key)); !ok {
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

func (filp *FakeImplementofLocalProcess) put(key, value string) {
	filp.mu.Lock()
	defer filp.mu.Unlock()
	filp.fakeDB[key] = value
}

func (filp *FakeImplementofLocalProcess) get(key string) (string, bool) {
	filp.mu.RLock()
	defer filp.mu.RUnlock()
	value, ok := filp.fakeDB[key]
	return value, ok
}
