package service

import "computing-api/user/backend"

type CoreService struct {
	backend.BackendToChainAPI
	// ? cache pool?
	pool []backend.BackendToComputingAPI
}

func NewCoreService(chain backend.BackendToChainAPI) *CoreService {
	return &CoreService{
		BackendToChainAPI: chain,
		pool:              make([]backend.BackendToComputingAPI, 0),
	}
}

func (cs *CoreService) GetVirtualClient(addr string) {}
