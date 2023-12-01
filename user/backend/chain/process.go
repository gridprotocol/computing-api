package chain

type ChainProcessor struct {
	// cache list (periodly updating)
	CurrentList []string
}

func NewChainProcessor() *ChainProcessor {
	return &ChainProcessor{
		CurrentList: nil,
	}
}

func (cp *ChainProcessor) FetchList() error {
	cp.CurrentList = []string{"localhost:12345"}
	return nil
}

func (cp *ChainProcessor) CreateLease() error {
	return nil
}

func (cp *ChainProcessor) ShowLease() error {
	return nil
}

func (cp *ChainProcessor) StopLease() error {
	return nil
}
