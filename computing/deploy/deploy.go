package deploy

import (
	"computing-api/computing/deploy/decyaml"
	"computing-api/computing/docker"
	"computing-api/lib/logs"
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
)

var logger = logs.Logger("deploy")

// service endpoints, any ip in IPs is available with the same port
type EndPoint struct {
	IPs      []string // public ip addresses of all nodes in service
	NodePort int32    // node port of NodePort service
}

// deploy apps and services from a local file
func DeployLocal(filepath string) (*EndPoint, error) {
	// get k8s service
	k8s := docker.NewK8sService()

	fmt.Println("reading file:", filepath)

	// read yaml file into bytes
	data, err := decyaml.ReadYamlFile(filepath)
	if err != nil {
		return nil, err
	}

	//------- k8s operations

	logger.Debug("decoding yaml to obj")
	// parse yaml data into deployments and services
	deps, svcs, err := decyaml.ParseYaml(data)
	if err != nil {
		return nil, err
	}

	logger.Debug("parse yaml ok")

	// create all deployments
	for _, dep := range deps {
		// the given namespace must match the namespace in the deployment Object
		_, err = k8s.CreateDeployment(context.Background(), "default", dep)
		if err != nil {
			return nil, err
		}
	}

	// create all services
	for _, svc := range svcs {
		k8s.Clientset.CoreV1().Services("default").Create(context.Background(), svc, metav1.CreateOptions{})
	}

	// create a node port service with name: svc-appName, port: port
	npSvc, err := CreateNodePortSvc(deps[0])
	if err != nil {
		return nil, err
	}
	// get svc name
	svcName := npSvc.GetObjectMeta().GetName()
	fmt.Printf("nodePort service is created.\nservice name: %s\nport:%d\ntargetPort:%d\nNodePort: %d\n",
		svcName,
		npSvc.Spec.Ports[0].Port,
		npSvc.Spec.Ports[0].TargetPort.IntVal,
		npSvc.Spec.Ports[0].NodePort)

	// wait for all deployments to be ready
	var allReady bool
	for _, d := range deps {
		isReady, err := WaitReady(d)
		if err != nil {
			return nil, err
		}
		if isReady {
			allReady = true
			continue
		} else {
			allReady = false
			break
		}
	}

	// if all apps is ready, return service endpoint
	if allReady {
		fmt.Println("all app is ready")
		// endpoint of service
		ep := &EndPoint{
			IPs:      npSvc.Spec.ExternalIPs,
			NodePort: npSvc.Spec.Ports[0].NodePort,
		}

		return ep, nil
	} else {
		return nil, fmt.Errorf("deployment is failed to be ready after retrys")
	}
}

// deploy apps and services from a yaml file in remote url
func Deploy(url string) (*EndPoint, error) {
	// get k8s service
	k8s := docker.NewK8sService()

	fmt.Println("reading url:", url)

	// doawnload yaml url into bytes
	data, err := decyaml.ReadYamlUrl(url)
	if err != nil {
		return nil, err
	}

	//------- k8s operations

	logger.Debug("decoding yaml to obj")
	// parse yaml data into deployments and services
	deps, svcs, err := decyaml.ParseYaml(data)
	if err != nil {
		return nil, err
	}

	logger.Debug("parse yaml ok")

	// create all deployments
	for _, dep := range deps {
		// the given namespace must match the namespace in the deployment Object
		_, err = k8s.CreateDeployment(context.Background(), "default", dep)
		if err != nil {
			return nil, err
		}
	}

	// create all services
	for _, svc := range svcs {
		k8s.Clientset.CoreV1().Services("default").Create(context.Background(), svc, metav1.CreateOptions{})
	}

	// create a node port service with name: svc-appName, port: port
	npSvc, err := CreateNodePortSvc(deps[0])
	if err != nil {
		return nil, err
	}
	// get svc name
	svcName := npSvc.GetObjectMeta().GetName()
	fmt.Printf("nodePort service is created.\nservice name: %s\nport:%d\ntargetPort:%d\nNodePort: %d\n",
		svcName,
		npSvc.Spec.Ports[0].Port,
		npSvc.Spec.Ports[0].TargetPort.IntVal,
		npSvc.Spec.Ports[0].NodePort)

	// wait for all deployments to be ready
	var allReady bool
	for _, d := range deps {
		isReady, err := WaitReady(d)
		if err != nil {
			return nil, err
		}
		if isReady {
			allReady = true
			continue
		} else {
			allReady = false
			break
		}
	}

	// if all apps is ready, return service endpoint
	if allReady {
		fmt.Println("all app is ready")
		// endpoint of service
		ep := &EndPoint{
			IPs:      npSvc.Spec.ExternalIPs,
			NodePort: npSvc.Spec.Ports[0].NodePort,
		}

		return ep, nil
	} else {
		return nil, fmt.Errorf("deployment is failed to be ready after retrys")
	}
}

// create a node port service for a deployment
func CreateNodePortSvc(d *appsv1.Deployment) (svc *corev1.Service, err error) {
	// get deployment name
	deployName := d.GetObjectMeta().GetName()

	k8s := docker.NewK8sService()
	nameSpace := "default"
	appName := deployName
	fmt.Println("app name:", deployName)
	// get containerPort from pod's container
	containerPort := d.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort
	// service's cluster port is set to containerPort here
	// it can be customized to a different port.
	port := containerPort
	npSvc, err := k8s.CreateNodePortService(context.TODO(), nameSpace, appName, port, containerPort)
	if err != nil {
		return nil, err
	}

	return npSvc, nil
}

// wait for deployment to be ready
func WaitReady(d *appsv1.Deployment) (bool, error) {
	k8s := docker.NewK8sService()
	deployName := d.GetObjectMeta().GetName()

	var retry uint
	for retry = 0; retry < 60; retry++ {
		// get current version of deployment
		deploymentsClient := k8s.Clientset.AppsV1().Deployments(corev1.NamespaceDefault)
		result, getErr := deploymentsClient.Get(context.TODO(), deployName, metav1.GetOptions{})
		if getErr != nil {
			return false, fmt.Errorf("failed to get latest version of deployment: %v", getErr)
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
			return true, nil
		}

		// wait to retry
		time.Sleep(1 * time.Second)
	}

	// retry timeout
	return false, nil

}
