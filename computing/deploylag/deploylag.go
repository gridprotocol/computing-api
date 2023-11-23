package deploylag

import (
	"context"
	"fmt"
	"path/filepath"

	"computing-api/computing/constant"
	"computing-api/computing/deploylag/yaml"
	"computing-api/computing/docker"

	"github.com/filswan/go-mcs-sdk/mcs/api/common/logs"
	appv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// get svc
var k8sService = docker.NewK8sService()

// create deployment with params
func CreateDeployment(
	n string, // name
	ns string, // namespace
	ml map[string]string, // match labels
	l map[string]string, // labels
	c []apiv1.Container, // containers
	r int32, // replicas
) error {
	// build deployment info struct
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: ns,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &r,
			Selector: &metav1.LabelSelector{
				MatchLabels: ml,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: l,
				},
				Spec: apiv1.PodSpec{
					// Containers: []apiv1.Container{
					// 	{
					// 		Name:  "web",
					// 		Image: "nginx:1.12",
					// 		Ports: []apiv1.ContainerPort{
					// 			{
					// 				Name:          "http",
					// 				Protocol:      apiv1.ProtocolTCP,
					// 				ContainerPort: 80,
					// 			},
					// 		},
					// 	},
					// },
					Containers: []apiv1.Container{
						{
							Name:  "httpd",
							Image: "httpd ",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
					//Containers: c,
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := k8sService.CreateDeployment(context.Background(), ns, deployment)
	if err != nil {
		return err
	}

	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	return nil
}

// create service
func CreateService(
	ns string, // namespace
	spaceUuid string, // space uuid
	containerPort int64, // container port
) error {
	// create service
	createService, err := k8sService.CreateService(context.TODO(), ns, spaceUuid, int32(containerPort))
	if err != nil {
		return fmt.Errorf("failed creata service, error: %w", err)
	}
	logs.GetLogger().Infof("Created service successfully: %s", createService.GetObjectMeta().GetName())

	return nil
}

// create ingress
func CreateIngress(
	ns string, // namespace
	spaceUuid string, // space uuid
	hostName string, // hostname
	containerPort int64, // container port
) error {
	// create ingress
	createIngress, err := k8sService.CreateIngress(context.TODO(), ns, spaceUuid, hostName, int32(containerPort))
	if err != nil {
		return fmt.Errorf("failed creata ingress, error: %w", err)
	}
	logs.GetLogger().Infof("Created Ingress successfully: %s", createIngress.GetObjectMeta().GetName())

	return nil
}

// parse yaml file to get all deployments and create them
func Yaml2Create(
	p string, // yaml file path
	ns string, // namespace
	uid string, // user id
	w string, // wallet address
	h string, // hostname
	r Resource, // hardware resource
) ([]corev1.Container, []corev1.Volume, error) {
	// parse containers for each deployment
	crs, err := yaml.HandlerYaml(p)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	// get memory quantity from hardware
	mem, err := resource.ParseQuantity(fmt.Sprintf("%d%s", r.Memory.Quantity, r.Memory.Unit))
	if err != nil {
		logs.GetLogger().Errorf("get memory failed, error: %+v", err)
		return nil, nil, err
	}
	// get storage quantity from hardware
	stor, err := resource.ParseQuantity(fmt.Sprintf("%d%s", r.Storage.Quantity, r.Storage.Unit))
	if err != nil {
		logs.GetLogger().Errorf("get storage failed, error: %+v", err)
		return nil, nil, err
	}

	k8sService := docker.NewK8sService()

	// make and create all deployments
	for _, cr := range crs {
		// volumeMount for container
		var volumeMount []corev1.VolumeMount
		// volumes for pods
		var volumes []corev1.Volume
		// get volumeMount and volumes from cr
		if cr.VolumeMounts.Path != "" {
			fileNameWithoutExt := filepath.Base(cr.VolumeMounts.Name[:len(cr.VolumeMounts.Name)-len(filepath.Ext(cr.VolumeMounts.Name))])
			configMap, err := k8sService.CreateConfigMap(context.TODO(), ns, uid, filepath.Dir(p), cr.VolumeMounts.Name)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, nil, err
			}
			configName := configMap.GetName()
			volumes = []corev1.Volume{
				{
					Name: uid + "-" + fileNameWithoutExt,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configName,
							},
						},
					},
				},
			}
			volumeMount = []corev1.VolumeMount{
				{
					Name:      uid + "-" + fileNameWithoutExt,
					MountPath: cr.VolumeMounts.Path,
				},
			}
		}

		var containers []corev1.Container
		// get all containers in this cr's Depends
		for _, depend := range cr.Depends {
			var handler = new(corev1.ExecAction)
			handler.Command = depend.ReadyCmd
			containers = append(containers, corev1.Container{
				Name:            uid + "-" + depend.Name,
				Image:           depend.ImageName,
				Command:         depend.Command,
				Args:            depend.Args,
				Env:             depend.Env,
				Ports:           depend.Ports,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources:       corev1.ResourceRequirements{},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: handler,
					},
					InitialDelaySeconds: 5,
					PeriodSeconds:       5,
				},
			})
		}

		// env var
		cr.Env = append(cr.Env, []corev1.EnvVar{
			{
				Name:  "wallet_address",
				Value: w,
			},
			{
				Name:  "space_uuid",
				Value: uid,
			},
			{
				Name:  "result_url",
				Value: h,
			},
			{
				Name:  "job_uuid",
				Value: uid,
			},
		}...)

		// get an additional container in cr
		containers = append(containers, corev1.Container{
			Name:            uid + "-" + cr.Name,
			Image:           cr.ImageName,
			Command:         cr.Command,
			Args:            cr.Args,
			Env:             cr.Env,
			Ports:           cr.Ports,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(r.Cpu.Quantity, resource.DecimalSI),
					corev1.ResourceMemory:           mem,
					corev1.ResourceEphemeralStorage: stor,
					"nvidia.com/gpu":                resource.MustParse(fmt.Sprintf("%d", r.Gpu.Quantity)),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(r.Cpu.Quantity, resource.DecimalSI),
					corev1.ResourceMemory:           mem,
					corev1.ResourceEphemeralStorage: stor,
					"nvidia.com/gpu":                resource.MustParse(fmt.Sprintf("%d", r.Gpu.Quantity)),
				},
			},
			VolumeMounts: volumeMount,
		})

		// create deployment with containers and volumes
		deployment := &appv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      constant.K8S_DEPLOY_NAME_PREFIX + uid,
				Namespace: ns,
			},

			Spec: appv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"lad_app": uid},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:    map[string]string{"lad_app": uid},
						Namespace: ns,
					},
					Spec: corev1.PodSpec{
						// for a pod to choose node to schedule
						//NodeSelector: docker.GenerateLabel(r.Gpu.Unit),
						Containers: containers,
						Volumes:    volumes,
					},
				},
			},
		}

		//fmt.Printf("deployment:\n%v\n", deployment)

		// call create deployment
		ret, err := k8sService.CreateDeployment(context.TODO(), ns, deployment)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}
		logs.GetLogger().Infof("Created deployment: %s", ret.GetObjectMeta().GetName())

	}

	return nil, nil, nil
}
