package model

type ComputingInput struct {
	Prompt  string
	Options map[string]string
}

type ComputingOutput struct {
	Result string
	Extra  map[string]string
}

// TODO: change to save in db
type DeployTask struct {
	TaskURI string
	Options map[string]string
	Expire  string
}

type Resources struct {
	Cpu     string
	Gpu     string
	Mem     string
	Storage string
}
