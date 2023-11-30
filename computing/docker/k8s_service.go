package docker

import (
	"computing-api/computing/model"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"k8s.io/client-go/util/retry"

	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"

	networkingv1 "k8s.io/api/networking/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientSet *kubernetes.Clientset
var k8sOnce sync.Once

type K8sService struct {
	Clientset *kubernetes.Clientset
	Version   string
}

func NewK8sService() *K8sService {
	var version string
	k8sOnce.Do(func() {
		config, err := rest.InClusterConfig()
		if err != nil {
			var kubeConfig *string
			if home := homedir.HomeDir(); home != "" {
				kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
			} else {
				kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
			}
			flag.Parse()
			config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig)
			if err != nil {
				logger.Errorf("Failed create k8s config, error: %v", err)
				return
			}
		}
		clientSet, err = kubernetes.NewForConfig(config)
		if err != nil {
			logger.Errorf("Failed create k8s clientset, error: %v", err)
			return
		}

		versionInfo, err := clientSet.Discovery().ServerVersion()
		if err != nil {
			logger.Errorf("Failed get k8s version, error: %v", err)
			return
		}
		version = versionInfo.String()
	})

	return &K8sService{
		Clientset: clientSet,
		Version:   version,
	}
}

func (s *K8sService) CreateDeployment(ctx context.Context, nameSpace string, deploy *appV1.Deployment) (result *appV1.Deployment, err error) {
	return s.Clientset.AppsV1().Deployments(nameSpace).Create(ctx, deploy, metaV1.CreateOptions{})
}

func (s *K8sService) DeleteDeployment(ctx context.Context, namespace, deploymentName string) error {
	return s.Clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metaV1.DeleteOptions{})
}

func (s *K8sService) DeletePod(ctx context.Context, namespace, spaceName string) error {
	return s.Clientset.CoreV1().Pods(namespace).DeleteCollection(ctx, *metaV1.NewDeleteOptions(0), metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("lad_app=%s", spaceName),
	})
}

func (s *K8sService) DeleteDeployRs(ctx context.Context, namespace, spaceName string) error {
	return s.Clientset.AppsV1().ReplicaSets(namespace).DeleteCollection(ctx, *metaV1.NewDeleteOptions(0), metaV1.ListOptions{
		LabelSelector: fmt.Sprintf("lad_app=%s", spaceName),
	})
}

func (s *K8sService) GetDeploymentImages(ctx context.Context, namespace, deploymentName string) ([]string, error) {
	deployment, err := s.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metaV1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var imageIds []string
	for _, container := range deployment.Spec.Template.Spec.Containers {
		imageIds = append(imageIds, container.Image)
	}
	return imageIds, nil
}

func (s *K8sService) GetServiceByName(ctx context.Context, namespace, serviceName string, opts metaV1.GetOptions) (result *coreV1.Service, err error) {
	return s.Clientset.CoreV1().Services(namespace).Get(ctx, serviceName, opts)
}

func (s *K8sService) CreateService(ctx context.Context, nameSpace, spaceName string, containerPort int32) (result *coreV1.Service, err error) {
	service := &coreV1.Service{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      model.K8S_SERVICE_NAME_PREFIX + spaceName,
			Namespace: nameSpace,
		},
		Spec: coreV1.ServiceSpec{
			Ports: []coreV1.ServicePort{
				{
					Name: "http",
					Port: containerPort,
				},
			},
			Selector: map[string]string{
				"app": spaceName,
			},
		},
	}
	return s.Clientset.CoreV1().Services(nameSpace).Create(ctx, service, metaV1.CreateOptions{})
}

