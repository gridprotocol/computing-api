package yaml

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

type ContainerResource struct {
	Name          string
	Count         int
	ImageName     string
	Command       []string
	Args          []string
	Env           []corev1.EnvVar
	Ports         []corev1.ContainerPort
	ResourceLimit corev1.ResourceList
	VolumeMounts  ConfigFile
	Depends       []ContainerResource
	ReadyCmd      []string
	GpuModel      string
}

type ConfigFile struct {
	Name string
	Path string
}

type Parser interface {
	Parse(yamlFile []byte) error
	GetConfig() interface{}
}

type ParserYamlV2 struct {
	config DeployYamlV2
}

func (p *ParserYamlV2) Parse(yamlFile []byte) error {
	var deploy DeployYamlV2
	if err := yaml.Unmarshal(yamlFile, &deploy); err != nil {
		return err
	}
	p.config = deploy
	return nil
}

func (p *ParserYamlV2) GetConfig() interface{} {
	return p.config
}

type Version struct {
	Version string `yaml:"version"`
}

// parse version from yaml
func getYAMLFileVersion(yamlFile []byte) (string, error) {
	var version Version
	err := yaml.Unmarshal(yamlFile, &version)
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

// parse yaml to get all container resources
func HandlerYaml(yamlFilePath string) ([]ContainerResource, error) {
	yamlFile, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed unable to read file, %w", err)
	}

	var containerResources []ContainerResource
	// version from yaml
	version, _ := getYAMLFileVersion(yamlFile)
	fmt.Println("yaml file:", yamlFilePath)
	fmt.Println("yaml version:", version)
	switch version {
	case "2.0":
		parser := &ParserYamlV2{}
		if err = parser.Parse(yamlFile); err != nil {
			return nil, fmt.Errorf("failed unable to parse YAML file, %w", err)
		}
		// get all resources for all services in yaml
		containerResources, err = parser.config.ServiceToK8sResource()
		if err != nil {
			return nil, fmt.Errorf("failed unable to parse YAML file for k8s, %w", err)
		}
	default:
		return nil, fmt.Errorf("not support yaml version: %s", version)
	}
	return containerResources, err
}