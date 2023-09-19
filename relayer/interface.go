package relayer

type RelayerPlatformAPI interface {
	RegisterService()
	RemoveService()
	ShowList()
}

// TODO: scheduler
