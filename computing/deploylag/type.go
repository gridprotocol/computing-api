package deploylag

type Resource struct {
	Cpu     Specification
	Memory  Specification
	Gpu     Specification
	Storage Specification
}

type Specification struct {
	Quantity int64
	Unit     string
}
