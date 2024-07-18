package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gridprotocol/computing-api/computing/deploy"

	ps "github.com/mitchellh/go-ps"
)

var (
	// rpc server addr
	addr     = flag.String("addr", "localhost:12345", "remote address of the rpc server")
	contract = "0xd46e8dd67c5d32be8058bb8eb970870f07244567"
	account  = "0x683642c22feDE752415D4793832Ab75EFdF6223c"
)

func main() {

	// get k8s service
	// svc := docker.NewK8sService()

	// decode yaml
	// namespace, err := deploy.DecNamespace("./namespace.yaml")
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = svc.CreateNameSpace(context.Background(), namespace, v1.CreateOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	// parse yaml into deps and svcs
	deps, svcs, err := deploy.ParseYamlFile("./hello.yaml")
	if err != nil {
		panic(err)
		return
	}

	//------- k8s operations
	// deploy with yaml file and create a nodePort service for it
	fmt.Println("deploying and create service")
	ep, err := deploy.Deploy(deps, svcs)
	if err != nil {
		panic(err)
	}

	// show all deployments' name
	for _, dep := range deps {
		fmt.Println("deployment name: ", dep.Name)
	}

	fmt.Println("deployment ok, nodePort service is created:")
	fmt.Printf("public ips: %s\n", ep.IPs)
	fmt.Printf("node port: %d\n", ep.NodePort)

	//----------- run port-forward
	fmt.Println("checking port-forward")
	b, err := isProcessRunning("kubectl")
	if err != nil {
		panic(err)
	}
	// if port-forward is not run, run it now
	if !b {
		fmt.Println("port-forward is not running, starting..")
		go runPortForward()
	} else {
		fmt.Println("port-forward is running, restarting..")
		// kill existing port-forward process
		processes, err := ps.Processes()
		if err != nil {
			panic("get processes failed")
		}
		for _, p := range processes {
			if strings.Contains(p.Executable(), "kubectl") {
				err := syscall.Kill(p.Pid(), syscall.SIGTERM)
				if err != nil {
					fmt.Println("kill process failed: ", err)
				}
			}
		}

		// run port-forward again
		go runPortForward()
	}

	/*
		time.Sleep(3 * time.Second)

		//--------- call gateway to compute
		flag.Parse()
		fmt.Println("rpc dialing")
		// rpc connection to local service
		conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("cannot connect to the server: %v", err)
		}
		defer conn.Close()
		// get rpc client
		c := proto.NewComputeServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		// new http request
		testReq, err := http.NewRequest("GET", "http://dummy/", nil)
		if err != nil {
			log.Fatalf("fail to create a request: %v", err)
		}
		bufReq := new(bytes.Buffer)
		// write req to buf
		testReq.WriteProxy(bufReq)

		// send rpc request to call gw.compute
		fmt.Println("sending request")
		protoRESP, err := c.Process(ctx, &proto.Request{Address: account, Request: bufReq.Bytes()})
		if err != nil {
			log.Fatalf("fail to process: %v", err)
		}
		log.Printf("[Process] %v\n", len(protoRESP.GetResponse()))

		// resp bytes to buffer
		bufResp := bytes.NewBuffer(protoRESP.GetResponse())
		resp, err := http.ReadResponse(bufio.NewReader(bufResp), testReq)
		if err != nil {
			log.Fatalf("fail to read response: %v", err)
		}

		// read body from response
		body := make([]byte, 100)
		resp.Body.Read(body)
		log.Printf("[Response] %s\n", body)
	*/
}

// check if process running
func isProcessRunning(processName string) (bool, error) {
	cmd := exec.Command("pgrep", processName)
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// run port-forward command
func runPortForward() {
	cmd := exec.Command("kubectl", "port-forward", "svc/svc-hello", "8080:8080")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("run port-forward error: ", err)
		return
	} else {
		fmt.Println("exec result: ", string(out))
	}
}
