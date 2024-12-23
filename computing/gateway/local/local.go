package local

import (
	"fmt"

	"github.com/gridprotocol/computing-api/computing/config"
	"github.com/gridprotocol/computing-api/computing/deploy"
	"github.com/gridprotocol/computing-api/computing/model"
	"github.com/gridprotocol/computing-api/lib/kv"
	"github.com/gridprotocol/computing-api/lib/logc"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var logger = logc.Logger("local")

const (
	testWhitelistMsg = "cheat"
)

// TODO: add cache
type GatewayLocalProcess struct {
	signExpire int64

	DB *kv.Database
}

func NewGatewayLocalProcess(db *kv.Database) *GatewayLocalProcess {
	glp := new(GatewayLocalProcess)

	glp.signExpire = int64(config.GetConfig().Local.SignExpire)
	glp.DB = db

	return glp
}

// TODO: cache
// verify auth info, signature and it's expire
func (glp *GatewayLocalProcess) CheckAuthInfo(ainfo *model.AuthInfo) bool {
	// check msg (time), if input=cheat, always ok
	if ainfo.Msg == testWhitelistMsg {
		return true
	}

	// check the expire of signature in a cookie, must within ts+signExpire
	if ok, err := checkExpire(ainfo.Msg, glp.signExpire); err != nil {
		logger.Error("Invalid time", err)
		return false
	} else {
		if !ok {
			logger.Error("Expired time", ainfo.Msg)
			return false
		}
	}

	// not nil
	if len(ainfo.Address) == 0 || len(ainfo.Sig) == 0 {
		logger.Error("Fail sig or address is nil")
		return false
	}

	// check signature
	ok, err := checkSignature(ainfo.Sig, ainfo.Address, ainfo.Msg)
	if err != nil {
		logger.Error("Bad signature, err: ", err)
		return false
	}

	return ok
}

// One approach is to record in a structure or in database
func (glp *GatewayLocalProcess) AssessPower() model.Resources {
	return model.Resources{}
}

func (glp *GatewayLocalProcess) Authorize(user string, lease model.Lease) error {
	if len(user) == 0 {
		return fmt.Errorf("user should not be empty")
	}
	if ok, err := glp.DB.Has(prefixKey(user, leasePrefix)); err != nil {
		logger.Error("Error occurs when reading db, err:", err)
		return err
	} else {
		if ok {
			return nil
		}
	}

	// set account -> lease
	lb, err := lease.Encode()
	if err != nil {
		return err
	}
	err = glp.DB.Put(prefixKey(user, leasePrefix), lb)
	if err != nil {
		return err
	}
	// add account to tasklist (db)
	// add contract address to watcherlist (db)
	return nil
}

// (flexiable, enable image change in the future, describe in the task file)
// TODO: 1. consider the edge case: already deployed, but fail to put into database
// TODO: 2. user -> lease -> resources -> yaml, which limits the resources a deployment uses
func (glp *GatewayLocalProcess) Deploy(deps []*appsv1.Deployment, svcs []*corev1.Service, user string) error {
	// k8s deploy service

	var ep *deploy.EndPoint
	var err error

	// deploy and create NodePort service
	ep, err = deploy.Deploy(deps, svcs, user)

	if err != nil {
		logger.Error("fail to deploy: ", err)
		return err
	}

	// use the service's NodePort to make an entrance
	entrance := fmt.Sprintf("http://localhost:%d", ep.NodePort)
	fmt.Println("entrance:", entrance)

	// record entrance
	err = glp.DB.Put(prefixKey(user, entrancePrefix), []byte(entrance))
	if err != nil {
		// should delete deployment or pod
		return err
	}
	return nil
}

func (glp *GatewayLocalProcess) GetEntrance(user string) (string, error) {
	ent, err := glp.DB.Get(prefixKey(user, entrancePrefix))
	if err != nil {
		return "", err
	}
	return string(ent), nil
}

// delete outdated or canceled record
// TODO: delete deployment and pod/service
func (glp *GatewayLocalProcess) Terminate(user string) error {
	keys := [][]byte{
		prefixKey(user, entrancePrefix),
		prefixKey(user, leasePrefix),
	}
	err := glp.DB.MultiDelete(keys)
	if err != nil {
		return err
	}
	return nil
}

func (glp *GatewayLocalProcess) Close() error {
	return glp.DB.Close()
}

// TODO: only forward the msg, not deal with it. Should use unified interface
// maybe input is a http request and output is a http response?
func (glp *GatewayLocalProcess) Compute(entrance string, input *model.ComputingInput, output *model.ComputingOutput) error {
	/*
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
	*/

	return nil
}
