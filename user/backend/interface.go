package backend

type UserBackendAPI interface {
	BackendToChainAPI
}

// Platform / User client
type BackendToChainAPI interface {
	FetchList() error
	CreateLease() error
	ShowLease() error
	StopLease() error
}

// User client (platform should not do the forwarding)
type BackendToComputingAPI interface {
	NewClient(targetURL string) error
	CloseClient() error
	Greet(msgType int32, input string, opts map[string]string) (string, error)
	Process(address string, apikey string, httpReq []byte) ([]byte, error)
}
