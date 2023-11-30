package deploy

import (
	"computing-api/computing/deploy/decodeYaml"
	"computing-api/computing/docker"
	"computing-api/lib/logs"
	"context"
	"fmt"
	"time"

	coreV1 "k8s.io/api/core/v1"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger = logs.Logger("deploy")

// service endpoints, any ip in IPs is available with the same port
type EndPoint struct {
	IPs      []string // public ip addresses of all nodes in service
	NodePort int32    // node port of NodePort service
}

// deploy an app and create a nodePort service for it
func Deploy(url string) (*EndPoint, error) {
	// get k8s service
	svc := docker.NewK8sService()

	fmt.Println("reading url:", url)

	data, err := decodeYaml.ReadYamlUrl(url)
	if err != nil {
		return nil, err
	}
	//------- k8s operations
	// decode yaml to deployment object
	logger.Debug("decoding yaml to obj")
	deployObject, err := decodeYaml.DecDeployment(data)
	if err != nil {
		return nil, err
	}

	logger.Debug("decode yaml ok")

	// create deployment
	// the given namespace must match the namespace in the deployment Object
	d, err := svc.CreateDeployment(context.Background(), "default", deployObject)
	if err != nil {
		return nil, err
	}
	// get deployment name
	deployName := d.GetObjectMeta().GetName()
	logger.Debugf("create deployment ok: %s", deployName)

	// create a node port service with name: svc-appName, port: port
	npSvc, err := svc.CreateNodePortService(context.TODO(), "default", "hello", int32(8080))
	if err != nil {
		return nil, err
	}

	// wait for deployment to be ready
	var retry uint
	for retry = 0; retry < 10; retry++ {
		// get current version of deployment
		deploymentsClient := svc.Clientset.AppsV1().Deployments(coreV1.NamespaceDefault)
		result, getErr := deploymentsClient.Get(context.TODO(), deployName, metaV1.GetOptions{})
		if getErr != nil {
			return nil, fmt.Errorf("failed to get latest version of deployment: %v", getErr)
		}

		// get deployment conditions
		conditions := result.Status.Conditions

		// check if deployment is ready
		isReady := false
		for _, c := range conditions {
			// check available to be True
			if c.Type == "Available" && c.Status == "True" {
				isReady = true
				logger.Debug("deployment is ready.")
				break
			}
		}
		// ready and return
		if isReady {
			// endpoint of service
			ep := &EndPoint{
				IPs:      npSvc.Spec.ExternalIPs,
				NodePort: npSvc.Spec.Ports[0].NodePort,
			}
			return ep, nil
		}

		// wait to retry
		time.Sleep(1 * time.Second)
	}

	// retry timeout
	return nil, fmt.Errorf("wait for deployment to be ready timeout")
}
