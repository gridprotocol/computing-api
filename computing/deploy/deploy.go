package deploy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gridprotocol/computing-api/computing/deploy/decyaml"
	"github.com/gridprotocol/computing-api/computing/docker"
	"github.com/gridprotocol/computing-api/lib/logc"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger = logc.Logger("deploy")

// service endpoints, any ip in IPs is available with the same port
type EndPoint struct {
	IPs      []string // public ip addresses of all nodes in service
	NodePort int32    // node port of NodePort service
}

// deploy apps
func Deploy(deps []*appsv1.Deployment, svcs []*corev1.Service, user string, nodeid uint64) (*EndPoint, error) {
	// get k8s service
	k8s := docker.NewK8sService()

	if len(deps) == 0 {
		return nil, fmt.Errorf("no deployment passed in")
	}

	// user address to lowercase
	lower := strings.ToLower(user)

	// append user address for each dep and svc to prevent object name conflict
	for _, dep := range deps {
		// 获取字符串长度
		length := len(lower)
		// 截取字符串的后8位
		last8Chars := lower[length-8:]

		// append lower user address to dep name
		dep.Name = fmt.Sprintf("%s-%s", dep.Name, last8Chars)

		logger.Debug("deploy, name: ", dep.Name)

		// check svc length
		if len(dep.Name) > 63 {
			return nil, fmt.Errorf("length of svc too long, must less than 63 chars: %s", dep.Name)
		}
	}

	// check deploy
	if len(deps) == 0 {
		return nil, fmt.Errorf("no deployment in yaml")
	}

	// check if svc exists
	dep0 := deps[0]
	svcName := fmt.Sprintf("svc-%s", dep0.Name)
	result, _ := k8s.GetServiceByName(context.Background(), "default", svcName, metav1.GetOptions{})
	if result.Name == svcName {
		logger.Debug("svc already exists, cancel deploy")
		return nil, fmt.Errorf("svc exists:%s, deploy cancelled", svcName)
	}

	// create all deployments
	for i, dep := range deps {
		logger.Debugf("dep num inyaml: %d", len(deps))
		logger.Debugf("create deploy index: %d", i)

		// set the nodeselector for this deployment
		ns := make(map[string]string)
		//ns["id"] = fmt.Sprintf("%d", nodeid)
		//ns["id"] = "0" // for test
		ns["kubernetes.io/hostname"] = "m21" // for test

		// set selector for this deployment
		dep.Spec.Template.Spec.NodeSelector = ns

		// logger.Debugf("create deploy, nodeSelector: id=%d", nodeid)

		// the given namespace must match the namespace in the deployment Object
		_, err := k8s.CreateDeployment(context.Background(), "default", dep)
		if err != nil {
			return nil, err
		}
	}

	// create all services
	logger.Debug("creating all svcs")
	for _, svc := range svcs {
		k8s.Clientset.CoreV1().Services("default").Create(context.Background(), svc, metav1.CreateOptions{})
	}

	var npSvc = new(corev1.Service)
	var nodePort int32
	var err error

	// if no service defined in yaml, create a nodePort svc for the deploy[0] on containerport[0]
	if len(svcs) == 0 {
		logger.Debug("no service in yaml, create svc for this deploy")
		// if port set in yaml, create svc for it
		if len(deps[0].Spec.Template.Spec.Containers) > 0 && len(deps[0].Spec.Template.Spec.Containers[0].Ports) > 0 {
			// create a node port service for the first dep with name: svc-appName, port: port
			npSvc, err = CreateNodePortSvc(deps[0])
			if err != nil {
				return nil, err
			}

			// get svc name
			svcName = npSvc.GetObjectMeta().GetName()
			fmt.Printf("nodePort service is created.\nservice name: %s\nport:%d\ntargetPort:%d\nNodePort: %d\n",
				svcName,
				npSvc.Spec.Ports[0].Port,
				npSvc.Spec.Ports[0].TargetPort.IntVal,
				npSvc.Spec.Ports[0].NodePort)

			nodePort = npSvc.Spec.Ports[0].NodePort
		}
	} else { // use existing svc's port
		logger.Debug("choose a nodeport for entrance")

		// use different nodePort for different app
		switch svcs[0].Name {
		case "provider-service":
			// choose the nodeport with 8081 targetport for provider
			for _, port := range svcs[0].Spec.Ports {
				if port.TargetPort.IntVal == 8081 {
					nodePort = port.NodePort
					break
				}
			}
		case "user-service":
			// choose the nodeport with 8080 targetport for user
			for _, port := range svcs[0].Spec.Ports {
				if port.TargetPort.IntVal == 8080 {
					nodePort = port.NodePort
					break
				}
			}
		default:
			// otherwise, use the first nodeport
			nodePort = svcs[0].Spec.Ports[0].NodePort
		}

		// // if only 1 port in svc[0], return it
		// if len(svcs[0].Spec.Ports) == 1 {
		// 	npSvc = svcs[0]
		// } else {
		// 	logger.Debug("multiple service found in yaml, use the svc with 8081 port")
		// 	// find 8081 target port for mefs-user or mefs-provider
		// 	for _, svc := range svcs {
		// 		if svc.Spec.Ports[0].TargetPort.IntVal == 8081 {
		// 			npSvc = svc
		// 			break
		// 		}
		// 	}
		// }
	}

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
		// check if svc created
		if npSvc == nil {
			return nil, nil
		} else {
			// endpoint of service
			ep := &EndPoint{
				IPs:      npSvc.Spec.ExternalIPs,
				NodePort: nodePort,
			}
			return ep, nil
		}
	} else {
		return nil, fmt.Errorf("deployment is failed to be ready after retrys")
	}
}

