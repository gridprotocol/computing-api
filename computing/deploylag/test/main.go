package main

func main() {
	// // create a service with name: svc-uuid, port: port
	// deploy.CreateService("default", "demo", 80)
	// // create a ingress with name: ing-uuid, hostname, service: svc-uuid:port,
	// deploy.CreateIngress("default", "demo", "demo.localdev.me", 80)

	// deploy from yaml v2
	// r := deploy.Resource{
	// 	Cpu:     deploy.Specification{Quantity: 1, Unit: ""},
	// 	Gpu:     deploy.Specification{Quantity: 0, Unit: "m"},
	// 	Memory:  deploy.Specification{Quantity: 1, Unit: "m"},
	// 	Storage: deploy.Specification{Quantity: 1, Unit: "m"},
	// }
	// deploy.Yaml2Create(
	// 	"./pacman.yaml",
	// 	"default",
	// 	"uid",
	// 	"wallet",
	// 	"hostname",
	// 	r,
	// )

}
