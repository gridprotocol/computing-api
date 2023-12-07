package gateway

import (
	"bufio"
	"bytes"
	"computing-api/computing/model"
	"net/http"
	"time"
)

// Namespace, appName, how to call and execute
type ComputingGateway struct {
	GatewayLocalProcessAPI
	GatewayRemoteProcessAPI
}

func NewComputingGateway(glp GatewayLocalProcessAPI, grp GatewayRemoteProcessAPI) *ComputingGateway {
	return &ComputingGateway{
		GatewayLocalProcessAPI:  glp,
		GatewayRemoteProcessAPI: grp,
	}
}

// TODO: only forward the msg, not deal with it. Should use unified interface
// maybe input is a http request and output is a http response?
func (cg *ComputingGateway) Compute(entrance string, input *model.ComputingInput, output *model.ComputingOutput) error {
	var req *http.Request
	var err error
	if len(input.Request) != 0 {
		// build request
		reqBuf := bytes.NewBuffer(input.Request)
		req, err = http.ReadRequest(bufio.NewReader(reqBuf))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest("GET", "http://example/", nil)
		if err != nil {
			return err
		}
	}

	// redirect entrance
	req.URL.Host = entrance
	req.Host = entrance
	req.RequestURI = ""

	// send request
	// TODO: set customized timeout
	client := &http.Client{Timeout: time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// return response
	resBuf := new(bytes.Buffer)
	res.Write(resBuf)
	output.Response = resBuf.Bytes()

	return nil
}

func StopTask() {
	// delete: ingress + service + deployment + ReplicaSetsController + pods
}