// create a node port service for a deployment
func CreateNodePortSvc(d *appsv1.Deployment) (svc *corev1.Service, err error) {
	// get deployment name
	deployName := d.GetObjectMeta().GetName()

	// get deploy labels
	labels := d.GetObjectMeta().GetLabels()
	// get deploy selector
	selector := labels["app.kubernetes.io/name"]
	fmt.Println("deployment selector: ", selector)
	// check deploy selector
	if selector == "" {
		return nil, fmt.Errorf("nil selector in deploy")
	}

	// k8s service
	k8s := docker.NewK8sService()
	// namespace
	nameSpace := "default"
	appName := deployName
	fmt.Println("app name:", deployName)

	// get containerPort from pod's container
	containerPort := d.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort
	// service's cluster port is set to containerPort here
	// it can be customized to a different port.
	port := containerPort

	// create a nodeport svc for a deployment
	npSvc, err := k8s.CreateNodePortService(context.TODO(), nameSpace, appName, port, containerPort, selector)
	if err != nil {
		return nil, err
	}

	return npSvc, nil
}

// wait for deployment to be ready
func WaitReady(d *appsv1.Deployment) (bool, error) {
	k8s := docker.NewK8sService()
	deployName := d.GetObjectMeta().GetName()

	logger.Debugf("wait deploy ready: %s", deployName)

	var retry uint
	for retry = 0; retry < 600; retry++ {
		// get current version of deployment
		deploymentsClient := k8s.Clientset.AppsV1().Deployments(corev1.NamespaceDefault)
		result, getErr := deploymentsClient.Get(context.TODO(), deployName, metav1.GetOptions{})
		if getErr != nil {
			logger.Debug("in wait ready, error get deployment with name:%s, err: %v ", deployName, getErr)
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

		logger.Debug("waiting ready..")
		// wait to retry
		time.Sleep(1 * time.Second)
	}

	// retry timeout
	return false, nil

}

// pase local yaml file into deps and svcs
func ParseYamlFile(filepath string) ([]*appsv1.Deployment, []*corev1.Service, error) {
	logger.Debug("reading file:", filepath)

	// read yaml file into bytes
	data, err := decyaml.ReadYamlFile(filepath)
	if err != nil {
		return nil, nil, err
	}
	logger.Debug("decoding yaml to objs")

	// parse yaml data into deployments and services
	deps, svcs, err := decyaml.ParseYaml(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parse yaml failed:%s", err.Error())
	}
	logger.Debug("parse yaml ok")

	return deps, svcs, nil
}

// pase yaml file with url into deps and svcs
func ParseYamlUrl(url string) ([]*appsv1.Deployment, []*corev1.Service, error) {
	fmt.Println("reading url:", url)

	// doawnload yaml url into bytes
	data, err := decyaml.ReadYamlUrl(url)
	if err != nil {
		return nil, nil, err
	}

	//------- k8s operations

	logger.Debug("decoding yaml to obj")
	// parse yaml data into deployments and services
	deps, svcs, err := decyaml.ParseYaml(data)
	if err != nil {
		return nil, nil, err
	}

	logger.Debug("parse yaml ok")
	return deps, svcs, nil
}

// delete a deployment and it's svc with name
func DelDeploy(depName string) error {
	// get k8s service
	k8s := docker.NewK8sService()

	// delete deployment
	logger.Info("delete dep: ", depName)
	err := k8s.DeleteDeployment(context.Background(), "default", depName)
	if err != nil {
		return err
	}

	// delete svc
	logger.Info("delete svc: ", "svc-", depName)
	err = k8s.DeleteService(context.Background(), "default", fmt.Sprintf("svc-%s", depName))
	if err != nil {
		return err
	}

	return nil
}

// clean all deploys on node
func Clean(deps []*appsv1.Deployment) error {
	// clean all deployments from k8s if error happend when deploy
	for _, dep := range deps {
		err := DelDeploy(dep.Name)
		return err
	}

	return nil
}