// create a nodeport service
func (s *K8sService) CreateNodePortService(ctx context.Context, nameSpace, appName string, containerPort int32) (result *coreV1.Service, err error) {
	service := &coreV1.Service{
		TypeMeta: metaV1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      model.K8S_SERVICE_NAME_PREFIX + appName,
			Namespace: nameSpace,
		},
		Spec: coreV1.ServiceSpec{
			Type: coreV1.ServiceTypeNodePort,
			Ports: []coreV1.ServicePort{
				{
					Name: "http",
					Port: containerPort,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/name": "load-balancer-example",
				//"app": appName,
			},
		},
	}
	// call api to create service
	res, err := s.Clientset.CoreV1().Services(nameSpace).Create(ctx, service, metaV1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *K8sService) DeleteService(ctx context.Context, namespace, serviceName string) error {
	return s.Clientset.CoreV1().Services(namespace).Delete(ctx, serviceName, metaV1.DeleteOptions{})
}

func (s *K8sService) CreateIngress(ctx context.Context, k8sNameSpace, spaceName, hostName string, port int32) (*networkingv1.Ingress, error) {
	var ingressClassName = "nginx"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metaV1.ObjectMeta{
			Name: model.K8S_INGRESS_NAME_PREFIX + spaceName,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/use-regex": "true",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []networkingv1.IngressRule{
				{
					Host: hostName,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/*",
									PathType: func() *networkingv1.PathType { t := networkingv1.PathTypePrefix; return &t }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: model.K8S_SERVICE_NAME_PREFIX + spaceName,
											Port: networkingv1.ServiceBackendPort{
												Number: port,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return s.Clientset.NetworkingV1().Ingresses(k8sNameSpace).Create(ctx, ingress, metaV1.CreateOptions{})
}

func (s *K8sService) DeleteIngress(ctx context.Context, nameSpace, ingressName string) error {
	return s.Clientset.NetworkingV1().Ingresses(nameSpace).Delete(ctx, ingressName, metaV1.DeleteOptions{})
}

func (s *K8sService) CreateConfigMap(ctx context.Context, k8sNameSpace, spaceName, basePath, configName string) (*coreV1.ConfigMap, error) {
	configFilePath := filepath.Join(basePath, configName)

	fileNameWithoutExt := filepath.Base(configName[:len(configName)-len(filepath.Ext(configName))])

	iniData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	configMap := &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name: spaceName + "-" + fileNameWithoutExt + "-" + generateString(4),
		},
		Data: map[string]string{
			configName: string(iniData),
		},
	}

	return s.Clientset.CoreV1().ConfigMaps(k8sNameSpace).Create(ctx, configMap, metaV1.CreateOptions{})
}

func (s *K8sService) GetPods(namespace, spaceName string) (bool, error) {
	listOption := metaV1.ListOptions{}
	if spaceName != "" {
		listOption = metaV1.ListOptions{
			LabelSelector: fmt.Sprintf("lad_app=%s", spaceName),
		}
	}
	podList, err := s.Clientset.CoreV1().Pods(namespace).List(context.TODO(), listOption)
	if err != nil {
		logger.Error(err)
		return false, err
	}
	if podList != nil && len(podList.Items) > 0 {
		return true, nil
	}
	return false, nil
}

func (s *K8sService) CreateNetworkPolicy(ctx context.Context, namespace string) (*networkingv1.NetworkPolicy, error) {
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      namespace + "-" + generateString(4),
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metaV1.LabelSelector{
								MatchLabels: map[string]string{
									"kubernetes.io/metadata.name": "ingress-nginx",
								},
							},
						},
					},
				},
			},
		},
	}

	return s.Clientset.NetworkingV1().NetworkPolicies(namespace).Create(ctx, networkPolicy, metaV1.CreateOptions{})
}

func (s *K8sService) CreateNameSpace(ctx context.Context, nameSpace *coreV1.Namespace, opts metaV1.CreateOptions) (result *coreV1.Namespace, err error) {
	return s.Clientset.CoreV1().Namespaces().Create(ctx, nameSpace, opts)
}

func (s *K8sService) GetNameSpace(ctx context.Context, nameSpace string, opts metaV1.GetOptions) (result *coreV1.Namespace, err error) {
	return s.Clientset.CoreV1().Namespaces().Get(ctx, nameSpace, opts)
}

func (s *K8sService) DeleteNameSpace(ctx context.Context, nameSpace string) error {
	return s.Clientset.CoreV1().Namespaces().Delete(ctx, nameSpace, metaV1.DeleteOptions{})
}

func (s *K8sService) ListUsedImage(ctx context.Context, nameSpace string) ([]string, error) {
	list, err := s.Clientset.CoreV1().Pods(nameSpace).List(ctx, metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var usedImages []string
	for _, item := range list.Items {
		for _, status := range item.Status.ContainerStatuses {
			usedImages = append(usedImages, status.Image)
		}
	}
	return usedImages, nil
}

func (s *K8sService) ListNamespace(ctx context.Context) ([]string, error) {
	list, err := s.Clientset.CoreV1().Namespaces().List(ctx, metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, item := range list.Items {
		namespaces = append(namespaces, item.Name)
	}
	return namespaces, nil
}

// func (s *K8sService) StatisticalSources(ctx context.Context) ([]*models.NodeResource, error) {
// 	activePods, err := allActivePods(s.Clientset)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var nodeList []*models.NodeResource

// 	nodes, err := s.Clientset.CoreV1().Nodes().List(ctx, metaV1.ListOptions{})
// 	if err != nil {
// 		logger.Error(err)
// 		return nil, err
// 	}

// 	nodeGpuInfoMap, err := s.GetPodLog(ctx)
// 	if err != nil {
// 		logger.Error(err)
// 		return nil, err
// 	}

// 	for _, node := range nodes.Items {
// 		nodeGpu, _, nodeResource := getNodeResource(activePods, &node)

// 		collectGpu := make(map[string]collectGpuInfo)
// 		if gpu, ok := nodeGpuInfoMap[node.Name]; ok {
// 			var gpuInfo struct {
// 				Gpu models.Gpu `json:"gpu"`
// 			}
// 			err := json.Unmarshal([]byte(gpu.String()), &gpuInfo)
// 			if err != nil {
// 				logger.Error(err)
// 				return nil, err
// 			}

// 			for index, gpuDetail := range gpuInfo.Gpu.Details {
// 				gpuName := strings.ReplaceAll(gpuDetail.ProductName, " ", "-")
// 				if v, ok := collectGpu[gpuName]; ok {
// 					v.count += 1
// 					collectGpu[gpuName] = v
// 				} else {
// 					collectGpu[gpuName] = collectGpuInfo{
// 						index,
// 						1,
// 						0,
// 					}
// 				}
// 			}

// 			for name, info := range collectGpu {
// 				runCount := int(nodeGpu[name])
// 				if num, ok := runTaskGpuResource.Load(name); ok {
// 					runCount += num.(int)
// 				}

// 				if runCount < info.count {
// 					info.remainNum = info.count - runCount
// 				} else {
// 					info.remainNum = 0
// 				}
// 				collectGpu[name] = info
// 			}

// 			var counter = make(map[string]int)
// 			newGpu := make([]models.GpuDetail, 0)
// 			for _, gpuDetail := range gpuInfo.Gpu.Details {
// 				gpuName := strings.ReplaceAll(gpuDetail.ProductName, " ", "-")
// 				newDetail := gpuDetail
// 				g := collectGpu[gpuName]
// 				if g.remainNum > 0 && counter[gpuName] < g.remainNum {
// 					newDetail.Status = models.Available
// 					counter[gpuName] += 1
// 				} else {
// 					newDetail.Status = models.Occupied
// 				}
// 				newGpu = append(newGpu, newDetail)
// 			}
// 			nodeResource.Gpu = models.Gpu{
// 				DriverVersion: gpuInfo.Gpu.DriverVersion,
// 				CudaVersion:   gpuInfo.Gpu.CudaVersion,
// 				AttachedGpus:  gpuInfo.Gpu.AttachedGpus,
// 				Details:       newGpu,
// 			}
// 		}
// 		nodeList = append(nodeList, nodeResource)
// 	}
// 	return nodeList, nil
// }

func (s *K8sService) GetPodLog(ctx context.Context) (map[string]*strings.Builder, error) {
	var num int64 = 1
	podLogOptions := coreV1.PodLogOptions{
		Container:  "",
		TailLines:  &num,
		Timestamps: false,
	}

	// get all pods in ns 'kube-system'
	podList, err := s.Clientset.CoreV1().Pods("kube-system").List(ctx, metaV1.ListOptions{
		LabelSelector: "app=resource-exporter",
	})
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	result := make(map[string]*strings.Builder)
	for _, pod := range podList.Items {
		req := s.Clientset.CoreV1().Pods("kube-system").GetLogs(pod.Name, &podLogOptions)
		buf, err := readLog(req)
		if err != nil {
			return nil, err
		}
		result[pod.Spec.NodeName] = buf
	}
	return result, nil
}

func (s *K8sService) AddNodeLabel(nodeName, key string) error {
	key = strings.ReplaceAll(key, " ", "-")

	node, err := s.Clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metaV1.GetOptions{})
	if err != nil {
		return err
	}
	node.Labels[key] = "true"
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, updateErr := s.Clientset.CoreV1().Nodes().Update(context.Background(), node, metaV1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		return fmt.Errorf("failed update node label: %w", retryErr)
	}
	return nil
}

func readLog(req *rest.Request) (*strings.Builder, error) {
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return nil, err
	}
	defer podLogs.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// generate a label from name
func GenerateLabel(name string) map[string]string {
	if name != "" {
		key := strings.ReplaceAll(name, " ", "-")
		return map[string]string{
			key: "true",
		}
	} else {
		return map[string]string{}
	}
}

func IsKubernetesVersionGreaterThan(version string, targetVersion string) bool {
	v1, err := parseKubernetesVersion(version)
	if err != nil {
		return false
	}

	v2, err := parseKubernetesVersion(targetVersion)
	if err != nil {
		return false
	}

	if v1.major > v2.major {
		return true
	} else if v1.major == v2.major && v1.minor > v2.minor {
		return true
	} else if v1.major == v2.major && v1.minor == v2.minor && v1.patch > v2.patch {
		return true
	}

	return false
}

type kubernetesVersion struct {
	major int
	minor int
	patch int
}

func parseKubernetesVersion(version string) (*kubernetesVersion, error) {
	v := &kubernetesVersion{}

	parts := strings.Split(strings.ReplaceAll(version, "v", ""), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format")
	}

	_, err := fmt.Sscanf(parts[0], "%d", &v.major)
	if err != nil {
		return nil, fmt.Errorf("failed to parse major version")
	}

	_, err = fmt.Sscanf(parts[1], "%d", &v.minor)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minor version")
	}

	_, err = fmt.Sscanf(parts[2], "%d", &v.patch)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patch version")
	}

	return v, nil
}

type collectGpuInfo struct {
	index     int
	count     int
	remainNum int
}
