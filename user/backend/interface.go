package backend

type UserBackendAPI interface {
	BackendInquireAPI
	BackendProcessAPI
	BackendNetworkAPI
}

type BackendInquireAPI interface {
	FetchList()
}

type BackendProcessAPI interface {
	ConstructTransaction()
	SetAuthority()
}

type BackendNetworkAPI interface {
	SendRequest()
	ReceiveResponse()
}
